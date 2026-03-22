---
title: bl var
description: Command reference for bl var
---

Manage reusable command and snippet variables

### Synopsis

Manage reusable variables stored in boiler.conf.json.

Usage patterns:
- bl var                 List all variables
- bl var name=value      Set or update a variable
- bl var name            Get one variable value
- bl unvar name          Remove a variable (use 'unvar' command)

Variable names are normalized internally:
- Optional prefixes : and bl__ are accepted
- Names are case-insensitive
- Hyphens and underscores are preserved

Examples:
- bl var API_URL=`https://api.example.com`
- bl var bl__TEAM_REG=`https://github.com/myorg/boiler`
- bl var api_url

```
bl var [name|name=value] [flags]
```

### Options

```
  -h, --help   help for var
```

