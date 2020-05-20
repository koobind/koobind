# Identity providers chaining

One of the most interesting feature of Koobind is its ability to 'chain' several identity providers.


<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Index**

- [Base Example](#base-example)
- [User enrichment](#user-enrichment)
- [Group renaming](#group-renaming)
- [Unique authentication reference](#unique-authentication-reference)
- [Chaining rules reference](#chaining-rules-reference)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Base Example

We will illustrate this by an example: Let's say we want to authenticate users against: 

- A corporate LDAP server.
- A departmental LDAP server.
- Our local CRD based directory.

The corresponding configuration file will look like:

```
logLevel: 0
adminGroup: "kooadmin"
providers:

  - name: ipa1
    type: ldap
    host: ipa1.vgr.broadsoftware.com
    port: 636
    ....

  - name: ldap1
    type: ldap
    host: ldap1
    port: 636
    ....

  - name: crdsys
    type: crd
```

Detail of LDAP configuration were removed for better readability. Full version [here](https://raw.githubusercontent.com/koobind/koobind/master/samples/configs/chain/config.yml). 

Now, let's say we have a user 'oriley' which is defined in both LDAP, but with different groups binding. 
And, we also have our 'admin' user defined in the **crdsys** provider, during [installation](installation.md).

A sample interaction could be:

```
$ kubectl koo login
Login:oriley
Password:
logged successfully..

$ kubectl koo whoami
user:oriley  id:870200001  groups:auditors,itdep,users
```

We see we can log under this user, it has a numerical id of 870200001 and belong to 3 groups (auditors,itdep,users).

A system admin would like to figure out how this was defined. 

For this, it first needs to log as admin:

```
$ kubectl koo login --user admin
Password:
logged successfully..
```

Then, it can issue the `koo describe user` subcommand:

```
$ kubectl koo describe user oriley
PROVIDER FOUND AUTH UID       GROUPS        EMAIL                COMMON NAME  COMMENT
ipa1     *     *    870200001 [users,itdep] oriley@mycompany.com Oliver RILEY
ldap1    *     +    2004      [auditors]                         Oliver RILEY
crdsys
```

We can see than:

- This user exists in both **ipa1** and **ldap1** providers
- This user will be authenticated by **ipa1** provider. 
- **ldap1** provider would be also able to authenticate this user by itself. But this will not be the case, has **ipa1** take precedence (Was before in the provider list).
- **ipa1** and **ldap1** provider both provide a numerical id. But only the one from **ipa1** (who authenticate this user) is retained.
- The resulting user's group list is the concatenation of all groups found in all providers

Of course, we can also figure out how our `admin` user is defined:

``` 
kubectl koo describe user admin
PROVIDER FOUND AUTH UID GROUPS                  EMAIL COMMON NAME COMMENT
ipa1
ldap1
crdsys   *     *        [clusteradmin,kooadmin]       Koo ADMIN
```

Now, let's assume we have a user 'jsmith' defined in **ldap1**. And also this user is already defined in our CRD provider, as performed in the [usage](usage.md) chapter.

```
$ kubectl koo describe user jsmith
PROVIDER FOUND AUTH UID    GROUPS     EMAIL                COMMON NAME  COMMENT
ipa1
ldap1    *     *    2005   [all,devs]                      Johnny SMITH
crdsys   *     +    100001 [devs]     jsmith@mycompany.com John SMITH
```

For this user to log successfully, only the password matching the value in the **ldap1** provider is valid. Ths password defined in our **crdsys** CRD provider is now 'hidden'. 

```
$ kubectl koo login
Login:jsmith
Password:
logged successfully..
[13:37:25 sa@kspray1:~]$ kubectl koo whoami
user:jsmith  id:2005  groups:all,devs
```

## User enrichment

Now, let's say we want to grant our 'oriley' user full admin rights on our cluster.

The obvious solution would be to create the kooadmin group on the corporate LDAP and to bind this user on. But, let's say this request can't be fulfilled (at least in a reasonable amount of time).
So, we don't want to modify the corporate LDAP provider. 

So, we will grant this locally by applying the following manifest:

```
---
apiVersion: directory.koobind.io/v1alpha1
kind: User
metadata:
  name: oriley
  namespace: koo-system
spec: {}
---
apiVersion: directory.koobind.io/v1alpha1
kind: GroupBinding
metadata:
  name: oriley-kooadmin
  namespace: koo-system
spec:
  user: oriley
  group: kooadmin
---
apiVersion: directory.koobind.io/v1alpha1
kind: GroupBinding
metadata:
  name: oriley-cluseradmin
  namespace: koo-system
spec:
  user: oriley
  group: clusteradmin
```

In this manifest we recreate the user, for local coherency and then bind it to the 'kooadmin' and 'clusteradmin' groups.

To apply it:

```
$ kubectl apply -f https://raw.githubusercontent.com/koobind/koobind/master/samples/oriley-admin.yaml
user.directory.koobind.io/oriley created
groupbinding.directory.koobind.io/oriley-kooadmin created
groupbinding.directory.koobind.io/oriley-cluseradmin created
```

Then:

```
$ kubectl koo login --user oriley  # Need to re-log to activate the new bindings
Password:
logged successfully..

$ kubectl koo describe user oriley
PROVIDER FOUND AUTH UID       GROUPS                  EMAIL                COMMON NAME  COMMENT
ipa1     *     *    870200001 [users,itdep]           oriley@mycompany.com Oliver RILEY
ldap1    *     +    2004      [auditors]                                   Oliver RILEY
crdsys   *                    [clusteradmin,kooadmin]                                   [No password]

# Ensure we have full admin rights on the cluster:
$ kubectl get pods --all-namespaces
NAMESPACE        NAME                                       READY   STATUS    RESTARTS   AGE
cert-manager     cert-manager-6f578f4565-n5k8l              1/1     Running   10         61d
cert-manager     cert-manager-cainjector-75b6bc7b8b-rfsvn   1/1     Running   25         24d
cert-manager     cert-manager-webhook-8444c4bc77-c62sv      1/1     Running   8          61d
...
```

## Group renaming

Let's have a look back on our user 'jsmith':

```
$ kubectl koo describe user jsmith
PROVIDER FOUND AUTH UID    GROUPS     EMAIL                COMMON NAME  COMMENT
ipa1
ldap1    *     *    2005   [all,devs]                      Johnny SMITH
crdsys   *     +    100001 [devs]     jsmith@mycompany.com John SMITH

$ kubectl koo login --user jsmith
Password:
logged successfully..

$ kubectl koo whoami
user:jsmith  id:2005  groups:all,devs
```

One can see both providers grant it access to the 'devs' group.

If you want to distinguish groups from different providers, the `groupPattern` parameter will allow you to add prefix or suffix for all groups of a given provider.

For example, if you modify the configuration of **ldap1** this way:

```
logLevel: 0
adminGroup: "kooadmin"
providers:
    ....

  - name: ldap1
    type: ldap
    host: ldap1
    port: 636
    groupPattern: "dep-%s"
    ....
```

Now, under 'admin' account:
 
```
$ kubectl koo describe user jsmith
PROVIDER FOUND AUTH UID    GROUPS             EMAIL                COMMON NAME  COMMENT
ipa1
ldap1    *     *    2005   [dep-all,dep-devs]                      Johnny SMITH
crdsys   *     +    100001 [devs]             jsmith@mycompany.com John SMITH

$ kubectl koo login --user jsmith
Password:
logged successfully..
[16:34:44 sa@kspray1:~]$ kubectl koo whoami
user:jsmith  id:2005  groups:dep-all,dep-devs,devs
``` 

## Unique authentication reference

Now, let's imagine a new corporate security policy would establish than all users credential must be defined only in the corporate LDAP server. 
In other words, there should be no way to define a user outside of this server.

This can be achieved by modifying the configuration, adding `credentialAuthority: no` flags to providers other than the corporate one:  

```
logLevel: 0
adminGroup: "kooadmin"
providers:

  - name: ipa1
    type: ldap
    host: ipa1.vgr.broadsoftware.com
    port: 636
    ....

  - name: ldap1
    type: ldap
    credentialAuthority: no
    host: ldap1
    port: 636
    ....

  - name: crdsys
    type: crd
    credentialAuthority: no
```

After restarting the server, one can check only users referenced in **ipa1** (Only 'oriley' in our case) can log on. But, of course, others provider still add their groups bindings.

## Chaining rules reference

Here is a more formal description of the chaining rules, expressed in pseudo-code:

```
groups = []
logged = false
for (provider in providerList) {
    if (providerEnabled == true) {
        if (user found in provider) {
            if logged == false and (a password is defined for this user in this provider) and provider.credentialAuthority == true) {
                if (checkPassword() == OK ) {
                    logged = true
                } else {
                    return "login failed"
                }
            }
            if provider.groupAuthority {
                for (group in user.groups) {
                    groups.append(format(provider.groupPatten, group))
                }
            } 
        }
    }
}
if logged == true {
    deduplicate(groups)
    sort(groups)
    return "login OK", groups
} else {
    return "login failed"
}
```

