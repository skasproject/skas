package clientauth

// This is the structure to define a client in our configuration. Used only on server side

type Config struct {
	Id     string `yaml:"id"`
	Secret string `yaml:"secret"`
}
