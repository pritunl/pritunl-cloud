package dns

import (
	"strings"
)

func matchDomains(x, y string) bool {
	if strings.Trim(x, ".") == strings.Trim(y, ".") {
		return true
	}
	return false
}

func matchTxt(x, y string) bool {
	if strings.Trim(x, "\"") == strings.Trim(y, "\"") {
		return true
	}
	return false
}

func extractDomain(domain string) string {
	domain = strings.Trim(domain, ".")
	parts := strings.Split(domain, ".")
	if len(parts) >= 2 {
		return parts[len(parts)-2] + "." + parts[len(parts)-1]
	}
	return domain
}

func cleanDomain(domain string) string {
	return strings.Trim(domain, ".")
}
