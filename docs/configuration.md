# Configuration

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Index**

  - [Overview](#overview)
  - [The koo-manager container configuration](#the-koo-manager-container-configuration)
  - [The configMap](#the-configmap)
  - [Global parameters reference](#global-parameters-reference)
- [The adminGroup](#the-admingroup)
  - [The namespaces](#the-namespaces)
  - [CRD Identity provider configuration](#crd-identity-provider-configuration)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->


## Overview

This chapter is about the `koo-manager` module configuration.

Configuration is mainly provided in a file (typically `config.yml`). This file is then embedded in a 'configMap' to be accessed from inside the `koo-manager` pod.

The configuration can also host some reference to other files, such as certificates. In this case, these files will be embedded also in the 'configMap'.

Some parameters can also be provided on the command line. This is mostly used when executing the process out of Kubernetes, in development phases. 
Command line parameters always take precedence over value from configuration file.
 
Also, most of the parameters have appropriate default values, in order to keep simple case simple.

Here is a simple configuration file, the basic one used in [Installation](installation.md) chapter

```
logLevel: 0
adminGroup: "kooadmin"
providers:
  - name: crdsys
    type: crd
```

It can be described in two parts:

- A set of global parameters (here logLevel and adminGroup)
- A list of provider, each with its specific coonfiguration.

## The koo-manager container configuration

Here is an extract of the deployment manifest:

```
....
      containers:
      - args:
        - --namespace
        - $(KOO_NAMESPACE)
        - --config
        - /etc/koo/cfg/config.yml
        command:
        - /manager
        env:
        - name: KOO_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        image: koobind/manager:X.X.X
        name: manager
        ....
        volumeMounts:
        - mountPath: /tmp/k8s-webhook-server/serving-certs
          name: cert
          readOnly: true
        - mountPath: /etc/koo/cfg
          name: config
      ....
      volumes:
      - name: cert
        secret:
          defaultMode: 420
          secretName: webhook-server-cert
      - configMap:
          name: mgrconfig
        name: config
```

One can see than the namespace is deduced from the launching environment. 

Another parameter is set to define the config file location. And this location refer to the `/etc/koo/cfg/` folder, which is mapped to the configMap `mgrconfig` 
(See `volumeMounts` and `volumes` entries)

Also, another volume is mounted on a secret `webhook-server-cert`. This secret host the server certificate and has been created automatically by the Kubernetes `Certificate-manager`. 

## The configMap

Given the basic configuration described above, here is the corresponding 'configMap' manifest:

```
apiVersion: v1
kind: ConfigMap
metadata:
  name: mgrconfig
  namespace: koo-system
data:
  config.yml: |2
    logLevel: 0
    adminGroup: "kooadmin"
    providers:
      - name: crdsys
        type: crd
```

So, if you need to add/change some values, the easiest way is to edit this manifest and apply it again.

Once applied, you must delete the `koo-manager` pod. It will be restarted (as part of a deployment) and will take the new configuration in account. 

> In case you don't already know this nice tool, have a look on [k9s](https://github.com/derailed/k9s) to perform such kubernetes admin task.

You can also use the `kubectl create configmap mgrconfig -n koo-system --from-file=.....` for more complex configuration. You will find a sample in the [LDAP](ldap.md) chapter. 

## Global parameters reference

| Name                       | Req? | Default | Description                                                                                                                  |
|----------------------------|:----:|:-------------------------------------:|------------------------------------------------------------------------------------------------|
| webhookServer.host         | no   | all                                   | Webhook server bind address.                                                                   |
| webhookServer.port         | no   | 8443                                  | Webhook server bind port                                                                       |
| webhookServer.certdir      | no   | /tmp/k8s-webhook-server/serving-certs | Path to the webhook server certificate folder. Must contain two files: `tls.crt` and `tls.key` |
| authServer.host            | no   | all                                   | Auth server bind address                                                                       |
| authServer.port            | no   | 8444                                  | Auth server bind port                                                                          |
| authServer.certdir         | no   | /tmp/k8s-webhook-server/serving-certs | Path to the Auth server certificate folder. Must contain two files: `tls.crt` and `tls.key`    |
| logLevel                   | no   | 0                                     | Log level (0:INFO; 1:DEBUG, 2:MoreDebug...)                                                    |
| logMode                    | no   | json                                  | Log mode: 'dev' or 'json'                                                                      |
| adminGroup                 | no   | -                                     | Only user belonging to this group will be able to access admin commands                        |
| inactivityTimeout          | no   | 30m                                   | After this period without token validation, the session expire                                 |
| sessionMaxTTL              | no   | 24H                                   | After this period, the session expire, in all case.                                            |
| clientTokenTTL             | no   | 30s                                   | Token caching value for the client (kubectl-koo)                                               |
| tokenStorage               | no   | crd                                   | How token are stored: 'memory' or 'crd'                                                        |
| namespace                  | no   | -                                     | Default namespace for tokenNamespace and CRD                                                   |
| tokenNamespace             | no   | `namespace` value                     | Tokens storage namespace when tokenStorage==crd                                                |
| LastHitStep                | no   | 3                                     | Delay to store lastHit in CRD, when tokenStorage==crd. In % of inactivityTimeout               |
| providers                  | yes  | -                                     | List of Identity providers                                                                     |

- There is no reason to modify `webhookServer` and `authServer` configuration (The first 6 parameters), except when running outside of Kubernetes, for development. 
The default value are appropriate with the deployment manifest and the `certificate-manager` configuration.   

- There is two token storage engines: In memory and in API server, as a Custom Resource. Of course, using In memory storage will unlog all users in case the `koo-manager` pod is restarted.

- Using CRD to store tokens, one should take care of not overloading the API Server. 
Internally, the `koo-manager` process handle a cache (thanks to the [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime) library), 
so the main issue is to limit write accesses. Normally, each time a token is used, we should update the `lastHit` field. To limit the rate of such operation, 
we update the field only after a small amount of time. This is defined by the `lastHitStep` parameter.  

# The adminGroup

By definition, all users belonging to this group will be able to perform management task on the `Koobind` system, such list/cancel tokens, describe users, etc...

The `koo-manager` will use this configuration value to validate if the logged user is allowed to perform an admin operation. (Typically using `kubectl-koo`).

To complete this, a good practice is to grant access to the members of this group to the `koo-system` namespace using the Kubernetes RBAC system. 
For example, this has been performed during the [installation process](installation.md#admin-configuration) by apply the `kooadmin.yaml` manifest.   

## The namespaces

A namespace different from the default value cas be defined for token storage. And also, several CRD providers can be defined in different namespace. 
A reason to do so can be to allow different access rules per namespace. 

> WARNING: In such case, all the alternate namespaces must exist for the `koo-manager` pod to start successfully!

## CRD Identity provider configuration

Here is the list of all parameters for the Identity provider based on CRD (Custom Resource Definition):

| Name                       | Req? | Default                  | Description                                                                                              |
|----------------------------|:----:|:------------------------:|----------------------------------------------------------------------------------------------------------|
| name                       | yes  | -                        | The provider name                                                                                        |
| type                       | yes  | -                        | Must be `crd` in this case.                                                                              |
| enabled                    | no   | true                     | Allow to disable this provider                                                                           |
| credentialAuthority        | no   | true                     | Is this provider authority for password checking                                                         |
| groupAuthority             | no   | true                     | Is user's group to be fetched and added to user's group list                                             |
| critical                   | no   | true                     | If true , a failure on this provider will leads 'invalid login'. Even if another provider grants access  |
| groupPattern               | no   | `%s`                     | Group pattern. Allow to add prefix and/or suffix to group name.                                          |
| uidOffset                  | no   | 0                        | Will be added to the returned Uid                                                                        |
| namespace                  | no   | global `namespace` value | The namespace to store all resources. 

The 8 first parameters are common to all Identity providers, and are relevant when building a chain of provider. See [Provider chaining](providerchain.md).

The only specific parameter is the namespace where to store the User/Group/GroupBinding ressources. For advanced configuration, 
the system will allow chaining several CRD based providers using differents namespace.


 