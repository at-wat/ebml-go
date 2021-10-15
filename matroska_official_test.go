//go:build matroska_official
// +build matroska_official

package ebml

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

const (
	testDataBaseURL = "https://raw.githubusercontent.com/Matroska-Org/matroska-test-files/master/test_files/"
	cacheDir        = "ebml-go-matroska-official-test-data"
)

func loadTestData(t *testing.T, file string) ([]byte, error) {
	var r io.ReadCloser
	var hasCache bool

	cacheFile := filepath.Join(os.TempDir(), cacheDir, file)
	if f, err := os.Open(cacheFile); err == nil {
		t.Logf("Using cache: %s", cacheFile)
		r = f
		hasCache = true
	} else {
		mkvResp, err := http.Get(testDataBaseURL + file)
		if err != nil {
			return nil, err
		}
		r = mkvResp.Body
	}

	b, err := ioutil.ReadAll(r)
	if err != nil {
		r.Close()
		return nil, err
	}
	r.Close()

	if !hasCache {
		os.MkdirAll(filepath.Join(os.TempDir(), cacheDir), 0755)
		if f, err := os.Create(cacheFile); err == nil {
			n, err := f.Write(b)
			f.Close()
			if err != nil || n != len(b) {
				os.Remove(cacheFile)
			} else {
				t.Logf("Saved cache: %s", cacheFile)
			}
		}
	}

	return b, nil
}

func TestMatroskaOfficial(t *testing.T) {
	testData := map[string]struct {
		filename string
		opts     []UnmarshalOption
	}{
		"Basic": {
			filename: "test1.mkv",
		},
		"NonDefaultTimecodeScaleAndAspectRatio": {
			filename: "test2.mkv",
		},
		"HeaderStrippingAndStandardBlock": {
			filename: "test3.mkv",
		},
		"LiveStreamRecording": {
			filename: "test4.mkv",
			opts:     []UnmarshalOption{WithIgnoreUnknown(true)},
		},
		"MultipleAudioSubtitles": {
			filename: "test5.mkv",
		},
		"DifferentEBMLHeadSizesAndCueLessSeeking": {
			filename: "test6.mkv",
		},
		"ExtraUnknownJunkElementsDamaged": {
			filename: "test7.mkv",
			opts:     []UnmarshalOption{WithIgnoreUnknown(true)},
		},
		"AudioGap": {
			filename: "test8.mkv",
		},
	}
	for name, tt := range testData {
		tt := tt
		t.Run(name, func(t *testing.T) {
			mkvRaw, err := loadTestData(t, tt.filename)
			if err != nil {
				t.Fatal(err)
			}

			var dump string
			for i := 0; i < 16 && i < len(mkvRaw); i++ {
				dump += fmt.Sprintf("%02x ", mkvRaw[i])
			}
			t.Logf("dump: %s", dump)

			var mkv map[string]interface{}
			if err := Unmarshal(bytes.NewReader(mkvRaw), &mkv, tt.opts...); err != nil {
				t.Fatalf("Failed to unmarshal: '%v'", err)
			}
			txt := fmt.Sprintf("%+v", mkv)
			if len(txt) > 512 {
				t.Logf("result: %s...", txt[:512])
			} else {
				t.Logf("result: %s", txt)
			}
		})
	}
}
