# Token lifecycle

[Back](../README.md)



Here is a summary of the initial interaction:

- The user issue a kubectl command (i.e. `kubectl get nodes`).
- kubectl request a token to koocli.
- As koocli does not host any token for now, it will request the user to provide a login and a password.
- Then koocli request a token to koo-manager, based on the provided credential.
- koo-manager check the credential against one or several Identity provider and return a token.
- koocli store the token localy and return it to kubectl.
- kubectl now issue the request to the API Server, with the token as authentication header.
- The API Server check the token validity and retrieve associated user and group binding by calling koo-manager. 
- Based on RBAC, the API server allow or denied the initial request.

If the user issue another command in a short period, the locally stored token will be used.


[Back](../README.md)

