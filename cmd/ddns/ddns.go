package main

import (
	"flag"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/mkxxq/ddns"
	"github.com/mkxxq/ddns/utils"
)

// defaultDNSServers 是查询域名当前记录用的解析器, 逗号分隔、按顺序尝试。
// 不用系统 resolver: 本机 resolver 指向旁路由, 答案不保证是权威侧记录,
// 拿它和出口 IP 比对会误判"记录已过期"。
const defaultDNSServers = "1.1.1.1:53,223.5.5.5:53"

func main() {
	var domain string
	flag.StringVar(&domain, "d", "", "the domain name to be modified (required).")
	var ddnsType string
	flag.StringVar(&ddnsType, "t", "cloudflare", "your dns provider, only cloudflare is supported.")
	var ipProvider string
	flag.StringVar(&ipProvider, "ip-provider", envOrDefault("DDNS_IP_PROVIDER", "jsonip"), "outer ip provider, support jsonip and routeros.")
	var routerOSAddr string
	flag.StringVar(&routerOSAddr, "routeros-addr", envOrDefault("ROUTEROS_ADDR", ""), "routeros api addr, default port 8728.")
	var routerOSUser string
	flag.StringVar(&routerOSUser, "routeros-user", envOrDefault("ROUTEROS_USER", ""), "routeros username.")
	var routerOSPass string
	flag.StringVar(&routerOSPass, "routeros-pass", envOrDefault("ROUTEROS_PASS", ""), "routeros password.")
	var routerOSInterface string
	flag.StringVar(&routerOSInterface, "routeros-interface", envOrDefault("ROUTEROS_INTERFACE", ""), "routeros interface name for export ip.")
	var dnsServers string
	flag.StringVar(&dnsServers, "dns-server", envOrDefault("DDNS_DNS_SERVER", defaultDNSServers), "external dns servers for domain lookup, comma separated, tried in order.")

	flag.Parse()

	if domain == "" {
		log.Panicln("-d is required")
	}

	servers := splitDNSServers(dnsServers)
	w := ddns.Watcher{
		Domain:     domain,
		IPGetter:   mustNewIPGetter(ipProvider, routerOSAddr, routerOSUser, routerOSPass, routerOSInterface),
		DNSServers: servers,
	}
	var cre ddns.DDns
	switch ddnsType {
	case "cloudflare", "cf":
		cre = ddns.NewCloudflareCredential()
	default:
		log.Panicf("unsupported dns provider: %s\n", ddnsType)
	}

	log.Printf("watching %s, dns servers %v, ip provider %s\n", domain, servers, ipProvider)
	for {
		time.Sleep(w.Run(cre))
	}
}

func splitDNSServers(value string) []string {
	var servers []string
	for _, part := range strings.Split(value, ",") {
		server := strings.TrimSpace(part)
		if server == "" {
			continue
		}
		// 裸 IP 补默认端口, 否则 net.Resolver 会拨 "-53" 之类的无效地址。
		if _, _, err := net.SplitHostPort(server); err != nil {
			server = net.JoinHostPort(server, "53")
		}
		servers = append(servers, server)
	}
	return servers
}

func envOrDefault(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func mustNewIPGetter(provider string, routerOSAddr string, routerOSUser string, routerOSPass string, routerOSInterface string) utils.OuterIPGetter {
	switch provider {
	case "", "jsonip":
		return utils.NewJSONIPClient("", nil)
	case "routeros":
		getter := utils.NewRouterOSIPClient(routerOSAddr, routerOSUser, routerOSPass, routerOSInterface)
		if getter.Addr == "" || getter.Username == "" || getter.Password == "" || getter.InterfaceName == "" {
			log.Panicln("routeros provider require ROUTEROS_ADDR, ROUTEROS_USER, ROUTEROS_PASS and ROUTEROS_INTERFACE")
		}
		return getter
	default:
		log.Panicf("unsupported ip provider: %s\n", provider)
	}
	return nil
}
