package writer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/fluent/fluent-logger-golang/fluent"
)

const (
	defaultWriteTimeout = time.Second
	defaultBuffLimit    = 81_920
	defaultMaxRetry     = 3
)

// FluentLogWriter is the log writer that implements io.Writer.
// It writes a log by Fluent Forwarder.
type FluentLogWriter struct {
	tag       string
	forwarder *fluent.Fluent
}

// NewFluentLogWriter creates a new FluentLogWriter.
func NewFluentLogWriter(tag, host string, port int) (*FluentLogWriter, error) {
	f, err := fluent.New(fluent.Config{
		FluentHost:   host,
		FluentPort:   port,
		Async:        true,
		MaxRetry:     defaultMaxRetry,
		WriteTimeout: defaultWriteTimeout,
		BufferLimit:  defaultBuffLimit,
	})
	if err != nil {
		return nil, fmt.Errorf("new fluent log writer: %w", err)
	}

	return NewFluentLogWriterFromForwarder(tag, f), nil
}

// NewFluentLogWriterFromForwarder creates a new FluentLigWriter from Fluent Forwarder.
func NewFluentLogWriterFromForwarder(tag string, f *fluent.Fluent) *FluentLogWriter {
	return &FluentLogWriter{
		tag:       tag,
		forwarder: f,
	}
}

// Close closes underlying connections with the Fluent daemon.
func (f *FluentLogWriter) Close() error {
	if err := f.forwarder.Close(); err != nil {
		return fmt.Errorf("close fluent log writer: %w", err)
	}
	return nil
}

// Write writes a log.
func (f *FluentLogWriter) Write(p []byte) (n int, err error) {
	var m map[string]interface{}
	if err := json.Unmarshal(p, &m); err != nil {
		return 0, fmt.Errorf("fluent logger write: %w", err)
	}

	// Keep Unmarshal's whole-input validation, then preserve integer digits for Fluent.
	decoder := json.NewDecoder(bytes.NewReader(p))
	decoder.UseNumber()
	if err := decoder.Decode(&m); err != nil {
		return 0, fmt.Errorf("fluent logger write: %w", err)
	}
	message, err := convertJSONNumbers(m)
	if err != nil {
		return 0, fmt.Errorf("fluent logger write: %w", err)
	}

	if err := f.forwarder.Post(f.tag, message); err != nil {
		return 0, fmt.Errorf("fluent logger write: %w", err)
	}

	return len(p), nil
}

func convertJSONNumbers(value interface{}) (interface{}, error) {
	switch value := value.(type) {
	case json.Number:
		if value != "-0" {
			if integer, err := value.Int64(); err == nil {
				return integer, nil
			}
			if integer, err := strconv.ParseUint(value.String(), 10, 64); err == nil {
				return integer, nil
			}
		}
		return value.Float64()
	case map[string]interface{}:
		for key, nested := range value {
			converted, err := convertJSONNumbers(nested)
			if err != nil {
				return nil, err
			}
			value[key] = converted
		}
	case []interface{}:
		for index, nested := range value {
			converted, err := convertJSONNumbers(nested)
			if err != nil {
				return nil, err
			}
			value[index] = converted
		}
	}
	return value, nil
}
