package log

import (
	"io"
	"os"
	"sync"

	cl "github.com/charmbracelet/log"
)

// Option configures a Logger during construction
type Option func(*charmWrapper) error

// WithFile adds file output (logs to both stderr and file)
func WithFile(path string) Option {
	return func(c *charmWrapper) error {
		return c.AddFile(path)
	}
}

// WithLevel sets the log level
func WithLevel(level Level) Option {
	return func(c *charmWrapper) error {
		c.SetLevel(level)
		return nil
	}
}

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

// NewIconLogger creates a logger with text format and emoji prefixes
func NewIconLogger(opts ...Option) (Logger, error) {
	l := cl.NewWithOptions(os.Stderr, cl.Options{
		Level:           cl.InfoLevel,
		ReportTimestamp: true,
	})

	// Set emoji-style level prefixes
	styles := cl.DefaultStyles()
	styles.Levels[cl.DebugLevel] = styles.Levels[cl.DebugLevel].SetString("🐛")
	styles.Levels[cl.InfoLevel] = styles.Levels[cl.InfoLevel].SetString("💡")
	styles.Levels[cl.WarnLevel] = styles.Levels[cl.WarnLevel].SetString("⚠️")
	styles.Levels[cl.ErrorLevel] = styles.Levels[cl.ErrorLevel].SetString("🚨")
	l.SetStyles(styles)

	wrapper := &charmWrapper{inner: l}

	for _, opt := range opts {
		if err := opt(wrapper); err != nil {
			return nil, err
		}
	}

	return wrapper, nil
}

// NewJSONLogger creates a JSON-formatted logger
func NewJSONLogger(opts ...Option) (Logger, error) {
	l := cl.NewWithOptions(os.Stderr, cl.Options{
		Level:           cl.InfoLevel,
		ReportTimestamp: true,
		Formatter:       cl.JSONFormatter,
	})

	wrapper := &charmWrapper{inner: l}

	for _, opt := range opts {
		if err := opt(wrapper); err != nil {
			return nil, err
		}
	}

	return wrapper, nil
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
