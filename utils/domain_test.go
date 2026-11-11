package utils

import (
	"errors"
	"net"
	"testing"
)

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
