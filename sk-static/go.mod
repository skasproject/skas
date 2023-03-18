module skas/sk-static

go 1.19

replace skas/sk-common v0.2.0 => ../sk-common

require (
	github.com/go-logr/logr v1.2.3
	github.com/spf13/pflag v1.0.5
	golang.org/x/crypto v0.5.0
	gopkg.in/yaml.v2 v2.4.0
	skas/sk-common v0.2.0
)

require (
	github.com/bombsimon/logrusr/v4 v4.0.0 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/sirupsen/logrus v1.9.0 // indirect
	golang.org/x/sys v0.4.0 // indirect
	gopkg.in/fsnotify.v1 v1.4.7 // indirect
)
