package accesslog

import (
	"strings"
)

type httpOption func(cfg *httpConfig)

// WithIgnoredPaths specifies methods and paths to be ignored by the logger.
// See path.Match method how to set path patterns
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
		cfg.headers = whs
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
