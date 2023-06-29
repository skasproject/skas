package skclient

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

// Config can also be saved in environment variables

const SK_CLIENT_URL = "SK_CLIENT_URL"
const SK_CLIENT_ROOT_CA_DATA = "SK_CLIENT_ROOT_CA_DATA"
const SK_CLIENT_INSECURE_SKIP_VERIFY = "SK_CLIENT_INSECURE_SKIP_VERIFY"
const SK_CLIENT_AUTH_ID = "SK_CLIENT_AUTH_ID"
const SK_CLIENT_AUTH_SECRET = "SK_CLIENT_AUTH_SECRET"
