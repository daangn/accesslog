package accesslog

import (
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
