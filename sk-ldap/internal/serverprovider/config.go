package serverprovider

// NB: These values are strongly inspired from dex configuration (https://github.com/dexidp/dex)

type Config struct {

	// The host and port of the LDAP server.
	// If port isn't supplied, it will be guessed based on the TLS configuration. 389 or 636.
	Host string `yaml:"host"`
	Port string `yaml:"port"`

	// Timeout on connection to ldap server. Default to 10
	TimeoutSec int `yaml:"timeoutSec"`

	// Required if LDAP host does not use TLS.
	InsecureNoSSL bool `yaml:"insecureNoSSL"`

	// Don't verify the CA.
	InsecureSkipVerify bool `yaml:"insecureSkipVerify"`

	// Connect to the insecure port then issue a StartTLS command to negotiate a
	// secure connection. If unsupplied secure connections will use the LDAPS
	// protocol.
	StartTLS bool `yaml:"startTLS"`

	// Path to a trusted root certificate file.
	RootCA string `yaml:"rootCA"`
	// Base64 encoded PEM data containing root CAs.
	RootCAData string `yaml:"rootCAData"`
	// Path to a client cert file
	ClientCert string `yaml:"clientCert"`
	// Path to a client private key file
	ClientKey string `yaml:"clientKey"`

	// BindDN and BindPW for an application service account. The connector uses these
	// credentials to search for users and groups.
	BindDN string `yaml:"bindDN"`
	BindPW string `yaml:"bindPW"`

	UserSearch struct {
		// BaseDN to start the search from. For example "cn=users,dc=example,dc=com"
		BaseDN string `yaml:"baseDN"`

		// Optional filter to apply when searching the directory. For example "(objectClass=person)"
		Filter string `yaml:"filter"`

		// Attribute to match against the login. This will be translated and combined
		// with the other filter as "(<loginAttr>=<login>)".
		LoginAttr string `yaml:"loginAttr"`

		// Can either be:
		// * "sub" - search the whole sub tree
		// * "one" - only search one level
		Scope string `yaml:"scope"`

		// The attribute providing the numerical user ID
		NumericalIdAttr string `yaml:"numericalIdAttr"`

		// The attribute providing the user's email
		EmailAttr string `yaml:"emailAttr"`

		// The attribute providing the user's common name
		CnAttr string `yaml:"cnAttr"`
	} `yaml:"userSearch"`

	// Group search configuration.
	GroupSearch struct {
		// BaseDN to start the search from. For example "cn=groups,dc=example,dc=com"
		BaseDN string `yaml:"baseDN"`

		// Optional filter to apply when searching the directory. For example "(objectClass=posixGroup)"
		Filter string `yaml:"filter"`

		Scope string `yaml:"scope"` // Defaults to "sub"

		// The attribute of the group that represents its name.
		NameAttr string `yaml:"nameAttr"`

		// The filter for group/user relationship will be: (<linkGroupAttr>=<Value of LinkUserAttr for the user>)
		// If there is several value for LinkUserAttr, we will loop on.
		LinkUserAttr  string `yaml:"linkUserAttr"`
		LinkGroupAttr string `yaml:"linkGroupAttr"`
	} `yaml:"groupSearch"`
}
