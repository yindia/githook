package internal

import (
	// The following are blank imports for database drivers.
	// This is the common way to register database drivers in Go.
	// See https://golang.org/doc/effective_go.html#blank_import for more details.
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)
