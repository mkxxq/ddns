package utils

import (
	"strings"
)

func ParseSubDomain(subDomain string) (string, string) {
	labels := strings.Split(subDomain, ".")

	if len(labels) > 2 {
		return labels[0], strings.Join(labels[1:], ".")
	} else if len(labels) == 2 {
		return "", subDomain
	} else {
		return "", ""
	}
}
