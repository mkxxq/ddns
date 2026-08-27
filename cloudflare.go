package ddns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const (
	cloudflareAPIBaseURL = "https://api.cloudflare.com/client/v4"
	cloudflareDefaultTTL = 60
)

type CloudflareCredential struct {
	token   string
	zoneID  string
	client  *http.Client
	baseURL string
}

type cloudflareZone struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type cloudflareDNSRecord struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Content string `json:"content"`
}

type cloudflareAPIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type cloudflareResponse struct {
	Success bool                 `json:"success"`
	Errors  []cloudflareAPIError `json:"errors"`
	Result  json.RawMessage      `json:"result"`
}

func NewCloudflareCredential() *CloudflareCredential {
	cre := &CloudflareCredential{
		token:   os.Getenv("CLOUDFLARE_API_TOKEN"),
		zoneID:  os.Getenv("CLOUDFLARE_ZONE_ID"),
		client:  &http.Client{},
		baseURL: cloudflareAPIBaseURL,
	}
	if cre.token == "" {
		log.Panicln("cloudflare provider require CLOUDFLARE_API_TOKEN")
	}
	return cre
}

func newCloudflareCredential(token string, zoneID string, baseURL string) *CloudflareCredential {
	if baseURL == "" {
		baseURL = cloudflareAPIBaseURL
	}
	return &CloudflareCredential{
		token:   token,
		zoneID:  zoneID,
		client:  &http.Client{},
		baseURL: baseURL,
	}
}

func (cre *CloudflareCredential) UpsertRecord(subDomain string, ip string) error {
	name := strings.TrimSuffix(strings.TrimSpace(subDomain), ".")
	if name == "" {
		return fmt.Errorf("error subDomain: %s", subDomain)
	}

	zoneID, err := cre.resolveZoneID(name)
	if err != nil {
		return err
	}

	records, err := cre.listDNSRecords(zoneID, name)
	if err != nil {
		return err
	}

	if len(records) == 0 {
		return cre.createDNSRecord(zoneID, name, ip)
	}
	return cre.updateDNSRecord(zoneID, records[0].ID, ip)
}

func (cre *CloudflareCredential) resolveZoneID(domain string) (string, error) {
	if cre.zoneID != "" {
		return cre.zoneID, nil
	}

	originalDomain := domain
	for candidate := domain; candidate != ""; {
		zones, err := cre.listZones(candidate)
		if err != nil {
			return "", err
		}
		if len(zones) > 0 {
			cre.zoneID = zones[0].ID
			return cre.zoneID, nil
		}

		idx := strings.Index(candidate, ".")
		if idx < 0 {
			break
		}
		candidate = candidate[idx+1:]
	}
	return "", fmt.Errorf("cloudflare zone not found for %s", originalDomain)
}

func (cre *CloudflareCredential) listZones(name string) ([]cloudflareZone, error) {
	query := url.Values{}
	query.Set("name", name)
	query.Set("status", "active")
	query.Set("per_page", "1")

	var zones []cloudflareZone
	err := cre.request(http.MethodGet, "/zones?"+query.Encode(), nil, &zones)
	return zones, err
}

func (cre *CloudflareCredential) listDNSRecords(zoneID string, name string) ([]cloudflareDNSRecord, error) {
	query := url.Values{}
	query.Set("type", "A")
	query.Set("name", name)
	query.Set("per_page", "100")

	var records []cloudflareDNSRecord
	err := cre.request(http.MethodGet, "/zones/"+zoneID+"/dns_records?"+query.Encode(), nil, &records)
	return records, err
}

func (cre *CloudflareCredential) createDNSRecord(zoneID string, name string, ip string) error {
	payload := map[string]interface{}{
		"type":    "A",
		"name":    name,
		"content": ip,
		"ttl":     cloudflareDefaultTTL,
		"proxied": false,
	}
	var record cloudflareDNSRecord
	return cre.request(http.MethodPost, "/zones/"+zoneID+"/dns_records", payload, &record)
}

func (cre *CloudflareCredential) updateDNSRecord(zoneID string, recordID string, ip string) error {
	payload := map[string]string{
		"content": ip,
	}
	var record cloudflareDNSRecord
	return cre.request(http.MethodPatch, "/zones/"+zoneID+"/dns_records/"+recordID, payload, &record)
}

func (cre *CloudflareCredential) request(method string, path string, payload interface{}, result interface{}) error {
	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, cre.baseURL+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+cre.token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := cre.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("cloudflare api error: status=%s body=%s", resp.Status, strings.TrimSpace(string(data)))
	}

	var envelope cloudflareResponse
	if err := json.Unmarshal(data, &envelope); err != nil {
		return fmt.Errorf("decode cloudflare response: %w", err)
	}
	if !envelope.Success {
		return cloudflareAPIErrors(envelope.Errors)
	}
	if result != nil && len(envelope.Result) > 0 {
		if err := json.Unmarshal(envelope.Result, result); err != nil {
			return fmt.Errorf("decode cloudflare result: %w", err)
		}
	}
	return nil
}

func cloudflareAPIErrors(errs []cloudflareAPIError) error {
	if len(errs) == 0 {
		return fmt.Errorf("cloudflare api error")
	}
	msgs := make([]string, 0, len(errs))
	for _, err := range errs {
		msgs = append(msgs, fmt.Sprintf("%d: %s", err.Code, err.Message))
	}
	return fmt.Errorf("cloudflare api error: %s", strings.Join(msgs, "; "))
}
