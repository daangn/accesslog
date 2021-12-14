package accesslog

import (
	"reflect"
	"testing"
)

func Test_metadataMap(t *testing.T) {
	tests := []struct {
		name string
		ms   []string
		want map[string]string
	}{
		{
			name: "just metadata",
			ms:   []string{"user-agent"},
			want: map[string]string{
				"user-agent": "",
			},
		},
		{
			name: "metadata with alias",
			ms:   []string{"user-agent:ua"},
			want: map[string]string{
				"user-agent": "ua",
			},
		},
		{
			name: "empty alias",
			ms:   []string{"user-agent:"},
			want: map[string]string{
				"user-agent": "",
			},
		},
		{
			name: "pseudo-header",
			ms:   []string{":authority"},
			want: map[string]string{
				":authority": "",
			},
		},
		{
			name: "pseudo-header with alias",
			ms:   []string{":authority:a"},
			want: map[string]string{
				":authority": "a",
			},
		},
		{
			name: "pseudo-header with empty alias",
			ms:   []string{":authority:"},
			want: map[string]string{
				":authority": "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := metadataMap(tt.ms); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("metadataMap() = %v, want %v", got, tt.want)
			}
		})
	}
}
