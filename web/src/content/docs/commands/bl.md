---
title: bl
description: Command reference for bl
---

Boiler - Code snippet and stack manager

### Synopsis

Boiler - A CLI tool to manage reusable code snippets and project stacks.

Store, version, and reuse your code across projects. Perfect for:
- Reusable utility functions (snippets)
- Project templates and boilerplates (stacks)
- Code patterns you use frequently
- Multi-language development workflows

All resources are versioned automatically, making it easy to manage multiple
variations of the same snippet or stack.

```
bl [flags]
```

### Examples

```
  # Initialize Boiler
  bl init

  # Store a snippet
  bl store ./utils/logger.js

  # Add snippet to project
  bl add logger

  # List all resources
  bl ls

  # Show paths
  bl path
```

### Options

```
  -h, --help      help for bl
  -V, --verbose   Enable verbose debug output
  -v, --version   Show version information
```

