package accesslog

import (
	"io"
	"os"
)

type config struct {
	writer           io.Writer
	httpLogFormatter HTTPLogFormatter
}

func defaults(cfg *config) {
	cfg.writer = os.Stdout
}

// Option represents an option that can be passed to middleware, interceptor or logger.
type Option func(cfg *config)

// WithWriter sets the given writer for the logger.
func WithWriter(writer io.Writer) Option {
	return func(cfg *config) {
		if writer != nil {
			cfg.writer = writer
		}
	}
}

// WithHTTPLogFormatter sets the given formatter for the logger.
func WithHTTPLogFormatter(formatter HTTPLogFormatter) Option {
	return func(cfg *config) {
		if formatter != nil {
			cfg.httpLogFormatter = formatter
		}
	}
}
