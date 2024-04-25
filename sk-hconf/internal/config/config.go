package config

import (
	"time"
)

type Config struct {
	// Shared by 'patcher' and 'monitor'
	ApiServerNamespace string `yaml:"apiServerNamespace"`
	ApiServerPodName   string `yaml:"apiServerPodName"`
	// ----------------------------- Used by patcher
	// The 5 following values are interpreted inside the container, so depends of the 'hostPath' configuration
	ApiServerManifestPath string `yaml:"apiServerManifestPath"`
	KubernetesCAPath      string `yaml:"kubernetesCAPath"`
	SkasFolder            string `yaml:"skasFolder"`
	BackupFolder          string `yaml:"backupFolder"`
	TmpFolder             string `yaml:"tmpFolder"`
	//
	HookConfigContent string `yaml:"hookConfigContent"`
	// This is where to lookup the CA used by the `skauth` module
	CertificateAuthority struct {
		Secret struct {
			Namespace string `yaml:"namespace"`
			Name      string `yaml:"name"`
		}
		KeyInData string `yaml:"keyInData"`
	} `yaml:"certificateAuthority"`
	WebhookCacheTtl time.Duration `yaml:"webhookCacheTtl"`
}
