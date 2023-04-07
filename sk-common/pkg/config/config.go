package config

var yes = true
var no = false

type ServiceClient struct {
	Id     string `yaml:"id"`
	Secret string `yaml:"secret"`
}

type ServiceConfig struct {
	Disabled bool            `yaml:"disabled"`
	Clients  []ServiceClient `yaml:"clients"`
}

type SkServerConfig struct {
	Interface string `yaml:"interface"`
	Port      int    `yaml:"port"`
	Ssl       *bool  `yaml:"ssl"`
	CertDir   string `yaml:"certDir"`  // CertDir is the directory that contains the server key and certificate.
	CertName  string `yaml:"certName"` // CertName is the server certificate name. Defaults to tls.crt.
	KeyName   string `yaml:"keyName"`  // KeyName is the server key name. Defaults to tls.key.
}

// NB: The ClientManager must accept all clients if bound on "127.0.0.1". Otherwise, allowed clients must be defined. See main.go

func (srv *SkServerConfig) Default(port int) {
	if srv.Interface == "" {
		srv.Interface = "127.0.0.1"
	}
	if srv.Port == 0 {
		srv.Port = port
	}
	if srv.Ssl == nil {
		if srv.Interface != "127.0.0.1" {
			srv.Ssl = &yes
		} else {
			srv.Ssl = &no
		}
	}
}
