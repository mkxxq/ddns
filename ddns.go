package ddns

import (
	"log"

	"github.com/mkxxq/ddns/utils"
)

type DDns interface {
	UpsertRecord(value string, ip string) error
}
type Watcher struct {
	Domain   string
	latestIP string
}

func (w *Watcher) Run(cre DDns) {
	currentIP, err := utils.GetOuterIp()
	if err != nil {
		log.Printf("get current ip err: %s\n", err)
		return
	}
	if currentIP == w.latestIP {
		return
	}
	err = cre.UpsertRecord(w.Domain, currentIP)
	if err != nil {
		log.Printf("update %s value error: %s\n", w.Domain, err)
		return
	}
	log.Printf("update domain: %s->%s success\n", w.Domain, currentIP)
	w.latestIP = currentIP
}
