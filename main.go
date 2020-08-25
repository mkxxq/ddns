package main

import (
	"flag"
	"log"
	"time"
	"watcher/utils"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
)

type watcher struct {
	accessKeyID     string
	accessKeySecret string
	region          string
	domain          string
	recordIP        string
	latestIP        string
}

func (w *watcher) getAliClient() (*alidns.Client, error) {
	return alidns.NewClientWithAccessKey(w.region, w.accessKeyID, w.accessKeySecret)
}

func (w *watcher) getLatestDomainRecord() ([]alidns.Record, error) {
	client, err := w.getAliClient()
	if err != nil {
		return nil, err
	}
	request := alidns.CreateDescribeSubDomainRecordsRequest()
	request.Scheme = "https"
	request.SubDomain = w.domain
	response, err := client.DescribeSubDomainRecords(request)
	if err != nil {
		return nil, err
	}
	return response.DomainRecords.Record, nil
}

func (w *watcher) updateDomainRecord(latestIP string, record *alidns.Record) error {
	client, err := w.getAliClient()
	if err != nil {
		return err
	}
	request := alidns.CreateUpdateDomainRecordRequest()
	request.Scheme = "https"

	request.RecordId = record.RecordId
	request.RR = record.RR
	request.Type = record.Type
	request.Value = latestIP

	_, err = client.UpdateDomainRecord(request)
	if err != nil {
		return err
	}
	return nil
}

func (w *watcher) Run() {
	latestIP, err := utils.GetOuterIp()
	if err != nil {
		log.Printf("get latest ip err: %s\n", err)
		return
	}
	if latestIP == w.latestIP {
		return
	}
	log.Printf("current ip: %s, latestip:%s", w.recordIP, latestIP)
	w.latestIP = latestIP
	records, err := w.getLatestDomainRecord()
	if err != nil {
		log.Printf("get latest domain records err: %s\n", err)
		return
	}
	for _, record := range records {
		if record.Value != w.latestIP {
			log.Printf("need update domain\n")
			err = w.updateDomainRecord(w.latestIP, &record)
			if err != nil {
				log.Printf("update domain failed, err:%s\n", err)
				continue
			}
			log.Printf("domain: %s, %s->%s success!", w.domain, w.recordIP, w.latestIP)
			w.recordIP = w.latestIP
		} else {
			w.recordIP = record.Value
		}
	}

}

func main() {
	var domain string
	flag.StringVar(&domain, "d", "www.google.com", "the domain name to be modified.")
	var accessKeyID string
	flag.StringVar(&accessKeyID, "i", "{aliyun access key id}", "your aliyun access key id.")
	var accessKeySecret string
	flag.StringVar(&accessKeySecret, "s", "{aliyun access key secret}", "your aliyun access key secret.")
	var region string
	flag.StringVar(&region, "r", "cn-hangzhou", "your aliyun region.")

	flag.Parse()

	w := watcher{
		domain:          domain,
		accessKeyID:     accessKeyID,
		accessKeySecret: accessKeySecret,
		region:          region,
	}

	for {
		w.Run()
		time.Sleep(time.Minute)
	}

}
