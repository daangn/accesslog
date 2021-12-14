package accesslog

import (
	"net/http"
	"strings"
)

type httpOption func(cfg *httpConfig)

// WithIgnoredPaths specifies methods and paths to be ignored by the logger.
// This only works when using chi.Router.
func WithIgnoredPaths(ips map[string][]string) httpOption {
	return func(cfg *httpConfig) {
		cfg.ignoredPaths = ips
	}
}

// WithHeaders specifies headers to be captured by the logger.
// The key of the hs is the name of the header.
// And the value is the name to set in the log field.
// If the value is omitted, the name of the header is used as it is.
func WithHeaders(hs map[string]string) httpOption {
	return func(cfg *httpConfig) {
		cfg.Headers = hs
	}
}

// WithClientIP specifies whether client ip should be captured by the logger.
func WithClientIP() httpOption {
	return func(cfg *httpConfig) {
		cfg.withClientIP = true
	}
}

var (
	trueClientIP          = http.CanonicalHeaderKey("True-Client-IP")
	xForwardedFor         = http.CanonicalHeaderKey("X-Forwarded-For")
	xRealIP               = http.CanonicalHeaderKey("X-Real-IP")
	xEnvoyExternalAddress = http.CanonicalHeaderKey("X-Envoy-External-Address")
)

func clientIP(h http.Header) string {
	if tcip := h.Get(trueClientIP); tcip != "" {
		return tcip
	} else if xrip := h.Get(xRealIP); xrip != "" {
		return xrip
	} else if xff := h.Get(xForwardedFor); xff != "" {
		i := strings.Index(xff, ",")
		if i == -1 {
			i = len(xff)
		}
		return xff[:i]
	} else if xeea := h.Get(xEnvoyExternalAddress); xeea != "" {
		return xeea
	}

	return ""
}
