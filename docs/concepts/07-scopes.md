---
title: 07 - Scopes
layout: default
parent: Concepts
---

# Scopes

A workspace can contain a large number of packages, but often only a few are relevant at any given point in time.  

Many dryad commands offer some combination of filters (like `--include` and `--exclude`) to select a specific part of a workspace to operate on.  But, it can be annoying to remember and typeextra arguments all the time.

To make it easier to logically partition parts of a workspace, dryad offers the ability to create shortcut commands through **scopes**.  A scope is a collection of command or argument "aliases".

Most dryad commands have a `--scope` option, which allows you to specify the scope to run a command in.  When a scope is provided that specifies arguments for that command, dryad will_rewrite the command arguments before execution, using the scoped settings for that command.

You can also set a default scope using `dryad scopes default set`.  If no scope is provided, dryad will automatically use the default scope.

The scope name `none` is reserved to use as an escape hatch from scoping.  Using `--scope=none` will bypass the default scope and run the command as-is.