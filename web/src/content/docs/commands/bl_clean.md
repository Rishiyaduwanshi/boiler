---
title: bl clean
description: Command reference for bl clean
---

Clean snippets, stacks, or store

### Synopsis

Remove snippets, stacks, or clear entire store.

You can:
- Remove specific resource by name
- Remove all snippets (use -s or --snippets flag)
- Remove all stacks (use -k or --stacks flag)
- Clear everything (use -a or --all flag)

Version-specific deletion is supported.

```
bl clean [resource] [flags]
```

### Examples

```
  # Remove specific snippet
  bl clean errorHandler@1.js

  # Remove specific stack
  bl clean express-api@1

  # Remove all snippets
  bl clean --snippets

  # Remove all stacks
  bl clean --stacks

  # Clear entire store
  bl clean --all
```

### Options

```
  -a, --all        Clean all resources
  -h, --help       help for clean
  -n, --snippets   Snippets only
  -k, --stacks     Stacks only
```

### Options inherited from parent commands

```
  -V, --verbose   Enable verbose debug output
```

