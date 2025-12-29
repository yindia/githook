package internal

import (
	"log"
	"os"
)

// NewLogger creates a new logger with a standardized prefix.
// The prefix will be "githooks" or "githooks/<component>" if a component name is provided.
func NewLogger(component string) *log.Logger {
	prefix := "githooks"
	if component != "" {
		prefix = prefix + "/" + component
	}
	return log.New(os.Stdout, prefix+" ", log.LstdFlags|log.Lmicroseconds)
}
