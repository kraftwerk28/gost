package core

import (
	logLib "log"
	"os"
)

var Log = new(logLib.Logger)

func InitializeLogger(filename string) {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	*Log = *logLib.New(f, "", logLib.Ltime|logLib.Ldate)
}
