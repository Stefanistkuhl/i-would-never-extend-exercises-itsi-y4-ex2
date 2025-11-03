package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

type Logger interface {
	Info(msg any, keysAndValues ...any)
	Debug(msg any, keysAndValues ...any)
	Warn(msg any, keysAndValues ...any)
	Error(msg any, keysAndValues ...any)
	Fatal(msg any, keysAndValues ...any)
	Print(msg any, keysAndValues ...any)
}

type MultiLogger struct {
	stdout *log.Logger
	file   *log.Logger
}

func (m *MultiLogger) Info(msg any, keysAndValues ...any) {
	m.stdout.Info(msg, keysAndValues...)
	m.file.Info(msg, keysAndValues...)
}

func (m *MultiLogger) Debug(msg any, keysAndValues ...any) {
	m.stdout.Debug(msg, keysAndValues...)
	m.file.Debug(msg, keysAndValues...)
}

func (m *MultiLogger) Warn(msg any, keysAndValues ...any) {
	m.stdout.Warn(msg, keysAndValues...)
	m.file.Warn(msg, keysAndValues...)
}

func (m *MultiLogger) Error(msg any, keysAndValues ...any) {
	m.stdout.Error(msg, keysAndValues...)
	m.file.Error(msg, keysAndValues...)
}

func (m *MultiLogger) Fatal(msg any, keysAndValues ...any) {
	m.stdout.Fatal(msg, keysAndValues...)
	m.file.Fatal(msg, keysAndValues...)
}

func (m *MultiLogger) Print(msg any, keysAndValues ...any) {
	m.stdout.Print(msg, keysAndValues...)
	m.file.Print(msg, keysAndValues...)
}

func New(prefix string, logDir string) (Logger, error) {
	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	logFile, err := os.OpenFile(
		filepath.Join(logDir, fmt.Sprintf("pcap-sorter-%s.log", time.Now().Format("2006-01-02"))),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0666,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	stdoutLogger := log.New(os.Stdout)
	stdoutLogger.SetPrefix(prefix)
	stdoutLogger.SetLevel(log.InfoLevel)
	styles := log.DefaultStyles()
	styles.Levels[log.InfoLevel] = styles.Levels[log.InfoLevel].Foreground(lipgloss.Color("#FF69B4"))
	styles.Levels[log.ErrorLevel] = styles.Levels[log.ErrorLevel].Foreground(lipgloss.Color("#FF1493"))
	styles.Levels[log.WarnLevel] = styles.Levels[log.WarnLevel].Foreground(lipgloss.Color("#FFB6C1"))
	styles.Levels[log.FatalLevel] = styles.Levels[log.FatalLevel].Foreground(lipgloss.Color("#DC143C"))
	stdoutLogger.SetStyles(styles)

	fileLogger := log.NewWithOptions(logFile, log.Options{
		Prefix:    prefix,
		Level:     log.InfoLevel,
		Formatter: log.JSONFormatter,
	})

	return &MultiLogger{
		stdout: stdoutLogger,
		file:   fileLogger,
	}, nil
}
