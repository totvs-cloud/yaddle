package config

//A OpenStackConfig represents a parsed OpenStack Config
type OpenStackConfig struct {
	BaseUrl    string `json:"baseUrl"`
	AuthUrl    string `json:"authUrl"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	TenantName string `json:"tenantName"`
	TenantID   string `json:"tenantID"`
}

var (
	//OpenStack global reference OpenStack Config
	OpenStack OpenStackConfig
)
