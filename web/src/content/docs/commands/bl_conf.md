---
title: bl conf
description: Command reference for bl conf
---

Manage boiler configuration

### Synopsis

View and manage Boiler configuration.

You can:
- View current configuration (default)
- Edit config in default editor (use -e or --edit)
- Reset to defaults (use -r or --reset)
- Set registry URL (use --set-registry)

Configuration includes paths, preferences, and behavior settings.

```
bl conf [flags]
```

### Examples

```
  # Show configuration
  bl conf

  # Edit configuration (uses defaultEditor from config, or $EDITOR env)
  bl conf -e

  # Edit with a specific editor
  bl conf -e code
  bl conf --edit notepad
  bl conf -e vim

  # Reset to defaults
  bl conf --reset

  # Set custom registry
  bl conf --set-registry https://github.com/myorg/boiler

  # Set to default registry
  bl conf --set-registry https://github.com/rishiyaduwanshi/boiler
```

### Options

```
  -e, --edit string           Edit configuration (optional: editor name)
  -h, --help                  help for conf
  -r, --reset                 Reset configuration to defaults
      --set-registry string   Set custom registry URL
```

### Options inherited from parent commands

```
  -V, --verbose   Enable verbose debug output
```

