package config

type ServiceClient struct {
	Id     string `yaml:"id"`
	Secret string `yaml:"secret"`
}

type ServiceConfig struct {
	Disabled bool            `yaml:"disabled"`
	Clients  []ServiceClient `yaml:"clients"`
}
