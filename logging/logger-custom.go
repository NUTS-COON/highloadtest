package logging

import (
	"fmt"
	"os"
	"time"
)

type logger struct {
	prefix   string
	minLevel LogLevel
	outputs  []Target
	logFile  *os.File
}

func New(minLevel LogLevel, prefix string) Logger {
	l := &logger{
		prefix:   prefix,
		minLevel: minLevel,
	}

	return l
}

func (l *logger) ToFile(path string) {
	l.SetTarget(NewFileTarget(path))
}

func (l *logger) SetTarget(t Target) {
	l.outputs = append(l.outputs, t)
}

func (l *logger) Debug(msg string) {
	l.write(msg, DEBUG)
}

func (l *logger) Info(msg string) {
	l.write(msg, INFO)
}

func (l *logger) Warn(msg string) {
	l.write(msg, WARN)
}

func (l *logger) Error(msg string) {
	l.write(msg, ERR)
}

func (l *logger) Close() error {
	return l.closeLogFile()
}

func (l *logger) closeLogFile() error {
	if l.logFile != nil {
		return l.logFile.Close()
	}

	return nil
}

func (l *logger) write(msg string, lvl LogLevel) {
	if l.minLevel <= lvl {
		for _, t := range l.outputs {
			t.Write(l.formatMsg(msg, lvl), lvl)
		}
	}
}

func (l *logger) formatMsg(msg string, lvl LogLevel) string {
	msg = fmt.Sprintf("%s %s %s", time.Now().UTC().Format(dateTimeFormat), lvl.String(), msg)
	if l.prefix != "" {
		msg = fmt.Sprintf("%s %s", l.prefix, msg)
	}

	return msg
}
