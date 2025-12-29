package worker

import (
	"log"
	"os"
)

type Logger interface {
	Printf(format string, args ...interface{})
}

type stdLogger struct{}

func (stdLogger) Printf(format string, args ...interface{}) {
	defaultWorkerLogger.Printf(format, args...)
}

var defaultWorkerLogger = log.New(os.Stdout, "githooks/worker ", log.LstdFlags|log.Lmicroseconds)
