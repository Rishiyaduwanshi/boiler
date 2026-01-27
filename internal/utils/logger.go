package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Logger struct {
	logPath string
	verbose bool
}

type LogLevel string

const (
	LogLevelInfo  LogLevel = "INFO"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelError LogLevel = "ERROR"
	LogLevelDebug LogLevel = "DEBUG"
)

func NewLogger(logDir string, verbose bool) (*Logger, error) {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	timestamp := time.Now().Format("2006-01-02")
	logPath := filepath.Join(logDir, fmt.Sprintf("boiler-%s.log", timestamp))

	return &Logger{
		logPath: logPath,
		verbose: verbose,
	}, nil
}

func (l *Logger) Log(level LogLevel, message string) error {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logMessage := fmt.Sprintf("[%s] %s: %s\n", timestamp, level, message)

	f, err := os.OpenFile(l.logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(logMessage); err != nil {
		return fmt.Errorf("failed to write log: %w", err)
	}

	if l.verbose || level == LogLevelError || level == LogLevelWarn {
		fmt.Print(logMessage)
	}

	return nil
}

func (l *Logger) Info(message string) {
	_ = l.Log(LogLevelInfo, message)
}

func (l *Logger) Warn(message string) {
	_ = l.Log(LogLevelWarn, message)
}

func (l *Logger) Error(message string) {
	_ = l.Log(LogLevelError, message)
}

func (l *Logger) Debug(message string) {
	if l.verbose {
		_ = l.Log(LogLevelDebug, message)
	}
}
