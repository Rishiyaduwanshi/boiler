---
title: bl info
description: Command reference for bl info
---

Show detailed information about a resource

### Synopsis

Display detailed information about a stored snippet or stack.

Shows:
- Full path in store
- File size (for snippets)
- File count and total size (for stacks)
- Last modified time
- Version information

```
bl info [resource] [flags]
```

### Examples

```
  # Show snippet info
  bl info errorHandler@1.js

  # Show stack info
  bl info express-api@1

  # Without version (shows all versions)
  bl info logger
```

### Options

```
  -h, --help   help for info
```

### Options inherited from parent commands

```
      --global    Force global scope for this command
      --local     Force local scope for this command
  -V, --verbose   Enable verbose debug output
```

