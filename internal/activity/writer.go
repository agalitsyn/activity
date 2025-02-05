package activity

import (
	"encoding/json"
	"io"
	"path/filepath"
	"time"

	"github.com/natefinch/lumberjack"
)

type Entry struct {
	CreatedAt time.Time `json:"created_at"`
	Apps      []App     `json:"apps"`
}

type LogActivityWriter struct {
	io.WriteCloser
}

func NewLogActivityWriter() *LogActivityWriter {
	// TODO: to config
	logFile := &lumberjack.Logger{
		Filename:   filepath.Join("logs", "activity.log"),
		MaxSize:    100,
		MaxBackups: 7,
		MaxAge:     14,
		Compress:   true,
	}

	return &LogActivityWriter{
		WriteCloser: logFile,
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
