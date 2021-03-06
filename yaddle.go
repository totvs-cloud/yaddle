package yaddle

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/totvs-cloud/yaddle/config"
)

// Token is a auth token for OpenStack APIs where ID is the token value
type Token struct {
	IssuedAt string    `json:"issued_at"`
	Expires  time.Time `json:"expires"`
	ID       string    `json:"id"`
}

// Access takes the Token
type Access struct {
	Token Token `json:"token"`
}

// AuthResponse is a response for OpenStack auth API where contains a token
type AuthResponse struct {
	Access Access `json:"access"`
}

// PasswordCredentials contains infor for access on OpenStack
type PasswordCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Auth is where passwordCredentials and id for tenant("master compute")
type Auth struct {
	PasswordCredentials PasswordCredentials `json:"passwordCredentials"`
	TenantID            string              `json:"tenantId"`
}

// AuthOpenStack is a payload for request on OpenStack auth API
type AuthOpenStack struct {
	Auth Auth `json:"auth"`
}

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

//OpenStackHosts is a Response of OpenStack Nova API
type OpenStackHosts struct {
	Hypervisors []Hypervisor `json:"hypervisors"`
}

// SetConfigs is function for set var config.OpenStack
func SetConfigs(configP config.OpenStackConfig) {
	config.OpenStack = configP
}

// AuthGetToken is http request from OpenStack auth API
func AuthGetToken() (*Token, error) {

	var response AuthResponse

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	reqURL := config.OpenStack.AuthUrl + "/v2.0/tokens"

	authPayload := AuthOpenStack{
		Auth: Auth{
			PasswordCredentials: PasswordCredentials{
				Username: config.OpenStack.Username,
				Password: config.OpenStack.Password,
			},
			TenantID: config.OpenStack.TenantID,
		},
	}

	authJSON, err := json.Marshal(authPayload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", reqURL, strings.NewReader(string(authJSON)))
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", "python-novaclient")
	req.Header.Add("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&response)

	return &response.Access.Token, nil

}

// TODO: Criar função base para os Gets

// GetHosts is http request from OpenStack Nova API
func GetHosts(authToken string) (*OpenStackHosts, error) {
	var response OpenStackHosts

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	reqURL := config.OpenStack.BaseUrl + "/v2/" + config.OpenStack.TenantID + "/os-hypervisors"

	req, err := http.NewRequest("GET", reqURL, nil)
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

// GetServers is http request from OpenStack Nova API
func GetServers(compute string, authToken string) (*OpenStackHosts, error) {
	var response OpenStackHosts

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	reqURL := config.OpenStack.BaseUrl + "/v2/" + config.OpenStack.TenantID + "/os-hypervisors/" + compute + "/servers"

	req, err := http.NewRequest("GET", reqURL, nil)
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

// ListServersFromHosts - give a list of hosts,  return: servers of each hosts
func ListServersFromHosts(hosts []Hypervisor, authToken string) (*OpenStackHosts, error) {
	var hyp []Hypervisor
	var hypResp OpenStackHosts

	for _, host := range hosts {
		servers, err := GetServers(host.HypervisorHostname, authToken)
		if err != nil {
			return nil, err
		}
		hyp = append(hyp, servers.Hypervisors[0])
	}
	hypResp.Hypervisors = hyp

	return &hypResp, nil
}

// GetAllHostsFullInfo is a abstraction of All functions necessary for full Hosts/Computes infos return
func GetAllHostsFullInfo() (*OpenStackHosts, error) {
	var hypResp OpenStackHosts

	token, err := AuthGetToken()
	if err != nil {
		return nil, err
	}

	hostsResp, err := GetHosts(token.ID)
	if err != nil {
		return nil, err
	}

	serversFromHosts, err := ListServersFromHosts(hostsResp.Hypervisors, token.ID)
	if err != nil {
		return nil, err
	}

	hypResp = *serversFromHosts

	return &hypResp, nil

}
