---
title: bl use
description: Command reference for bl use
---

Fetch a remote resource directly without saving to local store

### Synopsis

Fetch a remote resource directly without saving to local store.

The resource is fetched from remote source and copied directly to destination
without writing into local store metadata.

Stack placement:
- By default, stacks are copied inside a stack-named folder.
- Use --spread to copy stack contents directly into destination.

```
bl use [resource] [destination] [flags]
```

### Examples

```
  # GitHub repo as stack
  bl use alice/my-express-stack

  # GitLab repo
  bl use https://gitlab.com/alice/my-stack

  # Bitbucket repo
  bl use https://bitbucket.org/alice/my-stack

  # File from GitHub repo (snippet)
  bl use alice/snippets:js/errorHandler.js

  # Direct zip archive (any site)
  bl use https://mysite.com/templates/express.zip

  # Direct tar.gz archive
  bl use https://mysite.com/stack.tar.gz

  # Direct file URL (snippet)
  bl use https://mysite.com/snippets/logger.js

  # Resource from config variable
  bl use :starter_stack

  # Into a specific folder
  bl use alice/my-stack ./new-project

  # Spread stack contents directly into destination
  bl use alice/my-stack ./new-project --spread

  # Force overwrite destination conflicts
  bl use alice/my-stack ./new-project --force
```

### Options

```
  -f, --force         Force operation without confirmation
  -h, --help          help for use
  -m, --name string   Rename snippet in destination
  -n, --snippet       Treat resource as snippet (overrides auto-detection)
      --spread        Spread stack contents directly into destination
  -k, --stack         Treat resource as stack (overrides auto-detection)
```

### Options inherited from parent commands

```
      --global    Force global scope for this command
      --local     Force local scope for this command
  -V, --verbose   Enable verbose debug output
```

