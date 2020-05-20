# Ansible installation 

As an alternative to the manual installation procedure, it is possible to proceed with a fully automated installation by using [Ansible](https://www.ansible.com/).

You will find an Ansible role [at this location](https://github.com/BROADSoftware/ezcplugins/tree/master/k8s/koomgr/roles).

Keep in mind this role was developed targeting a vanilla Kubernetes cluster build with [kubespray](https://github.com/kubernetes-sigs/kubespray). As such, it does not pretend to be operational our of the box. 
In fact, it should be understood as a code base to adapt for your specific environment.

Also, note although this role will handle all the cluster part of the installation, the client part (`koobind-koo`, `kubeconfig` file) should still be performed manually.

In the simplest case, with the default configuration, this role may be used directly

```
- hosts: kube-master
  roles:
  - koomgr
```

Where `kube-master` is an Ansible group with all nodes hosting a kube-apiserver instance.

As usual with Ansible, have a look in the `..../defaults/main.yml` file to figure out all the variables which can be set to customize the deployment.

Here is a sample playbook with an LDAP provider definition (Freeipa in this case)

```
- hosts: kube-master
  vars:
    koomgr_config:

      logLevel: 0
      adminGroup: kooadmin

      providers:
        - type: ldap
          name: ipa
          bindDN: uid=admin,cn=users,cn=accounts,dc=mycompany,dc=com
          bindPW: myipapassword
          host: ipa.mycompany.com
          insecureNoSSL: false
          insecureSkipVerify: false
          port: 636
          rootCA: /some/local/folder/ipa/ca.pem
          startTLS: false
          userSearch:
            baseDN: cn=users,cn=accounts,dc=mycompany,dc=com
            emailAttr: mail
            filter: (objectClass=inetOrgPerson)
            loginAttr: uid
            numericalIdAttr: uidNumber
          groupSearch:
            baseDN: cn=groups,cn=accounts,dc=mycompany,dc=com
            filter: (objectClass=posixgroup)
            linkGroupAttr: member
            linkUserAttr: DN
            nameAttr: cn

        - name: crdsys
          type: crd
  roles:
    - koomgr
```

You will find full description of LDAP configuration [here](ldap.md). 

A point worth noting: The rootCA attribute in this definition must define the location of the CA file on the local, deployment system. 
The playbook will take care of copying the file on some location on the target node, and to mount this location in the `koo-manager` container.

It will also take care of modifying the rootCA attribute in the configuration to the 'in container' path.   

