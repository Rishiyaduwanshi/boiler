---
title: bl var
description: Command reference for bl var
---

Manage reusable command and snippet variables

### Synopsis

Manage reusable variables stored in boiler.conf.json.

Usage patterns:
- bl var                  List all variables
- bl var name=value       Set or update a variable
- bl var name             Get one variable value
- bl unvar name           Remove a variable (use 'unvar' command)

Variables can be used inline in ANY command using the bl__VAR_NAME syntax:
- When Boiler encounters bl__VAR_NAME in an argument, it replaces it with the variable's value.
- Useful for dynamic paths, URLs, and registry configurations.

Examples:
- # Set variables
- bl var org=rishiyaduwanshi
- bl var repo=boiler-templates

  # Use variables inline in a command (fetches github.com/rishiyaduwanshi/boiler-templates:auth)
  bl use github.com/bl__org/bl__repo:auth

```
bl var [name|name=value] [flags]
```

### Options

```
  -f, --force   Overwrite if already exists
  -h, --help    help for var
```

### Options inherited from parent commands

```
      --global    Force global scope for this command
      --local     Force local scope for this command
  -V, --verbose   Enable verbose debug output
```

