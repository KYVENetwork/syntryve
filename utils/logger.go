package utils

import (
	"github.com/rs/zerolog"
	"io"
	"os"
)

func SyntryveLogger(moduleName string) zerolog.Logger {
	writer := io.MultiWriter(os.Stdout)
	customConsoleWriter := zerolog.ConsoleWriter{Out: writer}
	customConsoleWriter.FormatCaller = func(i interface{}) string {
		return "\x1b[36m[syntryve]\x1b[0m"
	}

	logger := zerolog.New(customConsoleWriter).With().Str("module", moduleName).Timestamp().Logger()
	return logger
}
