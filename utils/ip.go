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

type getOuterIpOutput struct {
	IP string `json:"ip"`
}

func GetOuterIp() (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, GET_IP_URL, nil)
	if err != nil {
		return "", err
	}
	res, err := client.Do(req)
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
