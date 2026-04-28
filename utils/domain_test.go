package utils

import (
	"errors"
	"net"
	"testing"
)

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
			got, got1 := DecodeSubDomain(tt.args.subDomain)
			if got != tt.want {
				t.Errorf("ParseSubDomain() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("ParseSubDomain() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestLookupDomainIP(t *testing.T) {
	tests := []struct {
		name    string
		domain  string
		lookup  func(string) ([]net.IP, error)
		want    string
		wantErr bool
	}{
		{
			name:   "returns first ipv4",
			domain: "example.com",
			lookup: func(domain string) ([]net.IP, error) {
				return []net.IP{
					net.ParseIP("2001:db8::1"),
					net.ParseIP("1.2.3.4"),
					net.ParseIP("5.6.7.8"),
				}, nil
			},
			want: "1.2.3.4",
		},
		{
			name:   "empty domain",
			domain: " ",
			lookup: func(domain string) ([]net.IP, error) {
				t.Fatalf("lookup should not be called")
				return nil, nil
			},
			wantErr: true,
		},
		{
			name:   "lookup error",
			domain: "invalid.example",
			lookup: func(domain string) ([]net.IP, error) {
				return nil, errors.New("lookup failed")
			},
			wantErr: true,
		},
		{
			name:   "no ipv4",
			domain: "example.com",
			lookup: func(domain string) ([]net.IP, error) {
				return []net.IP{net.ParseIP("2001:db8::1")}, nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := lookupDomainIP(tt.domain, tt.lookup)
			if (err != nil) != tt.wantErr {
				t.Errorf("LookupDomainIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("LookupDomainIP() got = %v, want %v", got, tt.want)
			}
		})
	}
}
