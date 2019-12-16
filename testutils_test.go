package ebml

import (
	"bytes"
	"io"
)

type limitedDummyWriter struct {
	n     int
	limit int
}

func (s *limitedDummyWriter) Write(b []byte) (int, error) {
	s.n += len(b)
	if s.n > s.limit {
		return len(b) - (s.n - s.limit), bytes.ErrTooLarge
	}
	return len(b), nil
}

type delayedBrokenReader struct {
	b     []byte
	n     int
	limit int
}

func (s *delayedBrokenReader) Read(b []byte) (int, error) {
	p := s.n
	s.n += len(b)
	if s.n > s.limit {
		return len(b) - (s.n - s.limit), io.ErrClosedPipe
	}
	copy(b, s.b[p:p+len(b)])
	return len(b), nil
}

func isErr(err error, target interface{}) bool {
	switch v := target.(type) {
	case error:
		return err == v
	case func(err error) bool:
		return v(err)
	case nil:
		return err == nil
	}
	panic("invalid isErr target")
}
