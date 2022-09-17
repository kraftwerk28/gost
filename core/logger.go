package core

import (
	"io"
	logLib "log"
)

type Logger struct {}

func NewLogger() *Logger {
	return nil
}

var Log = new(logLib.Logger)

func InitializeLogger(out io.Writer) {
	*Log = *logLib.New(out, "", logLib.Ltime|logLib.Ldate)
}

func InitBlockletLogger(log **logLib.Logger) {
}

func LogFromBlocklet(b I3barBlocklet) {
}
