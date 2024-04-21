package global

import (
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"skas/sk-hconf/internal/config"
)

var Logger logr.Logger
var Config config.Config
var KubeClient client.Client
