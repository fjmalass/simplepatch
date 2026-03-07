package log

import (
	"io"
	"os"
	"sync"

	cl "github.com/charmbracelet/log"
)

type charmWrapper struct {
	inner  *cl.Logger
	mu     sync.Mutex
	file   *os.File
	closed bool
}

// Map our Level to charmbracelet Level
func toCharmLevel(l Level) cl.Level {
	switch l {
	case DebugLevel:
		return cl.DebugLevel
	case InfoLevel:
		return cl.InfoLevel
	case WarnLevel:
		return cl.WarnLevel
	case ErrorLevel:
		return cl.ErrorLevel
	default:
		return cl.InfoLevel
	}
}

// NewIconLogger creates a logger with text format
func NewIconLogger() Logger {
	l := cl.NewWithOptions(os.Stderr, cl.Options{
		Level:           cl.InfoLevel,
		ReportTimestamp: true,
	})

	return &charmWrapper{inner: l}
}

// NewJSONLogger creates a JSON-formatted logger
func NewJSONLogger() Logger {
	l := cl.NewWithOptions(os.Stderr, cl.Options{
		Level:           cl.InfoLevel,
		ReportTimestamp: true,
		Formatter:       cl.JSONFormatter,
	})

	return &charmWrapper{inner: l}
}

func (c *charmWrapper) SetLevel(level Level) {
	c.inner.SetLevel(toCharmLevel(level))
}

func (c *charmWrapper) AddFile(path string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return os.ErrClosed
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	multi := io.MultiWriter(os.Stderr, f)
	c.inner.SetOutput(multi)

	c.file = f
	return nil
}

func (c *charmWrapper) Debug(msg string, args ...any) {
	c.inner.Debug(msg, args...)
}

func (c *charmWrapper) Info(msg string, args ...any) {
	c.inner.Info(msg, args...)
}

func (c *charmWrapper) Warn(msg string, args ...any) {
	c.inner.Warn(msg, args...)
}

func (c *charmWrapper) Error(msg string, args ...any) {
	c.inner.Error(msg, args...)
}

func (c *charmWrapper) With(args ...any) Logger {
	return &charmWrapper{
		inner: c.inner.With(args...),
	}
}
