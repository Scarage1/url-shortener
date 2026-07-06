package security

import (
	"errors"
	"fmt"
	"net"
	neturl "net/url"
	"strings"
)

var ErrUnsafeURL = errors.New("unsafe URL")

type RulesScanner struct {
	blockedDomains map[string]struct{}
}

func NewRulesScanner(blockedDomains []string) URLScanner {

	domains := make(map[string]struct{}, len(blockedDomains))

	for _, domain := range blockedDomains {
		normalized := normalizeHost(domain)
		if normalized != "" {
			domains[normalized] = struct{}{}
		}
	}

	return &RulesScanner{
		blockedDomains: domains,
	}
}

func (s *RulesScanner) Check(rawURL string) error {

	parsedURL, err := neturl.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("%w: invalid URL", ErrUnsafeURL)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("%w: unsupported URL scheme", ErrUnsafeURL)
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("%w: missing host", ErrUnsafeURL)
	}

	if parsedURL.User != nil {
		return fmt.Errorf("%w: embedded credentials are not allowed", ErrUnsafeURL)
	}

	host := normalizeHost(parsedURL.Hostname())
	if host == "" {
		return fmt.Errorf("%w: invalid host", ErrUnsafeURL)
	}

	if isLocalHost(host) {
		return fmt.Errorf("%w: local destinations are not allowed", ErrUnsafeURL)
	}

	if ip := net.ParseIP(host); ip != nil && isPrivateIP(ip) {
		return fmt.Errorf("%w: private IP destinations are not allowed", ErrUnsafeURL)
	}

	if s.isBlockedDomain(host) {
		return fmt.Errorf("%w: blocked domain", ErrUnsafeURL)
	}

	return nil
}

func (s *RulesScanner) isBlockedDomain(host string) bool {

	for blocked := range s.blockedDomains {
		if host == blocked || strings.HasSuffix(host, "."+blocked) {
			return true
		}
	}

	return false
}

func normalizeHost(host string) string {

	return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(host)), ".")
}

func isLocalHost(host string) bool {

	switch host {
	case "localhost":
		return true
	}

	return strings.HasSuffix(host, ".localhost")
}

func isPrivateIP(ip net.IP) bool {

	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
		"169.254.0.0/16",
		"::1/128",
		"fc00::/7",
		"fe80::/10",
	}

	for _, cidr := range privateRanges {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}

		if network.Contains(ip) {
			return true
		}
	}

	return false
}
