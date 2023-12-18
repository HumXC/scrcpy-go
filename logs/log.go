package logs

import (
	"log/slog"
	"os"
)

var Default = slog.Default()
var ScrcpyOutput = os.Stdout

func GetLogger() *slog.Logger {
	return Default
}
