package readiness

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/go-logr/logr"
	"io"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net"
	"net/http"
	"net/url"
	"os"
	"skas/sk-hconf/internal/global"
	"strings"
	"time"
)

type Probe interface {
	IsReady() error
	WaitForDown(timeout time.Duration, mark bool) error
	WaitForUp(timeout time.Duration, mark bool) error
}

type probe struct {
	url string
	//*v1.HTTPGetAction
	httpClient *http.Client
	logger     logr.Logger
	pod        string // For logs
	//token      []byte
}

var _ Probe = &probe{}

func GetProbe(nodeName string) (Probe, error) {
	logger := global.Logger.WithName("probe").WithValues("nodeName", nodeName)
	// First, list pod
	pods, err := global.ClientSet.CoreV1().Pods(global.Config.ApiServerNamespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	foundApiSrv := 0
	for _, pod := range pods.Items {
		if strings.HasPrefix(pod.Name, global.Config.ApiServerPodName) {
			foundApiSrv++
			if pod.Spec.NodeName == nodeName {
				logger = logger.WithValues("pod", pod.Name)
				logger.V(1).Info("Find pod")
				ep := pod.Spec.Containers[0].ReadinessProbe.HTTPGet
				u := fmt.Sprintf("%s://%s:%s", ep.Scheme, ep.Host, ep.Port.String())
				u, err = url.JoinPath(u, ep.Path)
				if err != nil {
					return nil, err
				}
				probe := &probe{
					url:    u,
					logger: logger,
					pod:    pod.Name,
				}
				probe.httpClient, err = buildHttpClient(pod.Spec.Containers[0].ReadinessProbe.HTTPGet.Scheme, global.Config.KubernetesCAPath)
				if err != nil {
					return nil, err
				}
				// To validate the url
				_, err := http.NewRequest("GET", probe.url, nil)
				if err != nil {
					return nil, err
				}
				//probe.token, err = os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
				//if err != nil {
				//	return nil, err
				//}
				logger.Info("Setup Readiness", "url", probe.url)
				return probe, nil
			}
		}
	}
	if foundApiSrv == 0 {
		return nil, fmt.Errorf("no pod namad %s-... found in namespace '%s'", global.Config.ApiServerPodName, global.Config.ApiServerNamespace)
	} else {
		return nil, fmt.Errorf("no pod namad %s-... found for nodeName '%s'", global.Config.ApiServerPodName, nodeName)
	}
}

func (p *probe) IsReady() error {
	req, err := http.NewRequest("GET", p.url, nil)
	if err != nil {
		return err // Should not occurs, as http.NewRequest() has been tested in GetProbe()
	}
	//req.Header.Set("Authorization", "Bearer "+string(p.token))
	resp, err := p.httpClient.Do(req)
	if err != nil {
		p.logger.V(2).Info("%v on httpClient.Do()", err)
		return err
	}
	ba, err := io.ReadAll(resp.Body)
	if err != nil {
		p.logger.V(2).Info("%v on io.ReadAll()", err)
		return err
	}
	p.logger.V(2).Info("IsReady", "status", resp.Status, "statusCode", resp.StatusCode, "body", string(ba))
	if resp.StatusCode != 200 {
		return fmt.Errorf("statudCode:%d", resp.StatusCode)
	}
	return nil
}

func (p *probe) WaitForDown(timeout time.Duration, mark bool) error {
	limit := time.Now().Add(timeout)
	if mark {
		fmt.Printf("Wait for %s down:", p.pod)
	} else {
		p.logger.V(0).Info("Wait for pod down")
	}
	for {
		if mark {
			fmt.Printf(".")
		}
		if p.IsReady() != nil {
			if mark {
				fmt.Printf("DOWN\n")
			} else {
				p.logger.V(0).Info("pod DOWN")
			}
			return nil
		}
		//fmt.Printf("timeout: %s\n  now: %s\nlimit: %s\n", timeout, time.Now(), limit)
		if time.Now().After(limit) {
			if mark {
				fmt.Printf("TIMED OUT!\n")
			}
			return fmt.Errorf("time out expired on waitForDown(%s)", p.pod)
		}
		time.Sleep(time.Millisecond * 1000)
	}
}

func (p *probe) WaitForUp(timeout time.Duration, mark bool) error {
	limit := time.Now().Add(timeout)
	if mark {
		fmt.Printf("Wait for %s up:", p.pod)
	} else {
		p.logger.V(0).Info("Wait for pod up")
	}
	for {
		if mark {
			fmt.Printf(".")
		}
		if p.IsReady() == nil {
			if mark {
				fmt.Printf("UP\n")
			} else {
				p.logger.V(0).Info("pod UP")
			}
			return nil
		}
		//fmt.Printf("timeout: %s\n  now: %s\nlimit: %s\n", timeout, time.Now(), limit)
		if time.Now().After(limit) {
			if mark {
				fmt.Printf("TIMED OUT!\n")
			}
			return fmt.Errorf("time out expired on waitforUp(%s)", p.pod)
		}
		time.Sleep(time.Millisecond * 1000)
	}
}

func buildHttpClient(scheme v1.URIScheme, rootCAPath string) (*http.Client, error) {
	var tlsConfig *tls.Config = nil
	if strings.ToLower(string(scheme)) == "https" {
		pool, err := x509.SystemCertPool()
		if err != nil {
			return nil, err
		}
		tlsConfig = &tls.Config{RootCAs: pool, InsecureSkipVerify: false}
		err = appendCaFromFile(tlsConfig.RootCAs, rootCAPath)
		if err != nil {
			return nil, err
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
	return httpclient, nil
}

func appendCaFromFile(pool *x509.CertPool, caPath string) error {
	rootCaBytes, err := os.ReadFile(caPath)
	if err != nil {
		return fmt.Errorf("failed to read CA file '%s': %w", caPath, err)
	}
	if !pool.AppendCertsFromPEM(rootCaBytes) {
		return fmt.Errorf("invalid root CA certificate in file %s", caPath)
	}
	return nil
}
