# dryad roots

```
$ dryad roots --help
dryad roots

Description:
    commands to work with dryad roots

Sub-commands:
    dryad roots affected   take a list of files from stdin, and print a list of roots that may depend on those files
    dryad roots build      build selected roots in a garden
    dryad roots graph      print the local dependency graph of all roots in the garden
    dryad roots list       list all roots that are dependencies for the current root (or roots of the current garden, if the path is not a root)
    dryad roots owning     list all roots that are owners of the provided files. The files to check should be provided as relative or absolute paths through stdin.
    dryad roots path       return the path of the roots dir
```