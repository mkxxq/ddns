package main

import (
	"flag"
	"time"

	"github.com/mkxxq/ddns"
)

func main() {
	var domain string
	flag.StringVar(&domain, "d", "www.google.com", "the domain name to be modified.")
	var ddnsType string
	flag.StringVar(&ddnsType, "t", "aws", "your dns provider")

	flag.Parse()

	w := ddns.Watcher{
		Domain: domain,
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
