package yaddle

import (
	"crypto/tls"
	"encoding/json"
	"net/http"

	"gitlab-devops.totvs.com.br/golang/yaddle/config"
)

// Server is a VM into Host
type Server struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
}

// Hypervisor is a Compute/Host
type Hypervisor struct {
	Status             string   `json:"status"`
	State              string   `json:"state"`
	ID                 int      `json:"id"`
	HypervisorHostname string   `json:"hypervisor_hostname"`
	Servers            []Server `json:"servers"`
}

//ServersResponse is a Response of OpenStack Nova API
type ServersResponse struct {
	Hypervisors []Hypervisor `json:"hypervisors"`
}

// GetServers is http request  from OpenStack Nova API
func GetServers(compute string, authToken string) (*ServersResponse, error) {
	var response ServersResponse

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	reqUrl := config.OpenStack.BaseUrl + "/v2/8662e6ce659946be9213336d3deaf012/os-hypervisors/compute-2.dev.nuvem-intera.local/servers"

	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", "python-novaclient")
	req.Header.Add("Accept", "application/json")

	req.Header.Add("X-Auth-Token", authToken)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&response)

	return &response, nil
}
