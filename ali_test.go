package ddns

import (
	"testing"
)

func TestAliCredential_UpsertRecord(t *testing.T) {
	type args struct {
		subDomain string
		ip        string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "case 1", args: args{subDomain: "test.mkxxq.top", ip: "127.0.0.2"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli := NewAliCredentialWithEnv()
			if err := cli.UpsertRecord(tt.args.subDomain, tt.args.ip); (err != nil) != tt.wantErr {
				t.Errorf("AliCredential.UpsertRecord() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
