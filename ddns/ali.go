package ddns

import (
	"fmt"
	"log"
	"os"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
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

func NewAwsCredentialWithEnv() *AliCredential {
	region := os.Getenv("ALI_REGION")
	accessKeyID := os.Getenv("ALI_ACCESS_KEY")
	accessKeySecret := os.Getenv("ALI_SECRET_KEY")
	return NewAliCredential(accessKeyID, accessKeySecret, region)

}

func (cli *AliCredential) newClient() (*alidns.Client, error) {
	return alidns.NewClientWithAccessKey(cli.Region, cli.AccessKeyID, cli.AccessKeySecret)
}

func (cli *AliCredential) GetRecord(subDomain string) (*Record, error) {
	r := new(Record)
	client := cli.client
	request := alidns.CreateDescribeSubDomainRecordsRequest()
	request.Scheme = "https"
	request.SubDomain = subDomain
	response, err := client.DescribeSubDomainRecords(request)
	if err != nil {
		return nil, err
	}
	for _, record := range response.DomainRecords.Record {
		if record.RR+record.DomainName == subDomain {
			r.ID = record.RecordId
			r.RR = record.RR
			r.Type = record.Type
			r.Value = record.Value
			return r, nil
		}
	}
	return nil, fmt.Errorf("can`t found %s record.", subDomain)
}

func (cli *AliCredential) UpdateRecord(value string, r *Record) error {
	client := cli.client

	request := alidns.CreateUpdateDomainRecordRequest()
	request.Scheme = "https"

	request.RecordId = r.ID
	request.RR = r.RR
	request.Type = r.Type
	request.Value = value

	_, err := client.UpdateDomainRecord(request)
	if err != nil {
		return err
	}
	return nil
}
