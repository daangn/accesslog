package reqlog

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/fluent/fluent-logger-golang/fluent"
)

const (
	DefaultWriteTimeout = time.Second
	DefaultBuffLimit    = 81_920
	DefaultMaxRetry     = 3
)

// FluentLogWriter is the log writer that implements io.Writer.
// It writes a log by Fluent Forwarder.
type FluentLogWriter struct {
	tag    string
	logger *fluent.Fluent
}

// NewFluentLogWriter creates a new FluentLogWriter.
func NewFluentLogWriter(tag, host string, port int) (*FluentLogWriter, error) {
	f, err := fluent.New(fluent.Config{
		FluentHost:   host,
		FluentPort:   port,
		Async:        true,
		MaxRetry:     DefaultMaxRetry,
		WriteTimeout: DefaultWriteTimeout,
		BufferLimit:  DefaultBuffLimit,
	})
	if err != nil {
		return nil, fmt.Errorf("new fluent log writer: %w", err)
	}

	return &FluentLogWriter{
		tag:    tag,
		logger: f,
	}, nil
}

// Close closes underlying connections with the Fluent daemon.
func (f *FluentLogWriter) Close() error {
	return fmt.Errorf("close fluent log writer: %w", f.logger.Close())
}

// Write writes a log.
func (f *FluentLogWriter) Write(p []byte) (n int, err error) {
	var m map[string]interface{}
	if err := json.Unmarshal(p, &m); err != nil {
		return 0, fmt.Errorf("fluent logger write: %w", err)
	}

	if err := f.logger.Post(f.tag, m); err != nil {
		return 0, fmt.Errorf("fluent logger write: %w", err)
	}

	return len(p), nil
}
