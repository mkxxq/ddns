package utils

import (
	"context"
	"fmt"
	"net"
	"strings"
)

func DecodeSubDomain(subDomain string) (string, string) {
	labels := strings.Split(subDomain, ".")

	if len(labels) > 2 {
		return labels[0], strings.Join(labels[1:], ".")
	} else if len(labels) == 2 {
		return "", subDomain
	} else {
		return "", ""
	}
}

func EncodeSubDomain(rr string, domain string) string {
	return fmt.Sprintf("%s.%s", rr, domain)
}

// LookupDomainIP 使用系统默认 DNS 查询域名的 IPv4 地址
func LookupDomainIP(domain string) (string, error) {
	return lookupDomainIP(domain, net.LookupIP)
}

// LookupDomainIPWithDNS 使用指定的外部 DNS 服务器查询域名的 IPv4 地址
// dnsServer 格式为 "host:port"，如 "1.1.1.1:53"
func LookupDomainIPWithDNS(domain, dnsServer string) (string, error) {
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{}
			return d.DialContext(ctx, "udp", dnsServer)
		},
	}
	// 将 resolver.LookupIPAddr 包装成与 net.LookupIP 相同的签名
	lookup := func(host string) ([]net.IP, error) {
		addrs, err := resolver.LookupIPAddr(context.Background(), host)
		if err != nil {
			return nil, err
		}
		ips := make([]net.IP, len(addrs))
		for i, addr := range addrs {
			ips[i] = addr.IP
		}
		return ips, nil
	}
	return lookupDomainIP(domain, lookup)
}

func lookupDomainIP(domain string, lookup func(string) ([]net.IP, error)) (string, error) {
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return "", fmt.Errorf("domain is empty")
	}

	ips, err := lookup(domain)
	if err != nil {
		return "", fmt.Errorf("lookup domain %s: %w", domain, err)
	}

	for _, ip := range ips {
		ipv4 := ip.To4()
		if ipv4 != nil {
			return ipv4.String(), nil
		}
	}

	return "", fmt.Errorf("no ipv4 record found for domain %s", domain)
}
