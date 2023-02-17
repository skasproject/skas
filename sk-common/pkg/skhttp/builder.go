package skhttp

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"skas/sk-common/pkg/misc"
	"strings"
	"time"
)

// Inspired from oauth.go connector.

func New(conf *Config, altRootCAPaths string, altRootCaDatas string) (Client, error) {
	// Just a test for validity. Not used in this function
	u, err := url.Parse(conf.Url)
	if err != nil {
		return nil, fmt.Errorf("unable to parse url '%s': %w", conf.Url, err)
	}
	var tlsConfig *tls.Config = nil
	if strings.ToLower(u.Scheme) == "https" {
		pool, err := x509.SystemCertPool()
		if err != nil {
			return nil, err
		}
		tlsConfig = &tls.Config{RootCAs: pool, InsecureSkipVerify: conf.InsecureSkipVerify}
		if !conf.InsecureSkipVerify {
			caCount := 0
			rootCaPaths := []string{conf.RootCaPath, altRootCAPaths}
			for _, rootCaPath := range rootCaPaths {
				if rootCaPath != "" {
					if err := appendCaFromFile(tlsConfig.RootCAs, rootCaPath); err != nil {
						return nil, err
					}
					caCount++
				}
			}
			rootCaDatas := []string{conf.RootCaData, altRootCaDatas}
			for _, rootCaData := range rootCaDatas {
				if rootCaData != "" {
					if err := appendCaFromBase64(tlsConfig.RootCAs, rootCaData); err != nil {
						return nil, err
					}
					caCount++
				}
			}
			//if caCount == 0 {
			//	return nil, fmt.Errorf("no root CA certificate was configured")
			//}
		}
	}
	httpclient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
			Proxy:           http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
	return &client{
		Config:     *conf,
		httpClient: httpclient,
	}, nil
}

func appendCaFromFile(pool *x509.CertPool, caPath string) error {
	rootCABytes, err := os.ReadFile(caPath)
	if err != nil {
		return fmt.Errorf("failed to read CA file '%s': %w", caPath, err)
	}
	if !pool.AppendCertsFromPEM(rootCABytes) {
		return fmt.Errorf("invalid root CA certificate in file %s", caPath)
	}
	return nil
}

func appendCaFromBase64(pool *x509.CertPool, b64 string) error {
	data := make([]byte, base64.StdEncoding.DecodedLen(len(b64)))
	_, err := base64.StdEncoding.Decode(data, []byte(b64))
	if err != nil {
		return fmt.Errorf("error while parsing base64 root ca data %s : %w", misc.ShortenString(b64), err)
	}
	if !pool.AppendCertsFromPEM(data) {
		return fmt.Errorf("invalid root CA certificate in %s", misc.ShortenString(b64))
	}
	return nil
}
