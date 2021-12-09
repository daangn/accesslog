package accesslog

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// DefaultGRPCLogger is default gRPC Logger.
var DefaultGRPCLogger = NewGRPCLogger(os.Stdout, NewDefaultGRPCLogFormatter())

// GRPCLogger is logger for gRPC access logging.
type GRPCLogger struct {
	l *zerolog.Logger
	f GRPCLogFormatter
}

// NewGRPCLogger returns a new GRPCLogger.
func NewGRPCLogger(w io.Writer, f GRPCLogFormatter) *GRPCLogger {
	l := zerolog.New(w)
	return &GRPCLogger{
		l: &l,
		f: f,
	}
}

// NewLogEntry returns a New LogEntry.
func (l *GRPCLogger) NewLogEntry(ctx context.Context, req interface{}, res *interface{}, info *grpc.UnaryServerInfo, err *error) LogEntry {
	return l.f.NewLogEntry(l.l, ctx, req, res, info, err)
}

// GRPCLogFormatter is the interface for NewLogEntry method.
type GRPCLogFormatter interface {
	NewLogEntry(l *zerolog.Logger, ctx context.Context, req interface{}, res *interface{}, info *grpc.UnaryServerInfo, err *error) LogEntry
}

type grpcConfig struct {
	ignoredMethods    map[string]struct{}
	withMetadataField bool
	ignoredMetadata   map[string]struct{}
	withRequestField  bool
	withResponseField bool
}

type grpcOption func(cfg *grpcConfig)

// WithIgnoredMethods specifies full methods to be ignored by the server side interceptor.
// When an incoming request's full method is in ms, the request will not be captured.
func WithIgnoredMethods(ms ...string) grpcOption {
	ims := make(map[string]struct{}, len(ms))
	for _, e := range ms {
		ims[e] = struct{}{}
	}
	return func(cfg *grpcConfig) {
		cfg.ignoredMethods = ims
	}
}

// WithMetadataField specifies whether gRPC metadata should be captured by the logger.
func WithMetadataField() grpcOption {
	return func(cfg *grpcConfig) {
		cfg.withMetadataField = true
	}
}

// WithIgnoredMetadata specifies keys to be ignored while tracing the metadata. Must be used
// in conjunction with WithMetadataField.
func WithIgnoredMetadata(ms ...string) grpcOption {
	ims := make(map[string]struct{}, len(ms))
	for _, e := range ms {
		ims[e] = struct{}{}
	}
	return func(cfg *grpcConfig) {
		cfg.ignoredMetadata = ims
	}
}

// WithRequestField specifies whether gRPC requests should be captured by the logger.
func WithRequestField() grpcOption {
	return func(cfg *grpcConfig) {
		cfg.withRequestField = true
	}
}

// WithResponseField specifies whether gRPC responses should be captured by the logger.
func WithResponseField() grpcOption {
	return func(cfg *grpcConfig) {
		cfg.withResponseField = true
	}
}

// DefaultGRPCLogFormatter is default GRPCLogFormatter.
type DefaultGRPCLogFormatter struct {
	cfg *grpcConfig
}

// NewDefaultGRPCLogFormatter returns a new DefaultGRPCLogFormatter.
func NewDefaultGRPCLogFormatter(opts ...grpcOption) *DefaultGRPCLogFormatter {
	cfg := new(grpcConfig)
	for _, fn := range opts {
		fn(cfg)
	}

	return &DefaultGRPCLogFormatter{cfg: cfg}
}

// NewLogEntry returns a New LogEntry formatted in DefaultGRPCLogFormatter.
func (f *DefaultGRPCLogFormatter) NewLogEntry(l *zerolog.Logger, ctx context.Context, req interface{}, res *interface{}, info *grpc.UnaryServerInfo, err *error) LogEntry {
	return &DefaultGRPCLogEntry{
		l:    l,
		cfg:  f.cfg,
		ctx:  ctx,
		req:  req,
		res:  res,
		info: info,
		add:  []func(e *zerolog.Event){},
		err:  err,
	}
}

// DefaultGRPCLogEntry is the LogEntry formatted in DefaultGRPCLogFormatter.
type DefaultGRPCLogEntry struct {
	l    *zerolog.Logger
	cfg  *grpcConfig
	ctx  context.Context
	req  interface{}
	res  *interface{}
	info *grpc.UnaryServerInfo
	add  []func(e *zerolog.Event)
	err  *error
}

// Add adds function for adding fields to log event.
func (le *DefaultGRPCLogEntry) Add(f func(e *zerolog.Event)) {
	le.add = append(le.add, f)
}

// Write writes a log.
func (le *DefaultGRPCLogEntry) Write(t time.Time) {
	if _, ok := le.cfg.ignoredMethods[le.info.FullMethod]; ok {
		return
	}

	e := le.l.Log().
		Str("protocol", "grpc").
		Str("method", le.info.FullMethod).
		Str("status", status.Code(*le.err).String()).
		Str("time", t.UTC().Format(time.RFC3339Nano)).
		Dur("elapsed(ms)", time.Since(t))

	if p, ok := peer.FromContext(le.ctx); ok {
		e.Str("peer", p.Addr.String())
	}

	if le.cfg.withMetadataField {
		md, _ := metadata.FromIncomingContext(le.ctx)
		ms := make(map[string][]string, len(md))
		for k, v := range md {
			if _, ok := le.cfg.ignoredMetadata[k]; !ok {
				ms[k] = v
			}
		}
		if len(ms) > 0 {
			b, err := json.Marshal(ms)
			if err == nil {
				e.Str("metadata", string(b))
			}
		}
	}
	if le.cfg.withRequestField {
		var m jsonpb.Marshaler
		if p, ok := le.req.(proto.Message); ok {
			if s, err := m.MarshalToString(p); err == nil {
				e.Str("req", s)
			}
		}
	}
	if le.cfg.withResponseField {
		var m jsonpb.Marshaler
		if p, ok := (*le.res).(proto.Message); ok {
			if s, err := m.MarshalToString(p); err == nil {
				e.Str("res", s)
			}
		}
	}

	for _, f := range le.add {
		f(e)
	}

	e.Send()
}
