package serverprovider

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

func New(ldapConfig *Config, baseLog logr.Logger, configFolder string) (commonHandlers.IdentityServerProvider, error) {

	prvd := ldapIdentityServerProvider{
		Config: ldapConfig,
	}
	if prvd.Host == "" {
		return &prvd, fmt.Errorf("missing required ldap.host")
	}
	if prvd.UserSearch.BaseDN == "" {
		return &prvd, fmt.Errorf("missing required ldap.userSearch.baseDN")
	}
	if prvd.UserSearch.LoginAttr == "" {
		return &prvd, fmt.Errorf("missing required ldap.userSearch.loginAttr")
	}
	if prvd.GroupSearch.BaseDN != "" {
		if prvd.GroupSearch.BaseDN == "" {
			return &prvd, fmt.Errorf("missing required ldap.groupSearch.baseDN")
		}
		if prvd.GroupSearch.NameAttr == "" {
			return &prvd, fmt.Errorf("missing required ldap.groupSearch.nameAttr")
		}
		if prvd.GroupSearch.LinkGroupAttr == "" {
			return &prvd, fmt.Errorf("missing required ldap.groupSearch.linkGroupAttr")
		}
		if prvd.GroupSearch.LinkUserAttr == "" {
			return &prvd, fmt.Errorf("missing required ldap.groupSearch.linkUserAttr")
		}
	}
	prvd.logger = baseLog.WithName("ldap")
	// Setup default value
	if prvd.Port == "" {
		if prvd.InsecureNoSSL {
			prvd.Port = "389"
		} else {
			prvd.Port = "636"
		}
	}
	prvd.hostPort = fmt.Sprintf("%s:%s", prvd.Host, prvd.Port)
	if prvd.TimeoutSec == 0 {
		prvd.TimeoutSec = 10
	}
	// WARNING: This is a global variable
	ldap.DefaultTimeout = time.Duration(prvd.TimeoutSec) * time.Second

	//prvd.logger.V(2).Info("paths", "configFolder", configFolder, "RootCA", prvd.RootCA, "ClientCert", prvd.ClientCert, "ClientKey", prvd.ClientKey)
	prvd.RootCA = adjustPath(configFolder, prvd.RootCA)
	prvd.ClientCert = adjustPath(configFolder, prvd.ClientCert)
	prvd.ClientKey = adjustPath(configFolder, prvd.ClientKey)
	//prvd.logger.V(2).Info("adjusted paths", "RootCA", prvd.RootCA, "ClientCert", prvd.ClientCert, "ClientKey", prvd.ClientKey)

	prvd.tlsConfig = &tls.Config{ServerName: prvd.Host, InsecureSkipVerify: prvd.InsecureSkipVerify}
	if prvd.RootCA != "" || len(prvd.RootCAData) != 0 {
		var data []byte
		if len(prvd.RootCAData) != 0 {
			data = make([]byte, base64.StdEncoding.DecodedLen(len(prvd.RootCAData)))
			_, err := base64.StdEncoding.Decode(data, []byte(prvd.RootCAData))
			if err != nil {
				return &prvd, fmt.Errorf("error while parsing RootCAData : %w", err)
			}
		} else {
			var err error
			if data, err = os.ReadFile(prvd.RootCA); err != nil {
				return &prvd, fmt.Errorf("error while reading CA file: %w", err)
			}
		}
		rootCAs := x509.NewCertPool()
		if !rootCAs.AppendCertsFromPEM(data) {
			return &prvd, fmt.Errorf("no certs found in ca file")
		}
		prvd.tlsConfig.RootCAs = rootCAs
	}

	if prvd.ClientKey != "" && prvd.ClientCert != "" {
		cert, err := tls.LoadX509KeyPair(prvd.ClientCert, prvd.ClientKey)
		if err != nil {
			return &prvd, fmt.Errorf("load client cert failed: %v", err)
		}
		prvd.tlsConfig.Certificates = append(prvd.tlsConfig.Certificates, cert)
	}
	var ok bool
	prvd.userSearchScope, ok = parseScope(prvd.UserSearch.Scope)
	if !ok {
		return &prvd, fmt.Errorf("userSearch.Scope unknown value %q", prvd.UserSearch.Scope)
	}
	prvd.groupSearchScope, ok = parseScope(prvd.GroupSearch.Scope)
	if !ok {
		return &prvd, fmt.Errorf("groupSearch.Scope unknown value %q", prvd.GroupSearch.Scope)
	}
	return &prvd, nil
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
