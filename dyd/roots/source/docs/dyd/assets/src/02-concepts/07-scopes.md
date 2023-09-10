# Scopes

A workspace can contain a large number of packages, but often only a few are relevant at any given point in time.  

Many dryad commands offer some combination of filters (like `--include` and `--exclude`) to select a specific part of a workspace to operate on.  But, it can be annoying to remember and typeextra arguments all the time.

To make it easier to logically partition parts of a workspace, dryad offers the ability to create shortcut commands through **scopes**.  A scope is a collection of command or argument "aliases".

Most dryad commands have a `--scope` option, which allows you to specify the scope to run a command in.  When a scope is provided that specifies arguments for that command, dryad will_rewrite the command arguments before execution, using the scoped settings for that command.

You can also set a default scope using `dryad scopes default set`.  If no scope is provided, dryad will automatically use the default scope.

The scope name `none` is reserved to use as an escape hatch from scoping.  Using `--scope=none` will bypass the default scope and run the command as-is.

## Scripts

Dryad scope settings are by default plain-text files that include extra arguments to add to a command.  The one exception is the `dryad run <name>` / `dryad script run <name>` command.  This command interprets the corresponding scope setting as a standalone script to be executed, which allows you to add any commands you'd like to `dryad run`.  Keep in mind these scripts are scoped to the current scope, so you can use this to do things like create two different `dryad run build` commands for two different scopes.

You can create a scoped script using the `dryad script edit <name>` command, and see which scoped scripts are available to you using `dryad scripts list`.