---
title: bl new
description: Command reference for bl new
---

Run a Boiler command script (.bl)

### Synopsis

Run a .bl command script to generate, inject, or modify code in your project.

Boiler automatically parses any flags (like --ts or --port=3000) and maps them to script variables.

Script resolution is scope-aware:
- No boiler.local.json (or scope=global) : looks in ~/.boiler/commands/
- scope=local in boiler.local.json        : looks in ./bl/commands/ first, falls back to ~/.boiler/commands/
- --global flag                           : forces ~/.boiler/commands/ only
- --local flag                            : forces ./bl/commands/ only

```
bl new [script_name] [args...] [flags]
```

### Examples

```
  # Run the routes.bl script
  bl new routes

  # Run with positional arguments
  bl new routes user auth

  # Run with flags
  bl new routes --ts --port=3000

  # Force global commands directory
  bl new routes --global
```

### Options

```
  -h, --help   help for new
```

### Options inherited from parent commands

```
      --global    Force global scope for this command
      --local     Force local scope for this command
  -V, --verbose   Enable verbose debug output
```

