package internal

import (
	"log"
	"os"
)

func NewLogger(component string) *log.Logger {
	prefix := "githooks"
	if component != "" {
		prefix = prefix + "/" + component
	}
	return log.New(os.Stdout, prefix+" ", log.LstdFlags|log.Lmicroseconds)
}
