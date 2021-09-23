package logging

import (
	"strings"
)

const (
	dateTimeFormat = "2006-01-02T15:04:05.000"
)

type LogLevel int

const (
	DEBUG = iota
	INFO
	WARN
	ERR
	FATAL
)

type Logger interface {
	SetTarget(t Target)
	ToFile(path string)
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
	Close() error
}

type Target interface {
	Write(msg string, lvl LogLevel)
	Close() error
}

func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	}

	return ""
}

func ParseLogLevel(s string) LogLevel {
	s = strings.ToLower(s)
	switch s {
	case "debug":
		return DEBUG
	case "info":
		return INFO
	case "warn":
		return WARN
	case "error":
		return ERR
	}

	return INFO
}
