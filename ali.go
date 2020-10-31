package ddns

import (
	"fmt"
	"log"
	"os"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
	"github.com/mkxxq/ddns/utils"
)

type AliCredential struct {
	AccessKeyID     string
	AccessKeySecret string
	Region          string
	client          *alidns.Client
}

func NewAliCredential(accessKeyID string, accessKeySecret string, region string) *AliCredential {
	cre := &AliCredential{AccessKeyID: accessKeyID, AccessKeySecret: accessKeySecret, Region: region}
	client, err := cre.newClient()
	if err != nil {
		log.Panicln(err)
	}
	cre.client = client
	return cre
}

func newAliCredentialWithEnv() (*AliCredential, error) {
	region := os.Getenv("ALI_REGION")
	if region == "" {
		return nil, fmt.Errorf("env value ALI_REGION not define")
	}
	accessKeyID := os.Getenv("ALI_ACCESS_KEY")

	if accessKeyID == "" {
		return nil, fmt.Errorf("env value ALI_ACCESS_KEY not define")
	}

	accessKeySecret := os.Getenv("ALI_SECRET_KEY")
	if accessKeySecret == "" {
		return nil, fmt.Errorf("env value ALI_SECRET_KEY not define")
	}
	return NewAliCredential(accessKeyID, accessKeySecret, region), nil
}

func NewAliCredentialWithEnv() *AliCredential {
	cre, err := newAliCredentialWithEnv()
	if err != nil {
		log.Panicln(err)
	}
	return cre
}

func (cli *AliCredential) newClient() (*alidns.Client, error) {
	return alidns.NewClientWithAccessKey(cli.Region, cli.AccessKeyID, cli.AccessKeySecret)
}

func (cli *AliCredential) getRecord(rr string, domain string) (*alidns.Record, error) {
	client := cli.client
	request := alidns.CreateDescribeSubDomainRecordsRequest()
	request.Scheme = "https"
	request.SubDomain = fmt.Sprintf("%s.%s", rr, domain)
	response, err := client.DescribeSubDomainRecords(request)
	if err != nil {
		return nil, err
	}
	for _, record := range response.DomainRecords.Record {
		if record.RR == rr && record.DomainName == domain {

			return &record, nil
		}
	}
	return nil, fmt.Errorf("can`t found %s.%s record.", rr, domain)
}

func (cli *AliCredential) UpsertRecord(subDomain string, ip string) error {
	client := cli.client
	rr, domain := utils.DecodeSubDomain(subDomain)
	record, err := cli.getRecord(rr, domain)
	if err != nil {
		request := alidns.CreateAddDomainRecordRequest()
		request.Scheme = "https"
		request.DomainName = domain
		request.Value = ip
		request.Type = "A"
		request.RR = rr
		_, err := client.AddDomainRecord(request)
		if err != nil {
			return err
		}
	} else {
		request := alidns.CreateUpdateDomainRecordRequest()
		request.Scheme = "https"

		request.RecordId = record.RecordId
		request.RR = record.RR
		request.Type = record.Type
		request.Value = ip

		_, err := client.UpdateDomainRecord(request)
		if err != nil {
			return err
		}
	}

	return nil

}
