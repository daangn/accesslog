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

// WithHeaders specifies headers to be captured by the logger. pseudo-headers also can be treated.
// If you want alias for logging, write like header:alias.
// e.g. "content-type:ct", this metadata will be logged like "ct": "application/json"
func WithHeaders(hs ...string) httpOption {
	whs := headerMap(hs)
	return func(cfg *httpConfig) {
		cfg.Headers = whs
	}
}

func headerMap(hs []string) map[string]string {
	hm := make(map[string]string, len(hs))
	for _, h := range hs {
		i := strings.Index(h, ":")
		if i == -1 {
			hm[h] = ""
		} else if i == 0 {
			li := strings.LastIndex(h, ":")
			if li == 0 {
				hm[h] = ""
			} else if li < len(h)-1 {
				hm[h[:li]] = h[li+1:]
			} else {
				hm[h[:li]] = ""
			}
		} else if i < len(h)-1 {
			hm[h[:i]] = h[i+1:]
		} else {
			hm[h[:i]] = ""
		}
	}
	return hm
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
