package logging

import (
	"log"
	"os"
)

type ConsoleTarget struct {
	out *log.Logger
}

func NewConsoleTarget() *ConsoleTarget {
	l := log.New(os.Stdout, "", 0)
	return &ConsoleTarget{
		out: l,
	}

}

func (t *ConsoleTarget) Write(msg string, lvl LogLevel) {
	t.out.Println(msg)
}

func (t *ConsoleTarget) Close() error {
	return nil
}
