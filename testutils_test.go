package ebml

import (
	"bytes"
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
