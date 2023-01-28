package clientprovider

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"skas/sk-merge/internal/config"
	"strings"
	"time"
)

func New(conf config.ClientProviderConfig) (ClientProvider, error) {
	cp := &clientProvider{
		ClientProviderConfig: conf,
	}
	var err error
	cp.httpClient, err = newHTTPClient(cp.HttpClientConfig.Url,
		[]string{cp.HttpClientConfig.RootCaPath, config.Conf.RootCaPath},
		[]string{cp.HttpClientConfig.RootCaData, config.Conf.RootCaData},
		cp.HttpClientConfig.InsecureSkipVerify,
	)
	if err != nil {
		return nil, err
	}
	return cp, nil
}

// TODO: Candidate to be shared in sk-common
// Copied from oauth.go connector. rawUrl parameter is just to
func newHTTPClient(rawUrl string, rootCAPaths []string, rootCaDatas []string, insecureSkipVerify bool) (*http.Client, error) {
	u, err := url.Parse(rawUrl)
	if err != nil {
		return nil, fmt.Errorf("unable to parse url '%s': %w", rawUrl, err)
	}
	var tlsConfig *tls.Config = nil
	if strings.ToLower(u.Scheme) == "https" {
		pool, err := x509.SystemCertPool()
		if err != nil {
			return nil, err
		}
		tlsConfig = &tls.Config{RootCAs: pool, InsecureSkipVerify: insecureSkipVerify}
		if !insecureSkipVerify {
			caCount := 0
			for _, rootCaPath := range rootCAPaths {
				if rootCaPath != "" {
					if err := appendCaFromFile(tlsConfig.RootCAs, rootCaPath); err != nil {
						return nil, err
					}
					caCount++
				}
			}
			for _, rootCaData := range rootCaDatas {
				if rootCaData != "" {
					if err := appendCaFromBase64(tlsConfig.RootCAs, rootCaData); err != nil {
						return nil, err
					}
					caCount++
				}
			}
			if caCount == 0 {
				return nil, fmt.Errorf("no root CA certificate was configured")
			}
		}
	}
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
			Proxy:           http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
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

func shortenB64(b64 string) string {
	if len(b64) <= 30 {
		return b64
	} else {
		return fmt.Sprintf("%s.......%s", b64[:10], b64[len(b64)-10:])
	}
}

func appendCaFromBase64(pool *x509.CertPool, b64 string) error {
	data := make([]byte, base64.StdEncoding.DecodedLen(len(b64)))
	_, err := base64.StdEncoding.Decode(data, []byte(b64))
	b64id := b64 + ".........."
	b64id = fmt.Sprintf("%s......%s", (b64 + "...........")[:10], b64id[len(b64)])
	if err != nil {
		return fmt.Errorf("error while parsing base64 root ca data %s : %w", shortenB64(b64), err)
	}
	if !pool.AppendCertsFromPEM(data) {
		return fmt.Errorf("invalid root CA certificate in %s", shortenB64(b64))
	}
	return nil
}
