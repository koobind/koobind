
## Still to do 


- Add crd provider
- A Blacklist mechanisme on user/group/groupBinding stuff, superseding other provided information

- Add a ChangePassword capabillity on crd provider
- Add an url to fetch the CA for the kubconfig (And maybe to fetch on full kubeconfig).

- Package and setup client as krew plugin (krew template ?)
- Build automation (goreleaser ?)
- versionning: https://medium.com/better-programming/how-to-version-your-docker-images-1d5c577ebf54


- ldap bind passwrd in secret ?
- Allow koocli context to be defined in an ENV variable
- Introduce realm (GroupBinding urlPath/Provider chain/token Lifecycle)  (Using https://github.com/gorilla/mux ?)
- Add some 'client secret' to koo serve and/or certificate protection

- Refactor certificate management (Use cluster CA)
- Check client certificate
- Security against BFA.
- Add some NetworkSecurityPolicies around auth and webhook servers
- Automatic reload (or suicide) on configuration (configMap) change. See ?
