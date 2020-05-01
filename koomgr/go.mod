module github.com/koobind/koobind/koomgr

go 1.14

replace github.com/koobind/koobind/common v0.1.0 => ../common

require (
	github.com/go-logr/logr v0.1.0
	github.com/golang-collections/collections v0.0.0-20130729185459-604e922904d3
	github.com/koobind/koobind/common v0.1.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.5.1
	go.uber.org/zap v1.15.0
	golang.org/x/crypto v0.0.0-20200220183623-bac4c82f6975
	gopkg.in/asn1-ber.v1 v1.0.0-20181015200546-f715ec2f112d // indirect
	gopkg.in/fsnotify.v1 v1.4.7
	gopkg.in/ldap.v2 v2.5.1
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v0.18.2
	sigs.k8s.io/controller-runtime v0.6.0
)
