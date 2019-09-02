package webm

import (
	"io"
)

type writerWithSizeCount struct {
	size int
	w    io.WriteCloser
}

func (w *writerWithSizeCount) Write(b []byte) (int, error) {
	w.size += len(b)
	return w.w.Write(b)
}

func (w *writerWithSizeCount) Clear() {
	w.size = 0
}

func (w *writerWithSizeCount) Close() {
	w.w.Close()
}

func (w *writerWithSizeCount) Size() int {
	return w.size
}
