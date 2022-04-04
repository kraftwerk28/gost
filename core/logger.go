package core

import (
	"io"
	logLib "log"
)

var Log = new(logLib.Logger)

func InitializeLogger(out io.Writer) {
	*Log = *logLib.New(out, "", logLib.Ltime|logLib.Ldate)
}

func InitBlockletLogger(log **logLib.Logger) {
}
