---
title: bl new
description: Command reference for bl new
---

Generate code using a Boiler script (.bl)

### Synopsis

Run a Boiler script (.bl) to generate, inject, or modify code in your project.

Boiler automatically parses any flags (like --ts or --port=3000) and maps them to script variables.
It looks for the script in the './bl/' folder of your current project.

```
bl new [script_name] [args...] [flags]
```

### Examples

```
  # Run the routes.bl script with positional arguments
  bl new routes user auth

  # Run with flags
  bl new routes --ts --port=3000
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

