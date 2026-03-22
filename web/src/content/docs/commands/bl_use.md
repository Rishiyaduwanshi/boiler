---
title: bl use
description: Command reference for bl use
---

Fetch a remote resource directly without saving to local store

### Synopsis

Fetch a remote resource directly without saving to local store.

The resource is fetched from remote source and copied directly to destination
without writing into local store metadata.

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

  # Force overwrite destination conflicts
  bl use alice/my-stack ./new-project --force
```

### Options

```
  -f, --force   Force operation without confirmation
  -h, --help    help for use
```

### Options inherited from parent commands

```
  -V, --verbose   Enable verbose debug output
```

