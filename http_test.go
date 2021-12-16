package accesslog

import (
	"net/http"
	"net/url"
	"testing"
)

func TestDefaultHTTPLogEntry_isIgnored(t *testing.T) {
	type fields struct {
		cfg *httpConfig
		r   *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "ignored",
			fields: fields{
				cfg: &httpConfig{
					ignoredPaths: map[string][]string{
						"GET": {"/abc"},
					},
				},
				r: &http.Request{
					Method: "GET",
					URL: &url.URL{Path: "/abc"},
				},
			},
			want: true,
		},
		{
			name: "method not matched",
			fields: fields{
				cfg: &httpConfig{
					ignoredPaths: map[string][]string{
						"GET": {"/abc"},
					},
				},
				r: &http.Request{
					Method: "POST",
					URL: &url.URL{Path: "/abc"},
				},
			},
			want: false,
		},
		{
			name: "with slash",
			fields: fields{
				cfg: &httpConfig{
					ignoredPaths: map[string][]string{
						"GET": {"/abc/def"},
					},
				},
				r: &http.Request{
					Method: "GET",
					URL:    &url.URL{Path: "/abc/def"},
				},
			},
			want: true,
		},
		{
			name: "path not start with slash",
			fields: fields{
				cfg: &httpConfig{
					ignoredPaths: map[string][]string{
						"GET": {"/abc/def"},
					},
				},
				r: &http.Request{
					Method: "GET",
					URL:    &url.URL{Path: "abc/def"},
				},
			},
			want: true,
		},
		{
			name: "with asterisk",
			fields: fields{
				cfg: &httpConfig{
					ignoredPaths: map[string][]string{
						"GET": {"/a*/def"},
					},
				},
				r: &http.Request{
					Method: "GET",
					URL:    &url.URL{Path: "/afasdfsd/def"},
				},
			},
			want: true,
		},
		{
			name: "with asterisk 2",
			fields: fields{
				cfg: &httpConfig{
					ignoredPaths: map[string][]string{
						"GET": {"/abc/*/def"},
					},
				},
				r: &http.Request{
					Method: "GET",
					URL:    &url.URL{Path: "/abc/fasdfsd/def"},
				},
			},
			want: true,
		},
		{
			name: "with asterisk but path has slash",
			fields: fields{
				cfg: &httpConfig{
					ignoredPaths: map[string][]string{
						"GET": {"/ab*"},
					},
				},
				r: &http.Request{
					Method: "GET",
					URL:    &url.URL{Path: "/abce/def"},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			le := &DefaultHTTPLogEntry{
				cfg: tt.fields.cfg,
				r:   tt.fields.r,
			}
			if got := le.isIgnored(); got != tt.want {
				t.Errorf("isIgnored() = %v, want %v", got, tt.want)
			}
		})
	}
}
