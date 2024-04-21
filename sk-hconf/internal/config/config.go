package config

import (
	"time"
)

type Config struct {
	// Shared by 'patcher' and 'monitor'
	ApiServerNamespace    string `yaml:"apiServerNamespace"`
	ApiServerManifestPath string `yaml:"apiServerManifestPath"`
	ApiServerPodName      string `yaml:"apiServerPodName"` // # TODO: Fatch from apiServerManifestPath/metadata.name
	// Used by patcher
	SkasFolder           string `yaml:"skasFolder"`
	HookConfigContent    string `yaml:"hookConfigContent"`
	CertificateAuthority struct {
		Secret struct {
			Namespace string `yaml:"namespace"`
			Name      string `yaml:"name"`
		}
		JsonPath string `yaml:"jsonPath"`
	} `yaml:"certificateAuthority"`
	WebhookCacheTtl time.Duration `yaml:"webhookCacheTtl"`
}
