package activity

import (
	"encoding/json"
	"io"
	"time"
)

type Entry struct {
	CreatedAt time.Time `json:"created_at"`
	Apps      []App     `json:"apps"`
}

type LogActivityWriter struct {
	io.WriteCloser
}

func NewLogActivityWriter(w io.WriteCloser) *LogActivityWriter {
	return &LogActivityWriter{
		WriteCloser: w,
	}
}

func (w *LogActivityWriter) WriteEntry(entry Entry) error {
	b, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	b = append(b, '\n')
	_, err = w.WriteCloser.Write(b)
	return err
}
