package ddns

import (
	"log"
	"net/http"

	"golang.org/x/net/http2"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
)

type AwsCredential struct {
	client *route53.Route53
}

func NewAwsCredential() *AwsCredential {
	cre := &AwsCredential{}
	err := cre.newClient()
	if err != nil {
		log.Panic(err)
	}
	return cre
}

func NewAwsCredentialEnv() *AwsCredential {
	return NewAwsCredential()

}

func newHTTPClient() *http.Client {
	tr := &http.Transport{}
	_ = http2.ConfigureTransport(tr)
	return &http.Client{Transport: tr}
}

func newAwsSession() (*session.Session, error) {
	cfg := &aws.Config{
		MaxRetries: aws.Int(3),
		HTTPClient: newHTTPClient(),
	}
	return session.NewSession(cfg)
}

func (cre *AwsCredential) newClient() error {
	sess, err := newAwsSession()
	if err != nil {
		return err
	}
	cre.client = route53.New(sess)
	return nil
}
