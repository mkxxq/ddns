package utils

import (
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

func LookupDomainIP(domain string) (string, error) {
	return lookupDomainIP(domain, net.LookupIP)
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
