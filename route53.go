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
	"github.com/mkxxq/ddns/utils"
)

type AwsCredential struct {
	client *route53.Route53
}

func NewAwsCredential() *AwsCredential {
	cre := &AwsCredential{}
	err := cre.newClient()
	if err != nil {
		log.Panicln(err)
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

func (cre *AwsCredential) UpsertRecord(subDomain string, ip string) error {
	_, domain := utils.DecodeSubDomain(subDomain)
	if domain == "" {
		return fmt.Errorf("error subDomain: %s", subDomain)
	}
	hostZoneID, err := cre.getHostedZone(domain)
	if err != nil {
		return err
	}
	input := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String("UPSERT"),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name: aws.String(subDomain),
						ResourceRecords: []*route53.ResourceRecord{
							{Value: aws.String(ip)},
						},
						TTL:  aws.Int64(60),
						Type: aws.String("A"),
					},
				},
			},
		},
		HostedZoneId: aws.String(hostZoneID),
	}
	_, err = cre.client.ChangeResourceRecordSets(input)
	if err != nil {
		return err
	}
	return nil
}
