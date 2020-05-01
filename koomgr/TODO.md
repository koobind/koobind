
## Still to do 

- Token in API Server 
- Docs
- Tests, tests, .....

# Improvments

- A Blacklist mechanisme on user/group/groupbinding stuff, superseding other provided information

A set of command/api on crd provider to:
- Create/update/delete user/group/groupbinding
- Unlog a user (ie after removing it from a group)
- Add a ChangePassword capability on crd provider
- Coherency check on each CRD providers

- Add an url to fetch the CA for the kubconfig (And maybe to fetch on full kubeconfig).

- Package and setup client as krew plugin (krew template ?)
- Build automation (goreleaser ?)
- versionning: https://medium.com/better-programming/how-to-version-your-docker-images-1d5c577ebf54

- ldap bind passwrd in secret ?
- Allow koocli context to be defined in an ENV variable
- Introduce realm (GroupBinding urlPath/Provider chain/token Lifecycle)  (Using https://github.com/gorilla/mux ?)
- Add some 'client secret' to koo serve and/or certificate protection

- Check client certificate
- Security against BFA.
- Add some NetworkSecurityPolicies around auth and webhook servers
- Automatic reload (or suicide) on configuration (configMap) change. See ?
