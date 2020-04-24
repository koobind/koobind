
# Installation in dev

cat >/usr/local/bin/kubectl-koo

```
#!/bin/bash

/Users/sa/dev/gX/gopath/bin/koocli "$@"
```
chmod +x /usr/local/bin/kubectl-koo



# Misc

https://stackoverflow.com/questions/2137357/getpasswd-functionality-in-go


go install && koocli login --rootCaFile ../../ezciac/certificates/CA/ca1.crt --server https://localhost:8443


