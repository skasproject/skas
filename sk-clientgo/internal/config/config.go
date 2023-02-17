package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	osuser "os/user"
	"path"
	"skas/sk-clientgo/internal/kubecontext"
	"skas/sk-clientgo/internal/loadsave"
	"skas/sk-clientgo/internal/log"
	"skas/sk-common/pkg/misc"
	"skas/sk-common/pkg/skhttp"
)

// Exposed global variables

var Conf *Config
var SkhttpClient skhttp.Client

type Config struct {
	skhttp.Config
}

func Load() {
	configPath := buildPath(kubecontext.KubeContext)
	if loadsave.LoadStuff(configPath, func(decoder *yaml.Decoder) error {
		return decoder.Decode(&Conf)
	}) {
		log.Log.V(1).Info("LoadConfig()", "path", configPath, "server", Conf.Url, "rootCaPath", Conf.RootCaPath, "rootCaData", misc.ShortenString(Conf.RootCaData), "clientId", Conf.ClientAuth.Id, "clientSecret", "*****", "insecureSkipVerify", Conf.InsecureSkipVerify)
	} else {
		log.Log.V(1).Info("LoadConfig() -> nil", "configPath", configPath)
	}
}

func Save() error {
	configPath := buildPath(kubecontext.KubeContext)
	log.Log.V(1).Info("SaveConfig()", "configPath", configPath, "server", Conf.Url, "rootCaPath", Conf.RootCaPath, "rootCaData", misc.ShortenString(Conf.RootCaData), "clientId", Conf.ClientAuth.Id, "clientSecret", "*****", "insecureSkipVerify", Conf.InsecureSkipVerify)
	err := loadsave.SaveStuff(configPath, func(encoder *yaml.Encoder) error {
		return encoder.Encode(Conf)
	})
	if err != nil {
		return fmt.Errorf("error while saving configuration in '%s': %w", configPath, err)
	}
	return nil
}

func buildPath(context string) string {
	usr, err := osuser.Current()
	if err != nil {
		panic(err)
	}
	return path.Join(usr.HomeDir, fmt.Sprintf(".kube/cache/skas/%s/config.json", context))
}

func InitHttpClient() error {
	var err error
	SkhttpClient, err = skhttp.New(&Conf.Config, "", "")
	if err != nil {
		return err
	}
	return nil
}
