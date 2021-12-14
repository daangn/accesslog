package accesslog

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
func WithMetadata(ms ...string) grpcOption {
	wm := make(map[string]struct{}, len(ms))
	for _, e := range ms {
		wm[e] = struct{}{}
	}
	return func(cfg *grpcConfig) {
		cfg.metadata = wm
	}
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
