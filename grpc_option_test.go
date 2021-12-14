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
			name: "all items are just header names",
			ms:   []string{"user-agent", "authority"},
			want: map[string]string{
				"user-agent": "",
				"authority":  "",
			},
		},
		{
			name: "some items have alias",
			ms:   []string{"user-agent:ua", "authority"},
			want: map[string]string{
				"user-agent": "ua",
				"authority":  "",
			},
		},
		{
			name: "invalid case",
			ms:   []string{"user-agent:", "authority"},
			want: map[string]string{
				"user-agent": "",
				"authority":  "",
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
