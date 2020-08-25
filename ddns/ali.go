package ddns

import (
	"fmt"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
)

type Ali struct {
	AccessKeyID     string
	AccessKeySecret string
	Region          string
}

func New() *Ali {
	return &Ali{}
}

func (cli *Ali) newClient() (*alidns.Client, error) {
	return alidns.NewClientWithAccessKey(cli.Region, cli.AccessKeyID, cli.AccessKeySecret)
}

func (cli *Ali) GetRecord(subDomain string) (*Record, error) {
	r := new(Record)
	client, err := cli.newClient()
	if err != nil {
		return nil, err
	}
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

func (cli *Ali) UpdateRecord(value string, r *Record) error {
	client, err := cli.newClient()
	if err != nil {
		return err
	}
	request := alidns.CreateUpdateDomainRecordRequest()
	request.Scheme = "https"

	request.RecordId = r.ID
	request.RR = r.RR
	request.Type = r.Type
	request.Value = value

	_, err = client.UpdateDomainRecord(request)
	if err != nil {
		return err
	}
	return nil
}
