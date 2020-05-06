# KOOBIND

koobind is a Kubernetes extension focussed on user authentication.

It can authenticate users in a fully autonomous way, or can leverage one or several LDAP identity providers such as OpenLDAP or ActiveDirectory.

One main feature is its ability to merge result from several identity providers, thus allowing handling of users and groups defined at different level (Corporate, Departmental, Team, Cluster, ...)

Another main advantage is it only require a ReadOnly access to the LDAP/AD server(s). User profile can then be enriched with local information.

## Index

- [Overview](#overview)
- [Installation](docs/installation.md)
- [Identity provider merging]
- [Bearer token lifecycle]
- [Configuration reference]

x

x

x

x

x
x

x

x

x

x
x

x

x

x

x
x

x

x

x

x
x

x

x

x

x
x

x

x

x

x
x

x

x

x

x
x

x

x

x

x
x

x

x

x

x
x

x

x

x

x
x

x

x

x

x
x

x

x

x

x
x

x

x

x

x
x

x

x

x

x
x

x

x

x

x

x

<a name="overview"></a>
## Overview

Technically, Koobind can be defined as:

- A token provider. Checking user credential and delivering time limited tokens.
- A Kubernetes Authentication Webhook, allowing API Server to validate the token associated to each request.
- A kubectl plugin.
- A set of CRD (Custom Resources Definition), allowing definition of users and groups as standard Kubernetes resources.

This involves the following components:

![](docs/koo1-Overview.jpg) 


x

x

x
x

x

x

x

x

x




