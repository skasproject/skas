package identitygetter

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"github.com/go-logr/logr"
	"gopkg.in/ldap.v2"
	"os"
	"path/filepath"
	commonHandlers "skas/sk-common/pkg/skserver/handlers"
	"time"
)

func New(ldapConfig *Config, baseLog logr.Logger, configFolder string) (commonHandlers.IdentityGetter, error) {

	ldapGetter := ldapIdentityGetter{
		Config: ldapConfig,
	}
	if ldapGetter.Host == "" {
		return &ldapGetter, fmt.Errorf("missing required ldap.host")
	}
	if ldapGetter.UserSearch.BaseDN == "" {
		return &ldapGetter, fmt.Errorf("missing required ldap.userSearch.baseDN")
	}
	if ldapGetter.UserSearch.LoginAttr == "" {
		return &ldapGetter, fmt.Errorf("missing required ldap.userSearch.loginAttr")
	}
	if ldapGetter.GroupSearch.BaseDN != "" {
		if ldapGetter.GroupSearch.BaseDN == "" {
			return &ldapGetter, fmt.Errorf("missing required ldap.groupSearch.baseDN")
		}
		if ldapGetter.GroupSearch.NameAttr == "" {
			return &ldapGetter, fmt.Errorf("missing required ldap.groupSearch.nameAttr")
		}
		if ldapGetter.GroupSearch.LinkGroupAttr == "" {
			return &ldapGetter, fmt.Errorf("missing required ldap.groupSearch.linkGroupAttr")
		}
		if ldapGetter.GroupSearch.LinkUserAttr == "" {
			return &ldapGetter, fmt.Errorf("missing required ldap.groupSearch.linkUserAttr")
		}
	}
	ldapGetter.logger = baseLog.WithName("ldap")
	// Setup default value
	if ldapGetter.Port == "" {
		if ldapGetter.InsecureNoSSL {
			ldapGetter.Port = "389"
		} else {
			ldapGetter.Port = "636"
		}
	}
	ldapGetter.hostPort = fmt.Sprintf("%s:%s", ldapGetter.Host, ldapGetter.Port)
	if ldapGetter.TimeoutSec == 0 {
		ldapGetter.TimeoutSec = 10
	}
	// WARNING: This is a global variable
	ldap.DefaultTimeout = time.Duration(ldapGetter.TimeoutSec) * time.Second

	//ldapGetter.logger.V(2).Info("paths", "configFolder", configFolder, "RootCA", ldapGetter.RootCaPath, "ClientCert", ldapGetter.ClientCert, "ClientKey", ldapGetter.ClientKey)
	ldapGetter.RootCaPath = adjustPath(configFolder, ldapGetter.RootCaPath)
	ldapGetter.ClientCert = adjustPath(configFolder, ldapGetter.ClientCert)
	ldapGetter.ClientKey = adjustPath(configFolder, ldapGetter.ClientKey)
	//ldapGetter.logger.V(2).Info("adjusted paths", "RootCA", ldapGetter.RootCaPath, "ClientCert", ldapGetter.ClientCert, "ClientKey", ldapGetter.ClientKey)

	ldapGetter.tlsConfig = &tls.Config{ServerName: ldapGetter.Host, InsecureSkipVerify: ldapGetter.InsecureSkipVerify}
	if ldapGetter.RootCaPath != "" || len(ldapGetter.RootCaData) != 0 {
		var data []byte
		if len(ldapGetter.RootCaData) != 0 {
			data = make([]byte, base64.StdEncoding.DecodedLen(len(ldapGetter.RootCaData)))
			_, err := base64.StdEncoding.Decode(data, []byte(ldapGetter.RootCaData))
			if err != nil {
				return &ldapGetter, fmt.Errorf("error while parsing RootCaData : %w", err)
			}
		} else {
			var err error
			if data, err = os.ReadFile(ldapGetter.RootCaPath); err != nil {
				return &ldapGetter, fmt.Errorf("error while reading CA file: %w", err)
			}
		}
		rootCAs := x509.NewCertPool()
		if !rootCAs.AppendCertsFromPEM(data) {
			return &ldapGetter, fmt.Errorf("no certs found in ca file")
		}
		ldapGetter.tlsConfig.RootCAs = rootCAs
	}

	if ldapGetter.ClientKey != "" && ldapGetter.ClientCert != "" {
		cert, err := tls.LoadX509KeyPair(ldapGetter.ClientCert, ldapGetter.ClientKey)
		if err != nil {
			return &ldapGetter, fmt.Errorf("load client cert failed: %v", err)
		}
		ldapGetter.tlsConfig.Certificates = append(ldapGetter.tlsConfig.Certificates, cert)
	}
	var ok bool
	ldapGetter.userSearchScope, ok = parseScope(ldapGetter.UserSearch.Scope)
	if !ok {
		return &ldapGetter, fmt.Errorf("userSearch.Scope unknown value %q", ldapGetter.UserSearch.Scope)
	}
	ldapGetter.groupSearchScope, ok = parseScope(ldapGetter.GroupSearch.Scope)
	if !ok {
		return &ldapGetter, fmt.Errorf("groupSearch.Scope unknown value %q", ldapGetter.GroupSearch.Scope)
	}
	return &ldapGetter, nil
}

func adjustPath(baseFolder string, path string) string {
	if path != "" {
		if !filepath.IsAbs(path) {
			path = filepath.Join(baseFolder, path)
		}
		path = filepath.Clean(path)
	}
	return path
}
