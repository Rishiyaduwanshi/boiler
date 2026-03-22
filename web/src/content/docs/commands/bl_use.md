---
title: bl use
description: Command reference for bl use
---

Fetch a remote resource directly without saving to local store

### Synopsis

Fetch a snippet or stack from any remote source and place it in the
current directory - no local store involved, no registry lookup needed.

Unlike 'bl add -r', this command is purely one-shot:
- Downloads the resource
- Copies it to the destination
- Does NOT save it to your local store

Provider is auto-detected from the URL (GitHub, GitLab, Bitbucket, generic).
Both .zip and .tar.gz archives are supported and auto-detected from the URL.

Supported formats:
- owner/repo                             GitHub repo as stack (default branch)
- owner/repo:path/to/file.js            File from GitHub repo
- `https://github.com/owner/repo`         GitHub full URL
- `https://gitlab.com/owner/repo`         GitLab repo
- `https://bitbucket.org/owner/repo`      Bitbucket repo
- `https://anysite.com/stack.zip`         Direct zip archive
- `https://anysite.com/stack.tar.gz`      Direct tar.gz archive
- `https://anysite.com/file.js`           Direct file (snippet)

```
bl use [resource] [flags]
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
  bl use alice/my-stack --to ./new-project
```

### Options

```
  -h, --help        help for use
  -t, --to string   Destination path (default ".")
```

