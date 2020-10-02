package utils

import "testing"

func TestParseSubDomain(t *testing.T) {
	type args struct {
		subDomain string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 string
	}{
		{name: "case 1", args: args{subDomain: "google.com"}, want: "", want1: "google.com"},
		{name: "case 2", args: args{subDomain: "www.google.com"}, want: "www", want1: "google.com"},
		{name: "case 3", args: args{subDomain: "qiang.mail.google.com"}, want: "qiang", want1: "mail.google.com"},
		{name: "case 4", args: args{subDomain: "com"}, want: "", want1: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := ParseSubDomain(tt.args.subDomain)
			if got != tt.want {
				t.Errorf("ParseSubDomain() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("ParseSubDomain() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
