package security

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const googleSafeBrowsingEndpoint = "https://safebrowsing.googleapis.com/v4/threatMatches:find"

type GoogleSafeBrowsingScanner struct {
	APIKey     string
	HTTPClient *http.Client
	Endpoint   string
}

func NewGoogleSafeBrowsingScanner(apiKey string) URLScanner {

	return &GoogleSafeBrowsingScanner{
		APIKey: apiKey,
		HTTPClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		Endpoint: googleSafeBrowsingEndpoint,
	}
}

func (s *GoogleSafeBrowsingScanner) Check(rawURL string) error {

	if s.APIKey == "" {
		return nil
	}

	body := map[string]interface{}{
		"client": map[string]string{
			"clientId":      "url-shortener",
			"clientVersion": "1.0.0",
		},
		"threatInfo": map[string]interface{}{
			"threatTypes":      []string{"MALWARE", "SOCIAL_ENGINEERING", "UNWANTED_SOFTWARE"},
			"platformTypes":    []string{"ANY_PLATFORM"},
			"threatEntryTypes": []string{"URL"},
			"threatEntries": []map[string]string{
				{"url": rawURL},
			},
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal safe browsing request: %w", err)
	}

	endpoint := s.Endpoint
	if endpoint == "" {
		endpoint = googleSafeBrowsingEndpoint
	}

	req, err := http.NewRequest(
		http.MethodPost,
		endpoint+"?key="+s.APIKey,
		bytes.NewReader(payload),
	)
	if err != nil {
		return fmt.Errorf("build safe browsing request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := s.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("call safe browsing provider: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("safe browsing provider returned status %d", resp.StatusCode)
	}

	var result struct {
		Matches []struct {
			ThreatType string `json:"threatType"`
		} `json:"matches"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("decode safe browsing response: %w", err)
	}

	if len(result.Matches) > 0 {
		return fmt.Errorf("%w: flagged by safe browsing provider", ErrUnsafeURL)
	}

	return nil
}
