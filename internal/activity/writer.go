package activity

import (
	"io"
)

type LogActivityWriter struct {
	io.WriteCloser
}

func NewLogActivityWriter(w io.WriteCloser) *LogActivityWriter {
	return &LogActivityWriter{
		WriteCloser: w,
	}
}

func (w *LogActivityWriter) Write(b []byte) error {
	b = append(b, '\n')
	_, err := w.WriteCloser.Write(b)
	return err
}
