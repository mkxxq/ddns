package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/mkxxq/ddns"
	"github.com/mkxxq/ddns/utils"
)

func main() {
	var domain string
	flag.StringVar(&domain, "d", "www.google.com", "the domain name to be modified.")
	var ddnsType string
	flag.StringVar(&ddnsType, "t", "aws", "your dns provider, support aws and cloudflare.")
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
	var dnsServer string
	flag.StringVar(&dnsServer, "dns-server", envOrDefault("DDNS_DNS_SERVER", "1.1.1.1:53"), "external dns server addr for domain lookup, default 1.1.1.1:53")

	flag.Parse()

	w := ddns.Watcher{
		Domain:    domain,
		IPGetter:  mustNewIPGetter(ipProvider, routerOSAddr, routerOSUser, routerOSPass, routerOSInterface),
		DNSServer: dnsServer,
	}
	var cre ddns.DDns
	switch ddnsType {
	case "aws":
		cre = ddns.NewAwsCredential()
	case "cloudflare", "cf":
		cre = ddns.NewCloudflareCredential()
	default:
		log.Panicf("unsupported dns provider: %s\n", ddnsType)
	}

	for {
		w.Run(cre)
		time.Sleep(time.Minute)
	}

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
