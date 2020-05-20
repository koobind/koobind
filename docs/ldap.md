# LDAP configuration

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Index**

- [Overview](#overview)
- [Security considerations](#security-considerations)
- [Configuration reference](#configuration-reference)
- [Sample configurations](#sample-configurations)
- [ConfigMap building](#configmap-building)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Overview

The LDAP Identity Provider allow login/password authentication, backed by an LDAP directory.

The provider perform the following operation

- lookup the user based on the user login.
- Validate the user password.
- Build a list of groups the user belongs to.

When using in a multi provider context, some of these operation can be skipped, depending of configuration. More [here](providerchain.md)

> The Koobind LDAP identity provider was strongly inspired from the LDAP connector of [DEX project](https://github.com/dexidp/dex). 
Great thanks to all its contributors.

## Security considerations

Koobind attempts to bind with the backing LDAP server using the admin and end user's plain text password. 
Though some LDAP implementations allow passing hashed passwords, koobind doesn't support hashing and instead strongly recommends that all administrators just use TLS. 
This can often be achieved by using port 636 instead of 389, and by handling certificate authority. 

Koobind currently allows insecure connections to ensure connectivity with the wide variety of LDAP implementations and to ease initial setup. 
But such configuration should never be used in a production context, as they are actively leaking passwords.

## Configuration reference

| Name                       | Req? | Default | Description                                                                                                                  |
|----------------------------|:----:|:-------:|------------------------------------------------------------------------------------------------------------------------------|
| name                       | yes  | -       | The provider name                                                                                                            |
| type                       | yes  | -       | Must be `ldap` in this case.                                                                                                 |
| enabled                    | no   | true    | Allow to disable this provider                                                                                               |
| credentialAuthority        | no   | true    | Is this provider authority for password checking                                                                             |
| groupAuthority             | no   | true    | Is user's group to be fetched and added to user's group list                                                                 |
| critical                   | no   | true    | If true , a failure on this provider will leads 'invalid login'. Even if another provider grants access                      |
| groupPattern               | no   | `%s`    | Group pattern. Allow to add prefix and/or suffix to group name.                                                              |
| uidOffset                  | no   | 0       | Will be added to the returned Uid                                                                                            |
| host                       | yes  | -       | The hostname of the LDAP Server                                                                                              |
| port                       | yes  | (1)     | The port of the LDAP Server                                                                                                  |
| insecureNoSSL              | no   | false   | If true, LDAP host connection will be in clear text                                                                          |
| insecureSkipVerify         | no   | false   | If true, Server certificate check will be skipped                                                                            |
| startTLS                   | no   | false   | Connect to the insecure port then issue a StartTLS command to negotiate a secure connection.<br>If false secure connections will use the LDAPS protocol. |
| rootCA                     | no   | -       | (2) Path to a trusted root certificate file, to check LDAP server certificate                                                    |
| rootCAData                 | no   | -       | (2) Base64 encoded PEM data containing root CAs, to check LDAP server certificate                                                |
| clientCert                 | no   | -       | Path to client certificate file if LDAP server require client authentication.                                                |
| clientKey                  | no   | -       | Path to client key file if LDAP server require client authentication.                                                        |
| bindDN                     | yes  | -       | LDAP admin account. Used to search for users and groups. May be a ReadOnly access                                            |
| bindPW                     | yes  | -       | Password for the bindDN account                                                                                              |
| userSearch.baseDN          | yes  | -       | BaseDN to start the user search from. For example "cn=users,dc=example,dc=com"                                               |
| userSearch.filter          | no   | -       | Optional filter to apply when searching the directory. For example "(objectClass=person)"                                    |
| userSearch.loginAttr       | yes  | -       | Attribute to match against the login. This will be translated and combined with the other filter as "(<loginAttr>=<login>)". |
| userSearch.scope           | no   | sub     | Can either be:<br>- `sub`: search the whole sub tree<br>- `one`: only search one level                                       |
| userSearch.numericalIdAttr | no   | -       | The attribute providing the numerical user ID                                                                                |
| userSearch.emailAttr       | no   | -       | The attribute providing the user's email                                                                                     |
| userSearch.cnAttr          | no   | -       | The attribute providing the user's common name                                                                               |
| groupSearch.baseDN         | yes  | -       | BaseDN to start the groups search from. For example "cn=groups,dc=example,dc=com"                                            |
| groupSearch.filter         | no   | -       | Optional filter to apply when searching the directory. For example "(objectClass=posixGroup)"                                |
| groupSearch.scope          | no   | sub     | Can either be:<br>- `sub`: search the whole sub tree<br>- `one`: only search one level                                       |
| groupSearch.nameAttr       | yes  | -       | The attribute of the group that represents its name.                                                                         |
| groupSearch.linkGroupAttr  | yes  | -       | (3) The group entry attribute used for the group/user relationship                                                           |
| groupSearch.LinkUserAttr   | yes  | -       | (3) The user entry attribute used as value for the group/user relationship.                                                  |


- The 8 first parameters are common to all Identity providers.
- (1): 389 if insecureNoSSL, 636 otherwise.
- (2) rootCA and rootCAData are exclusive. The goal is to provide the CA who issued the LDAP server certificate. By providing a file (rootCA) or by   
- (3) The filter for group/user relationship will be: `<linkGroupAttr>=<Value of LinkUserAttr for the user>`. If there is several values for LinkUserAttr, system will loop on.

## Sample configurations

Here is a sample configuration aimed to connect to a FreeIPA LDAP server:

```
logLevel: 0
adminGroup: "kooadmin"
providers:
  - name: ipa1
    type: ldap
    host: ipa1.vgr.broadsoftware.com
    port: 636
    rootCA: /etc/koo/cfg/ipa1-cert.pem
    bindDN: uid=admin,cn=users,cn=accounts,dc=vgr,dc=broadsoftware,dc=com
    bindPW: ipaadmin
    userSearch:
      baseDN: cn=users,cn=accounts,dc=vgr,dc=broadsoftware,dc=com
      emailAttr: mail
      filter: (objectClass=inetOrgPerson)
      loginAttr: uid
      numericalIdAttr: uidNumber
      cnAttr: cn
    groupSearch:
      baseDN: cn=groups,cn=accounts,dc=vgr,dc=broadsoftware,dc=com
      filter: (objectClass=posixgroup)
      linkGroupAttr: member
      linkUserAttr: DN
      nameAttr: cn
```

And another sample, here aimed to access an OpenLDAP server:

```
logLevel: 0
adminGroup: "kooadmin"
providers:
  - name: ldap1
    type: ldap
    host: ldap1
    port: 636
    bindDN: cn=Manager,dc=vgr,dc=broadsoftware,dc=com
    bindPW: LdapAdmin
    rootCA: /etc/koo/cfg/ldap1-ca1.crt
    userSearch:
      baseDN: ou=Users,dc=vgr,dc=broadsoftware,dc=com
      filter: (objectClass=inetOrgPerson)
      loginAttr: uid
      emailAttr: mail
      numericalIdAttr: uidNumber
      cnAttr: cn
    groupSearch:
      baseDN: ou=Groups,dc=vgr,dc=broadsoftware,dc=com
      filter: (objectClass=posixgroup)
      nameAttr: cn
      linkGroupAttr: memberUid
      linkUserAttr: uid
```

Note there is no `insecureNoSSL` nor `insecureSkipVerify` parameters. This means we want to have a secure connection and enforce certificate validation. 
So, we must provide the CA of this server certificate.

For this, the `rootCA` parameters target the `/etc/koo/cfg` folder, which is mapped to the `mgrconfig` configMap (See [Configuration](configuration.md)). 

## ConfigMap building

Building a `configMap` with several files is quite simple: Put all files in a dedicated folder and use the appropriate `kubectl` commands.

For example, setup the following folder:

``` 
$ tree ipa1
ipa1
|-- config.yml
`-- ipa1-cert.pem
```

And issue the following commands:

```
$ kubectl -n koo-system delete configmap mgrconfig   # Need to delete previous instance
$ kubectl create configmap mgrconfig -n koo-system --from-file=./ipa1/ 
```

Of course, you will need to delete the `koo-manager` pod for the deployment to restart it and the configuration to be effective.

An alternate solution could be to generate an intermediate file:
 
``` 
$ kubectl create configmap mgrconfig -n koo-system --from-file=./ipa1/ --dry-run=client -o yaml >mgrconfig-ipa1.yaml
```

(On older version of kubectl, you should remove `=client` from the `--dry-run` option)

And apply it:

```
$ kubectl -n koo-system delete configmap mgrconfig   # Need to delete previous instance
$ kubectl apply -f mgrconfig-ipa1.yaml
```


