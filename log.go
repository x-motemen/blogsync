package main

import (
	"fmt"
	"os"

	"github.com/motemen/go-colorine"
)

var logger *colorine.Logger

func init() {
	logger = &colorine.Logger{
		Prefixes: colorine.Prefixes{
			"http":  colorine.Verbose,
			"store": colorine.Info,
			"error": colorine.Error,
			"":      colorine.Verbose,
		}}
	logger.SetOutput(os.Stderr)
}

func logf(prefix, pattern string, args ...interface{}) {
	logger.Log(prefix, fmt.Sprintf(pattern, args...))
}
