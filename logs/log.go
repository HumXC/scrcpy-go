package logs

import (
	"log/slog"
	"os"
)

var Default = slog.Default()
var ScrcpyOutput = os.Stderr

func GetLogger() *slog.Logger {
	return Default
}
