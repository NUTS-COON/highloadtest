package logging

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FileTarget struct {
	path  string
	out   map[LogLevel]*log.Logger
	files []*os.File
}

func NewFileTarget(path string) *FileTarget {
	return &FileTarget{
		path: path,
		out:  map[LogLevel]*log.Logger{},
	}
}

func (t *FileTarget) Write(msg string, lvl LogLevel) {
	out, ok := t.out[lvl]
	if !ok || out == nil {
		out = t.createTarget(lvl)
		t.out[lvl] = out
	}

	out.Println(msg)
}

func (t *FileTarget) Close() error {
	var err error
	for _, f := range t.files {
		err = f.Close()
	}
	return err
}

func (t *FileTarget) createTarget(lvl LogLevel) *log.Logger {
	logFile, err := os.OpenFile(t.getLogPath(lvl), os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}

	t.files = append(t.files, logFile)
	return log.New(logFile, "", 0)
}

func (t *FileTarget) getLogPath(lvl LogLevel) string {
	filename := filepath.Base(t.path)
	date := time.Now().UTC().Format("20060102")
	return strings.Replace(t.path, filename, fmt.Sprintf("%s-%s-%s", date, lvl.String(), filename), 1)
}
