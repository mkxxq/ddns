package utils

import (
	"fmt"
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
