
# Koobind usage


<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Index**

- [User and Group management](#user-and-group-management)
  - [User](#user)
  - [Group](#group)
  - [GroupBinding](#groupbinding)
- [Login / Logout](#login--logout)
- [Tokens](#tokens)
- [Context](#context)
- [Context store](#context-store)
- [--kubeconfig option](#--kubeconfig-option)
- [k9s](#k9s)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->


## User and Group management

If you have followed up the full installation process, you can now create and manage Kubernetes users and groups as resources.

Here is a sample you can apply as is:

```
$ export KUBECONFIG=/etc/koobind/kubeconfig
$ kubectl apply -f https://raw.githubusercontent.com/koobind/koobind/master/samples/users.yaml
Login:admin
Password:
logged successfully..
user.directory.koobind.io/jsmith created
group.directory.koobind.io/devs created
groupbinding.directory.koobind.io/jsmith-devs created
```

Of course, you will need to be allowed to create resources in the `koo-system` namespace. This is the case of the 'admin' user we created during installation procedure. We used it in this sample.

### User

Here is the User definition with all attributes:

```
---
apiVersion: directory.koobind.io/v1alpha1
kind: User
metadata:
  name: jsmith
  namespace: koo-system
spec:
  commonName: John SMITH
  passwordHash: $2a$10$qumINdiGJIM1si2wi8ceDOczChq2twfDEDa6DR7jiYL8rJNzeYtmu
  email: jsmith@mycompany.com
  uid: 100001
  comment: A sample user
  disabled: no
```

> The password matching this hash is `smithj`.

- The user login is the `metadata.name` field value.
- All `spec.*` attribute are optionals.
- Namespace must be `koo-system` (This is not 'hard coded'. See [configuration reference](config.md))
- `spec.commonName`, `spec.email`, `spec.uid` and `spec.comment` attributes are for user description, and have no impact on the way this user is handled by `Koobind` and Kubernetes.
- Setting `spec.disabled` to `True` would make this user unable to login.

The only way to define a password it to provide a 'hash'. To generate such appropriate value, a sub-command `kubectl koo hash` is provided:

```
export KUBECONFIG=/etc/koobind/kubeconfig
kubectl koo hash
Password:
Confirm password:
$2a$10$6B93pPYM5EqejYV2MDCOAuEgJCfXfNysTdTvTCUGc.ON0gEVEY6Q.
```

Also, note the `spec.passwordHash` field is optional. What is interest of a user without password ? 
One answer is when combining several identity providers, user identification ca be provided by another one and this definition may populate this user with more attribute and GroupBinding.

### Group

Here is the Group definition with all attributes:

```
---
apiVersion: directory.koobind.io/v1alpha1
kind: Group
metadata:
  name: devs
  namespace: koo-system
spec:
  description: All developpers
  disabled: no
```

- All `spec.*` attribute are optionals.
- Namespace must be `koo-system`.
- `spec.description` attribute have no impact on the way this group is handled by `Koobind` and Kubernetes.
- Setting `spec.disabled` to `True` would make this group not existing. It will not be reported on any user's group list. 

### GroupBinding

Here is the GroupBinding definition with all attributes:

```
---
apiVersion: directory.koobind.io/v1alpha1
kind: GroupBinding
metadata:
  name: jsmith-devs
  namespace: koo-system
spec:
  user: jsmith
  group: devs
  disabled: no

```

- `spec.user` and `spec.group` attribute are mandatory.
- Namespace must be `koo-system`.
- Setting `spec.disabled` to `True` would make this binding transparent. It will not contribute to any user's group list.  

## Login / Logout

As Users are Kubernetes resources, we can list them:

```
$ kubectl -n koo-system get users
NAME     COMMON NAME   EMAIL                  UID      COMMENT         DISABLED
admin    Koo ADMIN
jsmith   John SMITH    jsmith@mycompany.com   100001   A sample user
```

Provided of course we are still logged as 'admin'.

Now, if we want to test this new user, we need to logout first. For this, there is the `koo logout` subcommand;

```
$ kubectl koo logout
Bye!
$ kubectl -n koo-system get users
Login:jsmith
Password:
logged successfully..
Error from server (Forbidden): users.directory.koobind.io is forbidden: User "jsmith" cannot list resource "users" in API group "directory.koobind.io" in the namespace "koo-system"
```

Of course, this user is not allowed to perform the requested operation. 

An alternate solution is to explicitly login using the `koo login` subcommand:

```
kubectl koo login
Login:jsmith
Password:
logged successfully..

$ kubectl koo whoami
user:jsmith  id:100001  groups:devs
```

(If a user was previously logged, a logout is performed)

One can also provide credential on the command line:

```
$ kubectl koo login --user jsmith
Password:
logged successfully..

$ kubectl koo whoami
user:jsmith  id:100001  groups:devs
```

or:

```
$ kubectl koo login --user jsmith --password smithj
logged successfully..

$ kubectl koo whoami
user:jsmith  id:100001  groups:devs
```

> Of course, providing the password in clear text such this way is a huge security issue. Use it as your own risk. 

## Tokens

All kubernetes authentication system is based on 'bearer tokens'. Token are provided by the user on each request (either explicitly, or implicitly, in our case by the `kubctl-koo client`).
Then the kubernetes apiserver validate this token against some authentication backend. `koo-manager` in our case.

A token can host some information. This is the case of JWT token, where the user name and groups are encoded in the token itself. 
Or it can be a meaningless random string, which act as a key in the backend to retrieve user information. This is the way it is for `koobind`.

Another aspect about token is time to live. `Koobind` implements two time limits for a token:

- An absolute token duration (Default: 24 H)
- An inactivity time out (Default: 30 m). If a token is not used during this period, it will expire.

The last mechanism is similar to the usual cookie based session logic used for most Web applications.
 
One can display the active tokens using the `koo describe tokens` subcommand:

```
$ kubectl koo describe  tokens
TOKEN                            USER   UID    GROUPS                CREATED ON     LAST HIT
uzvbzgjrhoqzdqzpjxomamrxqopdedba admin         clusteradmin,kooadmin 05-17 11:09:04 11:09:04
boqpsyrvhxjlimkusvacuvdyvxyqmphc jsmith 100001 devs                  05-17 11:09:34 11:09:34
jtaoyzandgoxmsnnybwkfseqbztpfmtg admin         clusteradmin,kooadmin 05-17 11:09:51 11:10:26
```

Of course, you must be member of the 'kooadmin' group. If not:

```
$ kubectl koo describe tokens
ERROR: You are not allowed to perform this operation!
```

Token themself are stored as Kubernetes resources. So we can also display them using standard kubectl commands: 

```
$ kubectl -n koo-system get tokens
NAME                               USER NAME   USER ID   USER GROUPS               LAST HIT
boqpsyrvhxjlimkusvacuvdyvxyqmphc   jsmith      100001    [devs]                    2020-05-17T11:09:34Z
jtaoyzandgoxmsnnybwkfseqbztpfmtg   admin                 [clusteradmin kooadmin]   2020-05-17T11:13:17Z
uzvbzgjrhoqzdqzpjxomamrxqopdedba   admin                 [clusteradmin kooadmin]   2020-05-17T11:09:04Z
```

As tokens represents active sessions, it could be useful to be able to cancel them:

```
$ kubectl koo cancel token boqpsyrvhxjlimkusvacuvdyvxyqmphc
Token boqpsyrvhxjlimkusvacuvdyvxyqmphc is successfully cancelled
 
$ kubectl koo describe  tokens
 TOKEN                            USER  UID GROUPS                CREATED ON     LAST HIT
 uzvbzgjrhoqzdqzpjxomamrxqopdedba admin     clusteradmin,kooadmin 05-17 11:09:04 11:09:04
 jtaoyzandgoxmsnnybwkfseqbztpfmtg admin     clusteradmin,kooadmin 05-17 11:09:51 11:15:04
```

## Context

When testing authentication/authorization rules, switching back and forth to different users by logout/login can quickly be painful.

To avoid this, we can make use of the context mechanisms build in the 'kubeconfig' file. One solution could be to make two different kubeconfig files. 
If your client configuration has been deployed as described in the [installation](installation.md) section, duplicate the kubeconfig file and edit each version to have a different context name:  

```
$ cd /etc/koobind
$ sudo cp kubeconfig kubeconfig1
$ sudo cp kubeconfig kubeconfig2
$ sudo vi kubeconfig1

....
- context:
    cluster: mycluster.local
    user: koo-user
  name: koo1@mycluster.local
current-context: koo1@mycluster.local
....

$ sudo vi kubeconfig2

....
- context:
    cluster: mycluster.local
    user: koo-user
  name: koo2@mycluster.local
current-context: koo2@mycluster.local
....
```

Then, you can open two terminals and set appropriate value in KUBECONFIG environment variable:

On terminal1

```
$ export KUBECONFIG=/etc/koobind/kubeconfig1
$ kubectl -n koo-system get users
Login:admin
Password:
logged successfully..
NAME     COMMON NAME   EMAIL                  UID      COMMENT         DISABLED
admin    Koo ADMIN
jsmith   John SMITH    jsmith@mycompany.com   100001   A sample user

[15:57:39 sa@kspray1:/etc/koobind]$ kubectl koo whoami
user:admin  id:  groups:clusteradmin,kooadmin
```

On terminal 2

```
export KUBECONFIG=/etc/koobind/kubeconfig2
$ kubectl -n koo-system get users
Login:jsmith
Password:
logged successfully..
Error from server (Forbidden): users.directory.koobind.io is forbidden: User "jsmith" cannot list resource "users" in API group "directory.koobind.io" in the namespace "koo-system"

$ kubectl koo whoami
user:jsmith  id:100001  groups:devs
```

If you switch back and forth between the two terminals, you can check each session is isolated.

> An alternate solution would be to create a single kubeconfig file with several context inside. 
Then, you will be able to switch between with the command `kubectl config use-context <newContext>`.

## Context store

The `kubectl-koo` client active tokens are stored locally in the folder `~/.kube/cache/koo/<contextName>/`. 

A `koo config` subcommand will allow to display the different context used. With the active one spotted by a `*`.

``` 
$ kubectl koo config
  CONTEXT              SERVER                CA
  koo1@mycluster.local https://kspray1:31444 /etc/koobind/certs/koomgr-ca.crt
* koo2@mycluster.local https://kspray1:31444 /etc/koobind/certs/koomgr-ca.crt
  koo@mycluster.local  https://kspray1:31444 /etc/koobind/certs/koomgr-ca.crt
```

## --kubeconfig option

Another solution to switch between context is to provide the `--kubeconfig` option on the command line. For this to works:

- For all `koo ...` subcommands, this option must be provided after the `koo ...` subcommand:

```
$ kubectl koo logout --kubeconfig=/etc/koobind/kubeconfig1 
```

- A `context' parameter with the relevant context name must be provided to the command in the 'kubeconfig' files:

```
$ sudo vi kubeconfig1

....
- context:
    cluster: mycluster.local
    user: koo-user
  name: koo1@mycluster.local
current-context: koo1@mycluster.local
kind: Config
preferences: {}
users:
- name: koo-user
  user:
    exec:
      apiVersion: "client.authentication.k8s.io/v1beta1"
      command: kubectl-koo
      args:
      - auth
      - --server=https://kspray1:31444    # <---- Adjust FQDN to one of your node you included in the certificate
      - --rootCaFile=/etc/koobind/certs/koomgr-ca.crt
      - --context=koo1@mycluster.local
```

## k9s

To conclude this chapter, we would like to say two words about this great tool which is [k9s](https://github.com/derailed/k9s) 

As it is able to handle Custom Resources Definition out of the box, K9s is a perfect tool to dynamically display, modify or delete `koobind` resources.
 
Note than, as User and Group are ambiguous names, which are used also by others API, alias are provided to ensure ambiguous access.
 
For example, you can access this screen under `koouser` resource name:
  
![](draw/k9s-users.png)

This one using `kootoken`:

![](draw/k9s-tokens.png)

This one using `koogroup`:

![](draw/k9s-groups.png)

This one using `koogroupbinding`:

![](draw/k9s-groupbindings.png)

