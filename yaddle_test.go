package yaddle

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gitlab-devops.totvs.com.br/golang/yaddle/config"
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
	hostsResponse HostsResponse
)

// GetServers
var (
	server          Server
	hypervisorGS    Hypervisor
	serversResponse ServersResponse
)

// Global
var (
	authToken string
)

//AuthGetToken
func init() {

	t, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")

	//AuthGetToken Response
	token = Token{
		IssuedAt: "2018-02-14T19:42:42.848806",
		Expires:  t,
		ID:       "XYZ",
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

	hostsResponse = HostsResponse{
		Hypervisors: []Hypervisor{hypervisorGH},
	}

}

// GetServers
func init() {

	server = Server{
		UUID: "a67d8b68-47bb-49dd-88ad-8cf9844e62cd",
		Name: "instance-00003068",
	}

	hypervisorGS = Hypervisor{
		Status:             "enabled",
		State:              "down",
		ID:                 12,
		HypervisorHostname: "compute-2.dev.nuvem-intera.local",
		Servers:            []Server{server},
	}

	serversResponse = ServersResponse{
		Hypervisors: []Hypervisor{hypervisorGS},
	}

}

// Global
func init() {
	authToken = "ABC"

}

// Configs
func init() {

	config.OpenStack.Username = passwordCredentials.Username
	config.OpenStack.Password = passwordCredentials.Password

	config.OpenStack.TenantID = auth.TenantID

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

		}
	}))
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

func Test_GetHosts_WithValidConfigTenantIDAndToken_ReturnsValidServersResponse(t *testing.T) {

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

func Test_GetServers_WithValidHostAndToken_ReturnsValidServersResponse(t *testing.T) {

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
