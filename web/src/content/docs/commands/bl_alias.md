---
title: bl alias
description: Command reference for bl alias
---

Manage reusable command aliases

### Synopsis

Manage command aliases stored in boiler.conf.json.

Usage patterns:

  bl alias              List all aliases
  bl alias name=cmd     Set or update an alias
  bl alias name         Get one alias value
  bl unalias name       Remove an alias (use 'unalias' command)

Positional Arguments & Variables:
- Aliases support dynamic positional arguments (bl__1, bl__2, ...) and inline variables (bl__VAR_NAME).
- When positional arguments are used, any trailing unconsumed arguments are automatically appended.

Alias names are case-insensitive. Hyphens and underscores are preserved.

```bash
# Simple aliases
bl alias ll=ls
bl alias s=search

# Alias with positional argument - bl gi Node fetches Node.gitignore
bl alias gi='add github/gitignore:bl__1.gitignore . -m .gitignore -r'

# Alias referencing a stored var
bl var org=rishiyaduwanshi
bl alias templates='search --registry https://github.com/bl__org/boiler'
```

```
bl alias [name|name=command [args...]] [flags]
```

### Options

```
  -h, --help   help for alias
```

### Options inherited from parent commands

```
      --global    Force global scope for this command
      --local     Force local scope for this command
  -V, --verbose   Enable verbose debug output
```

