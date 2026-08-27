package ddns

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestCloudflareCredential_UpsertRecord_Create(t *testing.T) {
	var created map[string]interface{}
	cre := newCloudflareTestClient("zone123", func(r *http.Request) *http.Response {
		switch r.Method + " " + r.URL.Path {
		case "GET /zones/zone123/dns_records":
			return cloudflareTestResponse(http.StatusOK, `{"success":true,"result":[]}`)
		case "POST /zones/zone123/dns_records":
			if err := json.NewDecoder(r.Body).Decode(&created); err != nil {
				t.Fatalf("decode create payload: %v", err)
			}
			return cloudflareTestResponse(http.StatusOK, `{"success":true,"result":{"id":"record-new","content":"1.2.3.4"}}`)
		}
		return cloudflareTestResponse(http.StatusNotFound, "{}")
	})
	if err := cre.UpsertRecord("test.example.com", "1.2.3.4"); err != nil {
		t.Fatalf("UpsertRecord() error = %v", err)
	}

	if created["name"] != "test.example.com" || created["content"] != "1.2.3.4" {
		t.Fatalf("unexpected create payload: %#v", created)
	}
	if created["proxied"] != false || created["ttl"] != float64(cloudflareDefaultTTL) {
		t.Fatalf("unexpected create proxy/ttl: %#v", created)
	}
}

func TestCloudflareCredential_UpsertRecord_Update(t *testing.T) {
	var updated map[string]string
	cre := newCloudflareTestClient("zone123", func(r *http.Request) *http.Response {
		switch r.Method + " " + r.URL.Path {
		case "GET /zones/zone123/dns_records":
			return cloudflareTestResponse(http.StatusOK, `{"success":true,"result":[{"id":"record123","name":"test.example.com","type":"A","content":"5.6.7.8"}]}`)
		case "PATCH /zones/zone123/dns_records/record123":
			if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
				t.Fatalf("decode update payload: %v", err)
			}
			return cloudflareTestResponse(http.StatusOK, `{"success":true,"result":{"id":"record123","content":"1.2.3.4"}}`)
		}
		return cloudflareTestResponse(http.StatusNotFound, "{}")
	})
	if err := cre.UpsertRecord("test.example.com", "1.2.3.4"); err != nil {
		t.Fatalf("UpsertRecord() error = %v", err)
	}
	if updated["content"] != "1.2.3.4" {
		t.Fatalf("unexpected update payload: %#v", updated)
	}
}

func TestCloudflareCredential_UpsertRecord_FindZone(t *testing.T) {
	cre := newCloudflareTestClient("", func(r *http.Request) *http.Response {
		if r.Method == http.MethodGet && r.URL.Path == "/zones" {
			if r.URL.Query().Get("name") == "test.example.com" {
				return cloudflareTestResponse(http.StatusOK, `{"success":true,"result":[]}`)
			}
			if r.URL.Query().Get("name") == "example.com" {
				return cloudflareTestResponse(http.StatusOK, `{"success":true,"result":[{"id":"zone123","name":"example.com"}]}`)
			}
		}
		if r.Method == http.MethodGet && r.URL.Path == "/zones/zone123/dns_records" {
			return cloudflareTestResponse(http.StatusOK, `{"success":true,"result":[]}`)
		}
		if r.Method == http.MethodPost && r.URL.Path == "/zones/zone123/dns_records" {
			return cloudflareTestResponse(http.StatusOK, `{"success":true,"result":{"id":"record-new"}}`)
		}
		return cloudflareTestResponse(http.StatusNotFound, "{}")
	})
	if err := cre.UpsertRecord("test.example.com", "1.2.3.4"); err != nil {
		t.Fatalf("UpsertRecord() error = %v", err)
	}
	if cre.zoneID != "zone123" {
		t.Fatalf("zone id = %q, want %q", cre.zoneID, "zone123")
	}
}

func TestCloudflareCredential_UpsertRecord_Unauthorized(t *testing.T) {
	cre := newCloudflareTestClient("zone123", func(r *http.Request) *http.Response {
		return cloudflareTestResponse(http.StatusUnauthorized, `{"success":false,"errors":[{"code":9109,"message":"Invalid access token"}]}`)
	})
	err := cre.UpsertRecord("test.example.com", "1.2.3.4")
	if err == nil {
		t.Fatal("UpsertRecord() expected error")
	}
	if !strings.Contains(err.Error(), "401") || !strings.Contains(err.Error(), "Invalid access token") {
		t.Fatalf("unexpected error: %v", err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func newCloudflareTestClient(zoneID string, handler func(*http.Request) *http.Response) *CloudflareCredential {
	cre := newCloudflareCredential("token", zoneID, "https://cloudflare.test")
	cre.client = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return handler(req), nil
		}),
	}
	return cre
}

func cloudflareTestResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Status:     fmt.Sprintf("%d %s", status, http.StatusText(status)),
		Body:       ioutil.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}
