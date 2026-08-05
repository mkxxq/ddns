package ddns

import (
	"log"

	"github.com/mkxxq/ddns/utils"
)

type DDns interface {
	UpsertRecord(value string, ip string) error
}
type Watcher struct {
	Domain    string
	latestIP  string
	IPGetter  utils.OuterIPGetter
	DNSServer string // 外部 DNS 服务器地址，如 "1.1.1.1:53"，为空则使用系统默认 DNS
}

func (w *Watcher) Run(cre DDns) {
	getter := w.IPGetter
	if getter == nil {
		getter = utils.NewJSONIPClient("", nil)
	}
	domainIp, err := w.lookupDomainIP()
	if err == nil && w.latestIP != domainIp {
		log.Printf("%s ip is %s, need changed!\n", w.Domain, domainIp)
		w.latestIP = domainIp
	}
	currentIP, err := getter.GetOuterIP()
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

// lookupDomainIP 查询域名的 IPv4 地址，若配置了外部 DNS 服务器则使用指定服务器查询
func (w *Watcher) lookupDomainIP() (string, error) {
	if w.DNSServer != "" {
		return utils.LookupDomainIPWithDNS(w.Domain, w.DNSServer)
	}
	return utils.LookupDomainIP(w.Domain)
}
