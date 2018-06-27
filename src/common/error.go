package common

import (
	"fmt"
	"os"
)

func Error(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args)
	os.Exit(1)
}
