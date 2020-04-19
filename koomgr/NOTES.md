
# Project setup

```
mkdir koobind
cd koobind/
git init
mkdir koomgr
cd koomgr/
go mod init github.com/koobind/koobind/koomgr
kubebuilder init --domain koobind.io
kubebuilder edit --multigroup=true
```

Edit hack/boilerplate.go.txt ans main.go to adjust licensing
Edit go.mod to adjust go version

- Initial commit
