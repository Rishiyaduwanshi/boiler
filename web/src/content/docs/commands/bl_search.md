---
title: bl search
description: Command reference for bl search
---

Search for snippets or stacks

### Synopsis

Search for resources in your store by name.

Searches both snippets and stacks by default. Use flags to filter:
- Use -s or --snippets to search only snippets
- Use -k or --stacks to search only stacks

Search is case-insensitive and matches partial names.

Remote Search:
Use -r flag to search remote registry:
- Default registry from config
- Or specify custom: --registry `https://github.com/other/boiler`
- Use config variable reference: --registry :team_reg

```
bl search [query] [flags]
```

### Examples

```
  # Search for anything with 'error'
  bl search error

  # Search only snippets
  bl search logger --snippets

  # Search only stacks
  bl search express --stacks

  # Search remote registry
  bl search express -r

  # Search custom registry
  bl search express -r --registry https://github.com/myorg/boiler

  # Search registry from config variable
  bl search express -r --registry :team_reg
```

### Options

```
  -h, --help              help for search
      --registry string   Custom registry URL (overrides config)
  -r, --remote            Search remote registry
  -n, --snippets          Search only snippets
  -k, --stacks            Search only stacks
```

### Options inherited from parent commands

```
  -V, --verbose   Enable verbose debug output
```

