package logger

import (
	"log/slog"
	"os"

	"github.com/mattn/go-colorable"
	"hypera.dev/lib/slog/pretty"
)

type Logger struct {
	*slog.Logger
}

type LogOptions struct{}

func CreateLogger(options LogOptions) (*Logger, error) {
	opts := &pretty.Options{
		Level:     slog.LevelDebug,
		AddSource: true,
	}
	l := slog.New(pretty.NewHandler(colorable.NewColorable(os.Stderr), opts))
	return &Logger{l}, nil
}

func (l *Logger) Err(msg string, err error, args ...any) {
	argus := append(args, slog.Any("error", err))
	l.Error(msg, argus...)
}

func (l *Logger) FatalErr(msg string, err error, args ...any) {
	argus := append(args, slog.Any("error", err))
	l.Fatal(msg, argus)
}

func (l *Logger) Fatal(msg string, args ...any) {
	l.Error(msg, args...)
	os.Exit(1)
}
