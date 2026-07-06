package geo

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"
)

// Locator resolves a client IP address to a 2-letter ISO 3166-1 country code.
type Locator interface {
	CountryCode(ip string) string
}

// NoopLocator always returns "" — disables geo routing without any network calls.
// Used in tests and as a safe default.
type NoopLocator struct{}

func (NoopLocator) CountryCode(string) string { return "" }

// IPAPILocator uses the free ip-api.com endpoint to resolve country codes.
// Free tier: 45 requests/minute, no API key required.
// All errors are suppressed and return "" so geo routing degrades gracefully.
type IPAPILocator struct {
	client *http.Client
}

func NewIPAPILocator() *IPAPILocator {
	return &IPAPILocator{
		client: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
}

func (l *IPAPILocator) CountryCode(ip string) string {

	if ip == "" || isLocalOrPrivate(ip) {
		return ""
	}

	resp, err := l.client.Get(
		fmt.Sprintf("http://ip-api.com/json/%s?fields=countryCode", ip),
	)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	var result struct {
		CountryCode string `json:"countryCode"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ""
	}

	return result.CountryCode
}

// isLocalOrPrivate returns true for loopback and RFC-1918 addresses so we
// never make a lookup call for development traffic.
func isLocalOrPrivate(addr string) bool {

	ip := net.ParseIP(addr)
	if ip == nil {
		return addr == "localhost"
	}

	privateRanges := []string{
		"127.0.0.0/8",
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
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
