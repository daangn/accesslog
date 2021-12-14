package accesslog

import "strings"

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

// WithMetadata specifies headers to be captured by the logger.
// The key of the ms is the name of the metadata.
// And the value is the name to set in the log field.
// If the value is omitted, the name of the metadata is used as it is.
func WithMetadata(ms ...string) grpcOption {
	wms := metadataMap(ms)
	return func(cfg *grpcConfig) {
		cfg.metadata = wms
	}
}

func metadataMap(ms []string) map[string]string {
	mm := make(map[string]string, len(ms))
	for _, m := range ms {
		i := strings.Index(m, ":")
		switch {
		case i == -1:
			mm[m] = ""
		case i < len(m)-1:
			mm[m[:i]] = m[i+1:]
		default:
			mm[m[:i]] = ""
		}
	}
	return mm
}

// WithRequest specifies whether gRPC requests should be captured by the logger.
func WithRequest() grpcOption {
	return func(cfg *grpcConfig) {
		cfg.withResponse = true
	}
}

// WithResponse specifies whether gRPC responses should be captured by the logger.
func WithResponse() grpcOption {
	return func(cfg *grpcConfig) {
		cfg.withRequest = true
	}
}

// WithPeer specifies whether peer address should be captured by the logger.
func WithPeer() grpcOption {
	return func(cfg *grpcConfig) {
		cfg.withPeer = true
	}
}
