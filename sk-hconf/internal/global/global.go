package global

import (
	"github.com/go-logr/logr"
	"k8s.io/client-go/kubernetes"
	"skas/sk-hconf/internal/config"
)

var Logger logr.Logger
var Config config.Config
var ClientSet *kubernetes.Clientset
