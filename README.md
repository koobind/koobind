
# WARNING: This doc is currently out of phase with the code in this v0.3.0 branch

# KOOBIND

**Koobind** is a Kubernetes extension focused on user authentication.

It can authenticate users in a fully autonomous way, or can leverage one or several LDAP identity providers such as OpenLDAP or ActiveDirectory.

Beside handling LDAP, Koobind will also allow you to **manage users and groups as standard Kubernetes ressources**.

One main feature is its ability to **merge result from several identity providers**, thus allowing handling of users and groups defined at different level (Corporate, Departmental, Team, Cluster, ...)

Another main advantage is it **only require a ReadOnly access to the LDAP/AD server(s)**. User profile can then be enriched with local information.

**Koobind** is a 'pure' CLI tool. As such, it does not require any browser based interaction, thus allowing usage on remote system without any graphical UI.

## Warning

**This extension should be considered as Alpha release, still immature and under development. Use it at your own risk.**

And of course, all contribution are welcome.

## Index

- [Overview](#overview)
- [Installation](docs/installation.md)
- [Usage](docs/usage.md)
- [Configuration](docs/configuration.md)
- [LDAP configuration](docs/ldap.md)
- [Identity provider chaining](docs/providerchain.md)
- [Ansible installation](docs/ansible.md)
- [Design](docs/design.md)

## Overview

Technically, `Koobind` can be defined as:

- A token provider. Checking user's credential and delivering time limited tokens.
- A Kubernetes Authentication Webhook, allowing API Server to validate the token associated to each request.
- A kubectl plugin.
- A set of CRD (Custom Resources Definition), allowing definition of users and groups as standard Kubernetes resources.

This involves the following components:

![](docs/draw/koo1-Overview.jpg) 

- kubectl and Kubernetes API server are usual Kubernetes components.
- kubectl-koo is a kubectl plugin, providing a seamless user interaction.
- koo-manager is a pod running in Kubernetes and handling both requests from kubectl-koo and from the API Server. It also handles the logic to merge information from several Identity Providers.

Curently two kind of identity providers are supported:

- External LDAP server(s)
- CRD (Custom Resources Definition) based directory, stored in Kubernetes. 

From the Kubernetes point of view, `Koobind` must just answer two questions:

- Is this user successfully authenticated?
- What is the list of groups it belong to?

Based on this, the Kubernetes RBAC system will be able to allow or denied the requested operation.
