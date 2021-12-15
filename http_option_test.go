package accesslog

import (
	"net/http"
	"reflect"
	"testing"
)

func Test_headerMap(t *testing.T) {
	tests := []struct {
		name string
		hs   []string
		want map[string]string
	}{
		{
			name: "just header",
			hs:   []string{"user-agent"},
			want: map[string]string{
				"user-agent": "",
			},
		},
		{
			name: "header with alias",
			hs:   []string{"user-agent:ua"},
			want: map[string]string{
				"user-agent": "ua",
			},
		},
		{
			name: "empty alias",
			hs:   []string{"user-agent:"},
			want: map[string]string{
				"user-agent": "",
			},
		},
		{
			name: "pseudo-header",
			hs:   []string{":authority"},
			want: map[string]string{
				":authority": "",
			},
		},
		{
			name: "pseudo-header with alias",
			hs:   []string{":authority:a"},
			want: map[string]string{
				":authority": "a",
			},
		},
		{
			name: "pseudo-header with empty alias",
			hs:   []string{":authority:"},
			want: map[string]string{
				":authority": "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := headerMap(tt.hs); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("headerMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_clientIP(t *testing.T) {
	tests := []struct {
		name string
		h    http.Header
		want string
	}{
		{
			name: "true-client-ip",
			h: http.Header{
				"True-Client-Ip": []string{"255.255.255.255"},
			},
			want: "255.255.255.255",
		},
		{
			name: "x-forwarded-for",
			h: http.Header{
				"X-Forwarded-For": []string{"255.255.255.255"},
			},
			want: "255.255.255.255",
		},
		{
			name: "x-real-ip",
			h: http.Header{
				"X-Real-Ip": []string{"255.255.255.255"},
			},
			want: "255.255.255.255",
		},
		{
			name: "x-envoy-external-address",
			h: http.Header{
				"X-Envoy-External-Address": []string{"255.255.255.255"},
			},
			want: "255.255.255.255",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := clientIP(tt.h); got != tt.want {
				t.Errorf("clientIP() = %v, want %v", got, tt.want)
			}
		})
	}
}
