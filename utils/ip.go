package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	GET_IP_URL = "https://jsonip.com"
)

type OuterIPGetter interface {
	GetOuterIP() (string, error)
}

type getOuterIpOutput struct {
	IP string `json:"ip"`
}

type JSONIPClient struct {
	URL    string
	Client *http.Client
}

func NewJSONIPClient(url string, client *http.Client) *JSONIPClient {
	if url == "" {
		url = GET_IP_URL
	}
	if client == nil {
		client = &http.Client{}
	}
	return &JSONIPClient{URL: url, Client: client}
}

func (cli *JSONIPClient) GetOuterIP() (string, error) {
	req, err := http.NewRequest(http.MethodGet, cli.URL, nil)
	if err != nil {
		return "", err
	}
	res, err := cli.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	if res.StatusCode/100 == 2 {
		out := new(getOuterIpOutput)
		err = json.Unmarshal(body, out)
		if err != nil {
			return "", err
		}
		return out.IP, nil
	} else {
		return "", fmt.Errorf("status: %d, response: %s", res.StatusCode, string(body))
	}
}

func GetOuterIp() (string, error) {
	return NewJSONIPClient("", nil).GetOuterIP()
}
