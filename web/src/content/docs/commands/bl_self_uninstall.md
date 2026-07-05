---
title: bl self uninstall
description: Command reference for bl self uninstall
---

Uninstall Boiler CLI

### Synopsis

Uninstall Boiler CLI from your system.

This will:
- Locate and remove the Boiler binary
- Prompt for confirmation before deletion

After removal, you will need to manually clean the PATH entry
added by the installer (remove it from ~/.bashrc, ~/.zshrc, or
Windows System Environment Variables).

```
bl self uninstall [flags]
```

### Examples

```
  # Uninstall Boiler
  bl self uninstall
```

### Options

```
  -h, --help   help for uninstall
```

### Options inherited from parent commands

```
  -V, --verbose   Enable verbose debug output
```

