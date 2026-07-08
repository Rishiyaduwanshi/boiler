---
title: bl path
description: Command reference for bl path
---

Show boiler installation path

### Synopsis

Display all Boiler installation paths.

Shows:
- Root - Main Boiler directory
- Store - Where resources are stored
- Snippets - Snippet storage location
- Stacks - Stack storage location
- Logs - Log file directory
- Bin - Executable location

```
bl path [flags]
```

### Examples

```
  # Show all paths
  bl path

  # Use in scripts
  cd $(bl path | grep Store | cut -d: -f2)
```

### Options

```
  -h, --help   help for path
```

### Options inherited from parent commands

```
      --global    Force global scope for this command
      --local     Force local scope for this command
  -V, --verbose   Enable verbose debug output
```

