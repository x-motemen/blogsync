package main

import (
	"fmt"
	"os"

	"github.com/motemen/go-colorine"
)

var logger = &colorine.Logger{
	colorine.Prefixes{
		"http": colorine.Verbose,

		"download": colorine.Info,

		"error": colorine.Error,

		"": colorine.Verbose,
	},
}

func logf(prefix, pattern string, args ...interface{}) {
	logger.Log(prefix, fmt.Sprintf(pattern, args...))
}

func dieIf(err error) {
	if err != nil {
		logf("error", "%s", err)
		os.Exit(1)
	}
}
