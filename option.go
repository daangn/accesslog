package reqlog

import (
	"context"
	"io"
	"os"
)

type config struct {
	userGetter       UserGetter
	writer           io.Writer
	httpLogFormatter HTTPLogFormatter
}

func defaults(cfg *config) {
	cfg.writer = os.Stdout
	cfg.httpLogFormatter = &DefaultHTTPLogFormatter{}
}

// Option represents an option that can be passed to middleware, interceptor or logger.
type Option func(cfg *config)

// UserGetter is the interface for all types that implement getting user id.
type UserGetter interface {
	GetUserID(ctx context.Context) int64
}

// WithUserGetter sets the given user getter for the logger.
func WithUserGetter(getter UserGetter) Option {
	return func(cfg *config) {
		cfg.userGetter = getter
	}
}

// WithWriter sets the given writer for the logger.
func WithWriter(writer io.Writer) Option {
	return func(cfg *config) {
		cfg.writer = writer
	}
}

// WithHTTPLogFormatter sets the given formatter for the logger.
func WithHTTPLogFormatter(formater HTTPLogFormatter) Option {
	return func(cfg *config) {
		cfg.httpLogFormatter = formater
	}
}
