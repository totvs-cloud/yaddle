package yaddle

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/totvs-cloud/yaddle/config"
)

// AuthGetToken
var (

	// Response
	token        Token
	access       Access
	authResponse AuthResponse

	// Request
	passwordCredentials PasswordCredentials
	auth                Auth
	authOpenStack       AuthOpenStack
)

// GetHosts
var (
	hypervisorGH  Hypervisor
	hostsResponse OpenStackHosts
)

// GetServers
var (
	hypervisorGS    Hypervisor
	serversResponse OpenStackHosts
)

// ListServersFromHosts
var (
	hypervisorLSFH                Hypervisor
	listServersFromOpenStackHosts OpenStackHosts
)

// Global
var (
	server    Server
	authToken string
)

// Global
func init() {
	server = Server{
		UUID: "a67d8b68-47bb-49dd-88ad-8cf9844e62cd",
		Name: "instance-00003068",
	}

	authToken = "ABC"
}

//AuthGetToken
func init() {

	t, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")

	//AuthGetToken Response
	token = Token{
		IssuedAt: "2018-02-14T19:42:42.848806",
		Expires:  t,
		ID:       authToken,
	}

	access = Access{
		Token: token,
	}

	authResponse = AuthResponse{
		Access: access,
	}

	//AuthGetToken Request
	passwordCredentials = PasswordCredentials{
		Username: "jc123",
		Password: "jc123",
	}

	auth = Auth{
		PasswordCredentials: passwordCredentials,
		TenantID:            "8662e6ce659946be9213346d3deaf013",
	}

	authOpenStack = AuthOpenStack{
		Auth: auth,
	}

}

// GetHosts
func init() {

	hypervisorGH = Hypervisor{
		Status:             "enabled",
		State:              "down",
		ID:                 12,
		HypervisorHostname: "compute-2.dev.nuvem-intera.local",
	}

	hostsResponse = OpenStackHosts{
		Hypervisors: []Hypervisor{hypervisorGH},
	}

}

// GetServers
func init() {

	hypervisorGS = Hypervisor{
		Status:             "enabled",
		State:              "down",
		ID:                 12,
		HypervisorHostname: "compute-2.dev.nuvem-intera.local",
		Servers:            []Server{server},
	}

	serversResponse = OpenStackHosts{
		Hypervisors: []Hypervisor{hypervisorGS},
	}

}

// ListServersFromHosts
func init() {

	hypervisorLSFH = hypervisorGS
	hypervisorLSFH.HypervisorHostname = "compute-1.dev.nuvem-intera.local"

	listServersFromOpenStackHosts = OpenStackHosts{
		Hypervisors: []Hypervisor{hypervisorLSFH},
	}

}

func MockingServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		switch r.URL.Path {

		case "/v2.0/tokens":
			var reqToken AuthOpenStack
			defer r.Body.Close()
			json.NewDecoder(r.Body).Decode(&reqToken)

			if reqToken.Auth.PasswordCredentials.Password == config.OpenStack.Password &&
				reqToken.Auth.PasswordCredentials.Username == config.OpenStack.Username &&
				reqToken.Auth.TenantID == config.OpenStack.TenantID {

				resp, _ := json.Marshal(authResponse)
				fmt.Fprintln(w, string(resp))
			}

		case "/v2/" + config.OpenStack.TenantID + "/os-hypervisors":
			if r.Header["X-Auth-Token"][0] == authToken {
				resp, _ := json.Marshal(hostsResponse)
				fmt.Fprintln(w, string(resp))
			}

		case "/v2/" + config.OpenStack.TenantID + "/os-hypervisors/compute-2.dev.nuvem-intera.local/servers":
			if r.Header["X-Auth-Token"][0] == authToken {
				resp, _ := json.Marshal(serversResponse)
				fmt.Fprintln(w, string(resp))
			}

		case "/v2/" + config.OpenStack.TenantID + "/os-hypervisors/compute-1.dev.nuvem-intera.local/servers":
			if r.Header["X-Auth-Token"][0] == authToken {
				resp, _ := json.Marshal(listServersFromOpenStackHosts)
				fmt.Fprintln(w, string(resp))
			}

		}
	}))
}

func Test_SetConfigs_WithValidConfigOpenStack(t *testing.T) {
	configMock := config.OpenStackConfig{
		BaseUrl:    "http://minhaurl.com:6000",
		AuthUrl:    "http://minhaauthurl.com:7000",
		Username:   "jc321",
		Password:   "jc321",
		TenantName: "testeTenant",
		TenantID:   "8662e6ce659946be9213346d3deaf015",
	}

	SetConfigs(configMock)
	jConfig, _ := json.Marshal(configMock)
	jOpenStackConfig, _ := json.Marshal(config.OpenStack)
	if string(jConfig) != string(jOpenStackConfig) {
		t.Errorf(" Config Mock: %s != Config after SetConfigs: %s", jConfig, jOpenStackConfig)
	}
}

func Test_AuthGetToken_WithValidConfig_ReturnsValidAuthToken(t *testing.T) {
	res, _ := json.Marshal(token)

	// TODO: Ver se não é melhor usar a propriedade RequestURI
	httpMockingServer := MockingServer()
	config.OpenStack.AuthUrl = httpMockingServer.URL

	aToken, _ := AuthGetToken()
	authToken, _ := json.Marshal(aToken)

	if string(res) != string(authToken) {
		t.Errorf(" Response Mock: %s != Response AuthGetToken: %s", res, authToken)
	}
}

func Test_GetHosts_WithValidConfigTenantIDAndToken_ReturnsValidOpenStackHosts(t *testing.T) {

	res, _ := json.Marshal(hostsResponse)

	// TODO: Ver se não é melhor usar a propriedade RequestURI
	httpMockingServer := MockingServer()
	config.OpenStack.BaseUrl = httpMockingServer.URL

	hostsResp, _ := GetHosts(authToken)
	hosts, _ := json.Marshal(hostsResp)

	if string(res) != string(hosts) {
		t.Errorf(" Response Mock: %s != Response GetHosts: %s", res, hosts)
	}
}

func Test_GetServers_WithValidHostAndToken_ReturnsValidOpenStackHosts(t *testing.T) {

	res, _ := json.Marshal(serversResponse)

	// TODO: Ver se não é melhor usar a propriedade RequestURI
	httpMockingServer := MockingServer()
	config.OpenStack.BaseUrl = httpMockingServer.URL

	servers, _ := GetServers("compute-2.dev.nuvem-intera.local", authToken)
	getServers, _ := json.Marshal(servers)

	if string(res) != string(getServers) {
		t.Errorf(" Response Mock: %s != Response GetServers: %s", res, getServers)
	}
}

func Test_ListServersFromHosts_WithValidHostListAndToken_ReturnsValidHostsWithOpenStackHosts(t *testing.T) {

	listServersFromHostsMockResponse := listServersFromOpenStackHosts
	listServersFromHostsMockResponse = OpenStackHosts{
		Hypervisors: []Hypervisor{hypervisorGS, hypervisorLSFH},
	}

	res, _ := json.Marshal(listServersFromHostsMockResponse)

	// TODO: Ver se não é melhor usar a propriedade RequestURI
	httpMockingServer := MockingServer()
	config.OpenStack.BaseUrl = httpMockingServer.URL

	var OSHosts OpenStackHosts

	hypervisor1 := hypervisorGS
	hypervisor1.Servers = nil

	hypervisor2 := hypervisorLSFH
	hypervisor2.Servers = nil

	OSHosts = OpenStackHosts{
		Hypervisors: []Hypervisor{hypervisor1, hypervisor2},
	}

	serversFromHosts, _ := ListServersFromHosts(OSHosts.Hypervisors, authToken)
	listServersFromHosts, _ := json.Marshal(serversFromHosts)

	if string(res) != string(listServersFromHosts) {
		t.Errorf(" Response Mock: %s != Response ListServersFromHosts: %s", res, listServersFromHosts)
	}
}

func Test_GetAllHostsFullInfo_WithValidConfig_ReturnsValidHostsFullInfoResponse(t *testing.T) {

	listServersFromHostsMockResponse := listServersFromOpenStackHosts
	listServersFromHostsMockResponse = OpenStackHosts{
		Hypervisors: []Hypervisor{hypervisorGS},
	}

	res, _ := json.Marshal(listServersFromHostsMockResponse)

	// TODO: Ver se não é melhor usar a propriedade RequestURI
	httpMockingServer := MockingServer()
	config.OpenStack.BaseUrl = httpMockingServer.URL
	config.OpenStack.AuthUrl = httpMockingServer.URL

	serversFromHosts, err := GetAllHostsFullInfo()
	if err != nil {
		log.Println(err.Error())
	}
	getAllHostsFullInfo, _ := json.Marshal(serversFromHosts)

	if string(res) != string(getAllHostsFullInfo) {
		t.Errorf(" Response Mock: %s != Response GetAllHostsFullInfo: %s", res, getAllHostsFullInfo)
	}
}
