
# ctest

This test suite use a small tool: [ctest](https://github.com/SergeAlexandre/ctest).

To install it:
```
go get github.com/SergeAlexandre/ctest
```

Other requirement are:

- A kubernetes cluster with `koobind` installed.
- A kubectl configured to use `koobind`
- An `admin` account, member of the `kooadmin` group and having at least read access to the `koo-system` namespace.

 
