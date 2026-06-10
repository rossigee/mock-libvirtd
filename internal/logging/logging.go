package logging

import (
	"log/slog"
	"os"
	"sync/atomic"
)

var level atomic.Int32

func Init(levelStr string) {
	var lvl slog.Level
	switch levelStr {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}
	level.Store(int32(lvl))

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: lvl,
	})))
}

func SetLevel(lvl string) {
	var l slog.Level
	switch lvl {
	case "debug":
		l = slog.LevelDebug
	case "warn":
		l = slog.LevelWarn
	case "error":
		l = slog.LevelError
	case "info":
		l = slog.LevelInfo
	default:
		return
	}
	level.Store(int32(l))
	slog.Info("log level changed", slog.String("level", lvl))
}

func GetLevel() string {
	l := slog.Level(level.Load())
	switch l {
	case slog.LevelDebug:
		return "debug"
	case slog.LevelWarn:
		return "warn"
	case slog.LevelError:
		return "error"
	default:
		return "info"
	}
}
