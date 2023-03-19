package config

type SkServerConfig struct {
	BindAddr string `yaml:"bindAddr"`
	Ssl      bool   `yaml:"ssl"`
	CertDir  string `yaml:"certDir"`  // CertDir is the directory that contains the server key and certificate.
	CertName string `yaml:"certName"` // CertName is the server certificate name. Defaults to tls.crt.
	KeyName  string `yaml:"keyName"`  // KeyName is the server key name. Defaults to tls.key.
}
