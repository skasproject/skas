package cmwatcher

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"os"
	"path/filepath"
	"reflect"
	"skas/sk-common/pkg/datawatcher"
	"sync"
	"time"
)

var _ datawatcher.DataWatcher = &cmWatcher{}

type cmWatcher struct {
	sync.Mutex
	clientSet  *kubernetes.Clientset
	cmName     string
	cmKey      string
	context    context.Context
	namespace  string
	parserFunc datawatcher.ParserFunc
	logger     logr.Logger
	watcher    watch.Interface
	content    interface{}
}

func New(ctx context.Context, cmName string, cmKey string, parserFunc datawatcher.ParserFunc, logger logr.Logger, namespace string, kubeconfig string) (datawatcher.DataWatcher, error) {
	if namespace == "" {
		// Fetch current namespace
		file, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
		if err != nil {
			return nil, fmt.Errorf("watcher on 'cm:???/%s': Unable to lookup current namespace: %w", cmName, err)
		}
		namespace = string(file)
	}
	cmw := &cmWatcher{
		namespace:  namespace,
		cmName:     cmName,
		cmKey:      cmKey,
		context:    ctx,
		parserFunc: parserFunc,
		logger:     logger,
	}
	clientSet, err := buildClientSet(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("watcher on %s: Unable to build kubernetes client: %w", cmw.getName(), err)
	}
	// Read initial data
	configMap, err := clientSet.CoreV1().ConfigMaps(namespace).Get(ctx, cmName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("watcher on %s: Unable to access configMap: %w", cmw.getName(), err)
	}
	data, ok := configMap.Data[cmKey]
	if !ok {
		return nil, fmt.Errorf("watcher on %s: Unable to find key '%s' in configMap", cmw.getName(), cmKey)
	}
	cmw.content, err = parserFunc(data)
	if err != nil {
		return nil, fmt.Errorf("watcher on %s: Error while parsing data in users configMap: %w", cmw.getName(), err)
	}
	cmw.watcher, err = clientSet.CoreV1().ConfigMaps(namespace).Watch(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("watcher on %s: Unable to initialize watcher on configMap= %w", cmw.getName(), err)
	}
	return cmw, nil
}

func (cmw *cmWatcher) getName() string {
	return fmt.Sprintf("'cm:%s/%s'", cmw.namespace, cmw.cmName)
}

func (cmw *cmWatcher) Get() interface{} {
	cmw.Lock()
	defer cmw.Unlock()
	return cmw.content
}

func (cmw *cmWatcher) Run(ctx context.Context) error {
	return cmw.Start(ctx)
}

func (cmw *cmWatcher) Start(ctx context.Context) error {
	running := true
	go func() {
		for {
			select {
			case event, ok := <-cmw.watcher.ResultChan():
				if !ok {
					if !running {
						return
					}
					cmw.logger.Error(nil, "watcher stopped. Will  restart...", "watcher", cmw.getName())
					for {
						var err error
						cmw.watcher, err = cmw.clientSet.CoreV1().ConfigMaps(cmw.namespace).Watch(cmw.context, metav1.ListOptions{})
						if err != nil {
							cmw.logger.Error(err, "watcher unable to restart for now. Will retry in 10sec", "watcher", cmw.getName())
							time.Sleep(10 * time.Second)
						} else {
							cmw.logger.Info("watcher restarted successfully", "watcher", cmw.getName())
							break
						}
					}
				}
				configMap, ok := event.Object.(*v1.ConfigMap)
				if !ok {
					cmw.logger.Error(nil, "watcher: unexpected event type", "watcher", cmw.getName(), "type", reflect.TypeOf(event.Object))
					continue
				}
				cmw.logger.V(1).Info("watcher info", "watcher", cmw.getName(), "eventType", event.Type)
				if configMap.Name == cmw.cmName {
					data := ""
					if event.Type == "DELETED" {
						cmw.logger.Info("configMap has been deleted", "watcher", cmw.getName())
					} else {
						// We handle MODIFIED and also ADDED, even if redundant with Get(), as there is a (very small) risk to loose a modification occurring between Get() and Watch()
						data, ok = configMap.Data[cmw.cmKey]
						if !ok {
							cmw.logger.Error(fmt.Errorf("unable to find key in configMap"), "Users DB modification ineffective", "watcher", cmw.getName(), "key", cmw.cmKey)
							continue
						}
					}
					content, err := cmw.parserFunc(data)
					if err != nil {
						cmw.logger.Error(fmt.Errorf("invalid yaml data in users configMap: %w", err), "Users DB modification ineffective", "watcher", cmw.getName())
						continue
					}
					cmw.logger.Info("Users configMap has been successfully reloaded", "watcher", cmw.getName())
					cmw.Lock()
					cmw.content = content
					cmw.Unlock()
				}
			}
		}
	}()

	// Block until the stop channel is closed.
	<-ctx.Done()

	running = false
	cmw.watcher.Stop()

	return nil
}

func buildClientSet(kubeconfig string) (*kubernetes.Clientset, error) {
	var config *rest.Config = nil
	var err error
	if kubeconfig == "" {
		if envvar := os.Getenv("KUBECONFIG"); len(envvar) > 0 {
			kubeconfig = envvar
		}
	}
	if kubeconfig == "" {
		config, err = rest.InClusterConfig()
	}
	if config == nil {
		home := homedir.HomeDir()
		if kubeconfig == "" && home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
		}
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}
