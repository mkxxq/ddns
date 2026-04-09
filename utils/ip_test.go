package utils

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func TestJSONIPClient_GetOuterIP(t *testing.T) {
	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader(`{"ip":"1.2.3.4"}`)),
				Header:     make(http.Header),
			}, nil
		}),
	}

	cli := NewJSONIPClient("https://jsonip.test", client)
	got, err := cli.GetOuterIP()
	if err != nil {
		t.Fatal(err)
	}
	if got != "1.2.3.4" {
		t.Fatalf("got %s", got)
	}
}
