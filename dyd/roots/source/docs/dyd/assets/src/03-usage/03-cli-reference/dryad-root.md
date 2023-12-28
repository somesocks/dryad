# dryad root

```
$ dryad root --help
dryad root

Description:
    commands to work with a dryad root

Sub-commands:
    dryad root ancestors      list all roots the selected root depends on (directly and indirectly)
    dryad root build          build a specified root
    dryad root copy           make a copy of the specified root at a new location
    dryad root create         create a new root at the target path
    dryad root descendants    list all roots that depend on the selected root (directly and indirectly)
    dryad root develop        create a temporary development environment for a root
    dryad root link           link a root as a dependency of the current root
    dryad root move           move a root to a new location and correct all references
    dryad root path           return the base path of the current root
    dryad root replace        replace all references to one root with references to another
    dryad root requirements   list all requirements of this root
    dryad root unlink         remove a dependency linked to the current root
```