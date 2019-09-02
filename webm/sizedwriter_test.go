package webm

import (
	"bytes"
	"testing"
)

type bufferCloser struct {
	bytes.Buffer
	closed chan struct{}
}

func (b *bufferCloser) Close() error {
	close(b.closed)
	return nil
}

func TestWriterWithSizeCount(t *testing.T) {
	buf := &bufferCloser{closed: make(chan struct{})}
	w := &writerWithSizeCount{w: buf}

	if n, err := w.Write([]byte{0x01, 0x02}); err != nil {
		t.Fatalf("Failed to Write: %v", err)
	} else if n != 2 {
		t.Errorf("Unexpected return value of writerWithSizeCount.Write, expected: 2, got: %d", n)
	}
	if n := w.Size(); n != 2 {
		t.Errorf("Unexpected return value of writerWithSizeCount.Size(), expected: 2, got: %d", n)
	}

	w.Clear()

	if n := w.Size(); n != 0 {
		t.Errorf("Unexpected return value of writerWithSizeCount.Size(), expected: 0, got: %d", n)
	}

	if err := w.Close(); err != nil {
		t.Errorf("writerWithSizeCount.Close() doesn't propagate base io.WriteCloser.Close() return value")
	}
	select {
	case <-buf.closed:
	default:
		t.Errorf("Base io.WriteCloser is not closed by writerWithSizeCount.Close()")
	}
}
