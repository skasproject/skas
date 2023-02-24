package skhttp

// json mapping added as this is embedded in dex skas connector

type Config struct {
	Url                string `yaml:"url" json:"url"`
	RootCaPath         string `yaml:"rootCaPath" json:"rootCaPath"` // Path to a trusted root CA file
	RootCaData         string `yaml:"rootCaData" json:"rootCaData"` // Base64 encoded PEM data containing root CA
	InsecureSkipVerify bool   `yaml:"insecureSkipVerify" json:"insecureSkipVerify"`
	ClientAuth         struct {
		Id     string `yaml:"id" json:"id"`
		Secret string `yaml:"secret" json:"secret"`
	} `yaml:"clientAuth" json:"clientAuth"`
}
