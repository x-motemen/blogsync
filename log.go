package main

import (
	"fmt"

	"github.com/motemen/go-colorine"
)

var logger = &colorine.Logger{
	colorine.Prefixes{
		"http":  colorine.Verbose,
		"store": colorine.Info,
		"error": colorine.Error,
		"":      colorine.Verbose,
	},
}

func logf(prefix, pattern string, args ...interface{}) {
	logger.Log(prefix, fmt.Sprintf(pattern, args...))
}
