module skas/sk-ldap

go 1.21

replace skas/sk-common v0.2.2 => ../sk-common

require (
	github.com/go-logr/logr v1.2.3
	github.com/pior/runnable v0.11.0
	github.com/spf13/pflag v1.0.5
	gopkg.in/ldap.v2 v2.5.1
	skas/sk-common v0.2.2
)

require (
	github.com/bombsimon/logrusr/v4 v4.0.0 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/sirupsen/logrus v1.9.0 // indirect
	golang.org/x/sys v0.4.0 // indirect
	gopkg.in/asn1-ber.v1 v1.0.0-20181015200546-f715ec2f112d // indirect
	gopkg.in/fsnotify.v1 v1.4.7 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/apimachinery v0.26.1 // indirect
	k8s.io/klog/v2 v2.80.1 // indirect
	k8s.io/utils v0.0.0-20221128185143-99ec85e7a448 // indirect
)

