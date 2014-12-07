package log

import (
	"fmt"
	"io"
	"os"
	"time"
)

// writerLogger outputs the logs to the underlying writer
type writerLogger struct {
	w io.Writer
}

func NewConsoleLogger(config *LogConfig) (Logger, error) {
	return &writerLogger{w: os.Stdout}, nil
}

func (l *writerLogger) Info(message string) {
	l.print(message)
}

func (l *writerLogger) Warning(message string) {
	l.print(message)
}

func (l *writerLogger) Error(message string) {
	l.print(message)
}

func (l *writerLogger) Fatal(message string) {
	l.print(message)
}

func (l *writerLogger) print(message string) {
	fmt.Fprintf(l.w, "%v: %v\n", time.Now().UTC().Format(time.StampMilli), message)
}
