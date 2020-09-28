package ddns

import (
	"fmt"
	"log"
	"net/http"
	"strings"

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

// host zone id like "/hostedzone/Z0693175UDCZQTRUIO8H", wo only need "Z0693175UDCZQTRUIO8H"
func decodeHostedZoneID(id string) string {
	return strings.Replace(id, "/hostedzone/", "", -1)

}

func (cre *AwsCredential) getHostedZone(domain string) (string, error) {
	input := route53.ListHostedZonesByNameInput{
		DNSName: aws.String(domain),
	}
	client := cre.client
	output, err := client.ListHostedZonesByName(&input)
	if err != nil {
		return "", err
	}
	if !strings.HasSuffix(domain, ".") {
		domain = domain + "."
	}

	for _, zone := range output.HostedZones {
		if *zone.Name == domain {
			return decodeHostedZoneID(*zone.Id), nil
		}
	}
	return "", fmt.Errorf("%s, not found", domain)
}

func (cre *AwsCredential) GetRecord(subDomain string) (*Record, error) {

	input := route53.ListResourceRecordSetsInput{}
	r := new(Record)
	client := cre.client


	return r, nil
}
