---
title: bl store
description: Command reference for bl store
---

Store a folder/file as snippet or stack

### Synopsis

Store a file as a snippet or directory as a stack in your Boiler store.

Files are stored as snippets with version numbers.
Directories must have a boiler.stack.json config file (run 'bl init' first).

Version Management:
If snippet already exists, you'll be prompted with options:
- (o) Overwrite - Replace the latest version with new content
- (n) New version - Create a new incremental version
- (c) Cancel - Abort the operation
- First-time storage automatically creates version 1

Stacks require boiler.stack.json with:
- id: Stack name
- version: Version number
- ignore: Patterns to exclude

If a stack version already exists, you'll be prompted to overwrite.

```
bl store [path] [flags]
```

### Examples

```
  # Store current directory as stack
  bl store

  # Store specific file as snippet (first version)
  bl store ./utils/logger.js
  # Output: ✓ Stored snippet 'logger@1.js'

  # Store again - prompts for action
  bl store ./utils/logger.js
  # Prompt: Snippet 'logger' already exists (1 version(s)). Options:
  #   (o) Overwrite latest version (1)
  #   (n) Create new version (2)
  #   (c) Cancel

  # Store directory as stack
  bl store ./my-template

  # Store with custom name
  bl store ./config.js --name dbConfig.js
```

### Options

```
  -h, --help          help for store
      --name string   Name for the resource (auto-detected from path if not provided)
  -n, --snippet       Force store as snippet
  -k, --stack         Force store as stack
```

