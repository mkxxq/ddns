package ddns

import (
	"fmt"
	"log"
	"time"

	"github.com/mkxxq/ddns/utils"
)

const (
	// steadyInterval 是一致时的轮询间隔。
	steadyInterval = time.Minute
	// retryInterval 是出错后的重试间隔。家宽重拨后最长等一分钟才落盘,
	// 期间所有 CNAME 子域都还指向旧地址。
	retryInterval = 5 * time.Second
)

type DDns interface {
	UpsertRecord(value string, ip string) error
}

type Watcher struct {
	Domain string
	// DNSServers 是查询域名当前记录的 DNS 服务器列表("host:port"), 按顺序尝试。
	// 必须显式指定: 系统 resolver 可能指向旁路由, 答案不保证是权威侧的记录。
	DNSServers []string
	IPGetter   utils.OuterIPGetter
	// pendingIP 是最近一次写进 DNS 的地址。记录有 300s TTL, 写下去之后
	// 几轮查询读到的仍是旧值; 只靠"记录 == 出口 IP"判断会在这个窗口里
	// 反复重写同一条记录并刷日志。
	pendingIP string
	// lookup 仅供测试注入, 生产环境走 DNSServers。
	lookup func(string) (string, error)
}

// Run 执行一轮检查, 返回下一轮应该等待的时长: 出错时短间隔重试, 一致时等一个整分钟。
func (w *Watcher) Run(cre DDns) time.Duration {
	getter := w.IPGetter
	if getter == nil {
		getter = utils.NewJSONIPClient("", nil)
	}

	currentIP, err := getter.GetOuterIP()
	if err != nil {
		log.Printf("get current ip err: %s\n", err)
		return retryInterval
	}

	domainIP, err := w.lookupDomainIP()
	if err != nil {
		log.Printf("lookup domain %s err: %s\n", w.Domain, err)
	}

	if domainIP != "" && domainIP == currentIP {
		if w.pendingIP != "" {
			log.Printf("%s confirmed as %s\n", w.Domain, currentIP)
			w.pendingIP = ""
		}
		return steadyInterval
	}

	if currentIP == w.pendingIP {
		// 已经写过这个地址, 只是 DNS 缓存还没过期, 不要重复写。
		return steadyInterval
	}

	if domainIP != "" {
		log.Printf("%s ip is %s, need changed! (outer ip %s)\n", w.Domain, domainIP, currentIP)
	}

	if err := cre.UpsertRecord(w.Domain, currentIP); err != nil {
		log.Printf("update %s value error: %s\n", w.Domain, err)
		return retryInterval
	}
	log.Printf("update domain: %s->%s success\n", w.Domain, currentIP)
	w.pendingIP = currentIP
	return steadyInterval
}

// lookupDomainIP 查询域名的 IPv4 地址, 按配置顺序尝试各 DNS 服务器。
// 没有配置服务器时直接报错, 绝不静默退回系统 resolver(本机的指向旁路由)。
func (w *Watcher) lookupDomainIP() (string, error) {
	if w.lookup != nil {
		return w.lookup(w.Domain)
	}
	if len(w.DNSServers) == 0 {
		return "", fmt.Errorf("no dns server configured for %s", w.Domain)
	}
	return utils.LookupDomainIPWithDNSServers(w.Domain, w.DNSServers)
}
