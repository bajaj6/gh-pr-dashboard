package logging

import "fmt"

type Logger struct {
	Quiet bool
}

func (l *Logger) Log(args ...interface{}) {
	if !l.Quiet {
		fmt.Println(args...)
	}
}

func (l *Logger) Logf(format string, args ...interface{}) {
	if !l.Quiet {
		fmt.Printf(format+"\n", args...)
	}
}
