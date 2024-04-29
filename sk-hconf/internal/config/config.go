package config

import "time"

type Config struct {
	ApiServerNamespace string `yaml:"apiServerNamespace"`
	ApiServerPodName   string `yaml:"apiServerPodName"`
	Image              string `yaml:"image"` // Image for the patcher
	ImagePullPolicy    string `yaml:"imagePullPolicy"`
	ServiceAccount     string `yaml:"serviceAccount"`
	// ----------------------------- Used by patcher
	// The 5 following values are interpreted inside the container, so depends of the 'hostPath' configuration
	ApiServerManifestPath string        `yaml:"apiServerManifestPath"`
	KubernetesCAPath      string        `yaml:"kubernetesCAPath"`
	SkasFolder            string        `yaml:"skasFolder"`
	BackupFolder          string        `yaml:"backupFolder"`
	TmpFolder             string        `yaml:"tmpFolder"`
	WebhookCacheTtl       time.Duration `yaml:"webhookCacheTtl"`
	SkasNamespace         string        `yaml:"skasNamespace"`
	TimeoutApiServer      time.Duration `yaml:"timeoutApiServer"` // Timeout on apiserver restart
	// This is where to lookup the CA used by the `skauth` module
	CertificateAuthority struct {
		Secret struct {
			Namespace string `yaml:"namespace"`
			Name      string `yaml:"name"`
		}
		KeyInData string `yaml:"keyInData"`
	} `yaml:"certificateAuthority"`
}
