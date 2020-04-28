
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

```
kubebuilder create api --group directory --version v1alpha1 --kind User
Create Resource [y/n]
y
Create Controller [y/n]
n
```

```
kubebuilder create api --group directory --version v1alpha1 --kind Group
kubebuilder create api --group directory --version v1alpha1 --kind GroupBinding
kubebuilder create webhook --group directory --version v1alpha1 --kind User --defaulting --programmatic-validation
```

- commit (Add api resources)

```
make manifests
```

To simplify stuff, will remove all stuff related to
- conversion webhook
- prometheus
- leader election

```
kubebuilder create webhook --group directory --version v1alpha1 --kind Group --defaulting --programmatic-validation
kubebuilder create webhook --group directory --version v1alpha1 --kind GroupBinding --defaulting --programmatic-validation
```

# Test add samples

in kspray/koomgr/dev or pod
```
k apply -f ../../../../../koobind/koomgr/samples/drivers.yaml
```
