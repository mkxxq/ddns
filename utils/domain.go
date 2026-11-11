package utils

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"
)

// lookupTimeout 是单次 DNS 查询(含 UDP 建连与解析器内部重试)的上界。
// Go 的 resolver 在没有 deadline 时会挂满 20s(实测), 期间整个监控循环停摆,
// 期间家宽换 IP 也无法落盘。
var lookupTimeout = 5 * time.Second

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
	return LookupDomainIPWithDNSServers(domain, []string{dnsServer})
}

// LookupDomainIPWithDNSServers 按顺序尝试给定的外部 DNS 服务器，第一个成功的答案即返回。
// 必须显式给服务器而不是走系统 resolver：本机的系统 DNS 指向旁路由，答案不保证权威，
// 拿它和出口 IP 比对会得出错误的"记录已过期"结论。
func LookupDomainIPWithDNSServers(domain string, servers []string) (string, error) {
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return "", fmt.Errorf("domain is empty")
	}

	var failures []string
	tried := 0
	for _, server := range servers {
		server = strings.TrimSpace(server)
		if server == "" {
			continue
		}
		tried++
		ip, err := lookupDomainIP(domain, ipLookupFunc(server))
		if err == nil {
			return ip, nil
		}
		failures = append(failures, fmt.Sprintf("%s: %s", server, err))
	}
	if tried == 0 {
		return "", fmt.Errorf("no dns server configured for %s", domain)
	}
	return "", fmt.Errorf("no dns server could resolve %s: %s", domain, strings.Join(failures, "; "))
}

// ipLookupFunc 返回只查询指定 DNS 服务器的查找函数，并给整次查询加上超时。
func ipLookupFunc(server string) func(string) ([]net.IP, error) {
	return func(host string) ([]net.IP, error) {
		resolver := &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{Timeout: lookupTimeout}
				return d.DialContext(ctx, "udp", server)
			},
		}
		ctx, cancel := context.WithTimeout(context.Background(), lookupTimeout)
		defer cancel()

		addrs, err := resolver.LookupIPAddr(ctx, host)
		if err != nil {
			return nil, err
		}
		ips := make([]net.IP, len(addrs))
		for i, addr := range addrs {
			ips[i] = addr.IP
		}
		return ips, nil
	}
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
