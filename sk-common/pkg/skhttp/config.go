package skhttp

type Config struct {
	Url                string `yaml:"url"`
	RootCaPath         string `yaml:"rootCaPath"` // Path to a trusted root CA file
	RootCaData         string `yaml:"rootCaData"` // Base64 encoded PEM data containing root CA
	InsecureSkipVerify bool   `yaml:"insecureSkipVerify"`
	ClientAuth         struct {
		Id     string `yaml:"id"`
		Secret string `yaml:"secret"`
	} `yaml:"clientAuth"`
}
