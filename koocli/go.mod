module github.com/koobind/koobind/koocli

go 1.14

replace github.com/koobind/koobind/common v0.1.0 => ../common
replace github.com/koobind/koobind/koomgr v0.1.0 => ../koomgr

require (
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/koobind/koobind/common v0.1.0
	github.com/koobind/koobind/koomgr v0.1.0
	github.com/sirupsen/logrus v1.5.0
	github.com/spf13/cobra v1.0.0
	golang.org/x/crypto v0.0.0-20200414173820-0848c9571904
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v0.18.2
	k8s.io/utils v0.0.0-20200414100711-2df71ebbae66 // indirect
)
