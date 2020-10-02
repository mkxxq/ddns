package ddns

import (
	"testing"
)

func TestAwsCredential_getHostedZone(t *testing.T) {
	type args struct {
		domain string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "case 1", args: args{domain: "mkxxq.top"}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cre := NewAwsCredential()
			got, err := cre.getHostedZone(tt.args.domain)
			if (err != nil) != tt.wantErr {
				t.Errorf("AwsCredential.getHostedZone() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Logf("host zone ID: %v\n", got)

		})
	}
}
