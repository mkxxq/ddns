package main

import (
	"flag"
	"log"
	"time"
	"watcher/ddns"
	"watcher/utils"
)

type watcher struct {
	domain   string
	latestIP string
}

func (w *watcher) Run(cre ddns.DDns) {
	currentIP, err := utils.GetOuterIp()
	if err != nil {
		log.Printf("get current ip err: %s\n", err)
		return
	}
	if currentIP == w.latestIP {
		return
	}
	err = cre.UpsertRecord(w.domain, currentIP)
	if err != nil {
		log.Printf("update %s value error: %s\n", w.domain, err)
		return
	}
	log.Printf("update domain: %s->%s success\n", w.domain, currentIP)
	w.latestIP = currentIP
}

func main() {
	var domain string
	flag.StringVar(&domain, "d", "www.google.com", "the domain name to be modified.")
	var ddnsType string
	flag.StringVar(&ddnsType, "t", "aws", "your dns provider")

	flag.Parse()

	w := watcher{
		domain: domain,
	}
	var cre ddns.DDns
	if ddnsType == "aws" {
		cre = ddns.NewAwsCredential()
	} else {
		cre = ddns.NewAwsCredentialEnv()
	}

	for {
		w.Run(cre)
		time.Sleep(time.Minute)
	}

}
