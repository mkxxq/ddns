package utils

import "testing"

func TestGetOuterIp(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "case 1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetOuterIp()
			if err != nil {
				t.Error(err)
			}
			t.Log(got)
		})
	}
}
