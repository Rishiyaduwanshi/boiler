---
title: bl alias
description: Command reference for bl alias
---

Manage reusable command aliases

### Synopsis

Manage command aliases stored in boiler.conf.json.

Usage patterns:
- bl alias              List all aliases
- bl alias name=cmd     Set or update an alias
- bl alias name         Get one alias value
- bl unalias name       Remove an alias (use 'unalias' command)

Alias names are normalized internally:
- Names are case-insensitive
- Hyphens and underscores are preserved

Examples:
- bl alias ll=ls
- bl alias s=search
- bl alias ll

```
bl alias [name|name=command] [flags]
```

### Options

```
  -h, --help   help for alias
```

