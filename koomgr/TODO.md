
## Still to do 

- Add email to retrieved parameters (Return a list). And a common name
- Improve get (describe) user (Add a consolidated result)

- A user/group mapping using regexes. Or prefix/suffix to be reversible)
- Add a ChangePassword capabillity on crd provider
- A Blacklist mechanisme on user/group/binding stuff, superseding other provided information
- Change Binding to GroupBinding (Conflict with Binding in core)
- Add an url to fetch the CA for the kubconfig (And maybe to fetch on full kubeconfig).

- Package and setup client as krew plugin (krew template ?)
- Build automation (goreleaser ?)
- versionning: https://medium.com/better-programming/how-to-version-your-docker-images-1d5c577ebf54


- ldap bind passwrd in secret ?
- Allow koocli context to be defined in an ENV variable
- Introduce realm (Binding urlPath/Provider chain/token Lifecycle)  (Using https://github.com/gorilla/mux ?)
- Add some 'client secret' to koo serve and/or certificate protection

- Refactor certificate management (Use cluster CA)
- Check client certificate
- Security against BFA.
- Add some NetworkSecurityPolicies around auth and webhook servers
- Automatic reload (or suicide) on configuration (configMap) change. See ?
