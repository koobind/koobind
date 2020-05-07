# Koobind installation

[Back](../README.md)

The process described here will install koobind in a simple case, with only a first identity provider based on CRD.

Once this first step is completed, you will be able to easily add one (or several) LDAP/AD identity provider. 

## Manifest Deployment

First step is to deploy some Kubernetes manifests:
 
```
kubectl apply -f https://raw.githubusercontent.com/koobind/koobind/<release>/koomgr/yaml/crd.yaml
kubectl apply -f https://raw.githubusercontent.com/koobind/koobind/<release>/koomgr/yaml/pod/deploy.yaml
kubectl apply -f https://raw.githubusercontent.com/koobind/koobind/<release>/koomgr/yaml/rbac.yaml
```

Where `<release>` should be replaced by the latest appropriate [release value](https://github.com/koobind/koobind/releases)

Then you will need to deploy the initial configuration, as a configMap
```
kubectl apply -f https://raw.githubusercontent.com/koobind/koobind/sample/simpleconf.yaml
``` 

Note all deployment will occur inside the namespace `koo-system`

At this step, the koo-manager pod should be running:

```
kubectl -n koo-system get pod
```

And the logs should not mention any errors:

```
kubectl -n koo-system logs koo-manager-XXXXXXX
```

# Endpoints



# Next step

To test this initial installation, we now suggest you create a first user and group, allowed to deploy a sample 'hello Kubernetes' application in a dedicated namespace.


[Back](../README.md)
