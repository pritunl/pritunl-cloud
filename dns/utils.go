package dns

import (
	"strings"
)

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
