package activity

import (
	"encoding/json"
	"io"

	"github.com/agalitsyn/activity/internal/model"
)

type LogActivityWriter struct {
	io.WriteCloser
}

func NewLogActivityWriter(w io.WriteCloser) *LogActivityWriter {
	return &LogActivityWriter{
		WriteCloser: w,
	}
}

func (w *LogActivityWriter) WriteEntry(entry model.ActivityEntry) error {
	b, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	b = append(b, '\n')
	_, err = w.WriteCloser.Write(b)
	return err
}
