module skas/sk-padl

go 1.21

replace skas/sk-common v0.2.2 => ../sk-common

require (
	github.com/go-logr/logr v1.2.3
	github.com/nmcclain/ldap v0.0.0-20210720162743-7f8d1e44eeba
	github.com/spf13/pflag v1.0.5
	skas/sk-common v0.2.2
)

require (
	github.com/bombsimon/logrusr/v4 v4.0.0 // indirect
	github.com/nmcclain/asn1-ber v0.0.0-20170104154839-2661553a0484 // indirect
	github.com/sirupsen/logrus v1.9.0 // indirect
	golang.org/x/sys v0.4.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
