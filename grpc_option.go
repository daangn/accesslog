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

// WithMetadata specifies metadata to be captured by the logger. pseudo-headers in metadata also can be treated.
// If you want alias for logging, write like metadata:alias.
// e.g. "content-type:ct", this metadata will be logged like "ct": "[\"application/grpc\"]"
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
		if i == -1 {
			mm[m] = ""
		} else if i == 0 {
			li := strings.LastIndex(m, ":")
			if li == 0 {
				mm[m] = ""
			} else if li < len(m)-1 {
				mm[m[:li]] = m[li+1:]
			} else {
				mm[m[:li]] = ""
			}
		} else if i < len(m)-1 {
			mm[m[:i]] = m[i+1:]
		} else {
			mm[m[:i]] = ""
		}
	}
	return mm
}

// WithRequest specifies whether gRPC requests should be captured by the logger.
func WithRequest() grpcOption {
	return func(cfg *grpcConfig) {
		cfg.withRequest = true
	}
}

// WithResponse specifies whether gRPC responses should be captured by the logger.
func WithResponse() grpcOption {
	return func(cfg *grpcConfig) {
		cfg.withResponse = true
	}
}

// WithPeer specifies whether peer address should be captured by the logger.
func WithPeer() grpcOption {
	return func(cfg *grpcConfig) {
		cfg.withPeer = true
	}
}
