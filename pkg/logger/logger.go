package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"time"

	"be-ayaka/config"

	"gopkg.in/natefinch/lumberjack.v2"
)

// setupLogger initializes the logger with Lumberjack for log rotation
func SetUp(cfg *config.Config) {
	lumberjackLogger := &lumberjack.Logger{
		Filename:   cfg.Log.FilePath,
		MaxSize:    cfg.Log.MaxSize,    // megabytes
		MaxBackups: cfg.Log.MaxBackups, // number of backups
		MaxAge:     cfg.Log.MaxAge,     // days
		Compress:   cfg.Log.Compress,   // whether to compress old logs
	}

	multiWritter := io.MultiWriter(os.Stdout, lumberjackLogger)

	opts := &slog.HandlerOptions{
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// timestamp format
			if a.Key == slog.TimeKey {
				a.Key = "timestamp"
				a.Value = slog.StringValue(a.Value.Time().Format("2006-01-02 15:04:05"))
			}

			// msg -> message
			if a.Key == slog.MessageKey {
				a.Key = "message"
			}

			// source -> line
			if a.Key == slog.SourceKey {
				a.Key = "line"
			}

			// level -> level
			if a.Key == slog.LevelKey {
				a.Key = "level"
			}

			return a
		},
	}

	// JSON handler
	handler := slog.NewJSONHandler(multiWritter, opts)

	// add App Name and Port
	logger := slog.New(handler).With(
		slog.String("app", cfg.App.Name),
		slog.Int("port", cfg.Server.Port),
	)

	slog.SetDefault(logger)
}

// custom log for ayaka
func Log(userID string, levelStr string, message string) {
	var level slog.Level
	
	switch strings.ToUpper(levelStr) {
		case "DEBUG":
			level = slog.LevelDebug
		case "INFO":
			level = slog.LevelInfo
		case "WARN":
			level = slog.LevelWarn
		case "ERROR":
			level = slog.LevelError
		default:
			level = slog.LevelInfo
	}

	// capture caller info
	var pcs [1]uintptr
	// skip 2 levels to get the caller of Log()
	runtime.Callers(2, pcs[:])

	// create a log record with caller info and user ID
	record := slog.NewRecord(time.Now(), level, message, pcs[0])

	if userID != "" {
		record.AddAttrs(slog.String("user_id", userID))
	}

	// write the log record using the default logger's handler
	_ = slog.Default().Handler().Handle(context.Background(), record)
}