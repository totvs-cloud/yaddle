package yaddle

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab-devops.totvs.com.br/golang/yaddle/config"
)

func MockingServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		switch r.URL.Path {
		case "/v2/8662e6ce659946be9213336d3deaf012/os-hypervisors/compute-2.dev.nuvem-intera.local/servers":
			// log.Printf("%s", r.Header)
			if r.Header["X-Auth-Token"][0] == "ABC" {
				server := Server{
					UUID: "a67d8b68-47bb-49dd-88ad-8cf9844e62cd",
					Name: "instance-00003068",
				}

				hypervisor := Hypervisor{
					Status:             "enabled",
					State:              "down",
					ID:                 12,
					HypervisorHostname: "compute-2.dev.nuvem-intera.local",
					Servers:            []Server{server},
				}

				serversResponse := ServersResponse{
					Hypervisors: []Hypervisor{hypervisor},
				}
				resp, _ := json.Marshal(serversResponse)
				fmt.Fprintln(w, string(resp))
			}
		case "/v2.0/tokens":
			fmt.Fprintln(w, "ABC")
		}
	}))
}

func Test_GetServers_WithValidHostAndToken_ReturnsValidServersResponse(t *testing.T) {

	server := Server{
		UUID: "a67d8b68-47bb-49dd-88ad-8cf9844e62cd",
		Name: "instance-00003068",
	}

	hypervisor := Hypervisor{
		Status:             "enabled",
		State:              "down",
		ID:                 12,
		HypervisorHostname: "compute-2.dev.nuvem-intera.local",
		Servers:            []Server{server},
	}

	serversResponse := ServersResponse{
		Hypervisors: []Hypervisor{hypervisor},
	}
	res, _ := json.Marshal(serversResponse)

	token := "ABC"

	httpMockingServer := MockingServer()
	config.OpenStack.BaseUrl = httpMockingServer.URL

	servers, _ := GetServers("compute-2.dev.nuvem-intera.local", token)
	getServers, _ := json.Marshal(servers)

	if string(res) != string(getServers) {
		t.Errorf(" Response Mock: %s != Response GetServers: %s", res, getServers)
	}
}
