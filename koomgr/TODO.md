
## Still to do 

- A user/group mapping using regexes. Or prefix/suffix to be reversible)
- Automatic reload (or suicide) on configuration (configMap) change. See ?
- Add a ChangePassword capabillity on crd provider
- A Blacklist mechanisme on user/group/binding stuff, superseding other provided information
- Add email to retrieved parameters (Return a list). And a common name


- Package and setup client as krew plugin (krew template ?)
- Build automation (goreleaser ?)
- versionning: https://medium.com/better-programming/how-to-version-your-docker-images-1d5c577ebf54


- Refactor loggin using https://github.com/jeanphorn/log4go (Make a logr interface)

- ldap bind passwrd in secret ?
- Allow koocli context to be defined in an ENV variable
- Introduce realm (Binding urlPath/Provider chain/token Lifecycle)  (Using https://github.com/gorilla/mux ?)
- Add some 'client secret' to koo serve and/or certificate protection

- Refactor certificate management (Use cluster CA)
- Check client certificate
- Security against BFA.
- Setup auth server on the webhook server, but launch another server (so another port) for access from koocli, this allowing differents Network security policies.

