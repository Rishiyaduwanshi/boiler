---
title: bl add
description: Command reference for bl add
---

Add a snippet or stack to current directory

### Synopsis

Add a stored snippet or stack to your current directory.

The command copies resources from your store. For snippets with a single version,
you can use just the name (e.g., 'errorHandler' will auto-select version 1).
For multiple versions, you'll be prompted to choose.

Template Variables:
  Snippets can contain template variables using the format: bl__VAR_NAME
  When adding a snippet with variables, you'll be prompted to provide values:
    - Default values are shown in brackets (from __var declarations)
    - Press Enter to use default or type a custom value
    - Variables are replaced and metadata comments are removed in the final file

Stacks are also versioned and can be added by name or with explicit version.

Remote Resources:
  Use -r flag to fetch from remote source and save to local store.
  Provider is auto-detected from the URL (GitHub, GitLab, Bitbucket, generic).
  Resource is cached locally — subsequent uses don't need -r.

  For one-shot fetch without saving to local store, use 'bl use' instead.

    1. Registry:           bl add express@1 -r
       (registry set via: bl conf --set-registry <url>)

    2. GitHub short:       bl add owner/repo -r
    3. GitHub full URL:    bl add https://github.com/owner/repo -r
    4. GitLab:             bl add https://gitlab.com/owner/repo -r
    5. Bitbucket:          bl add https://bitbucket.org/owner/repo -r

    6. File from repo:     bl add owner/repo:path/to/file.js -r
    7. Direct file URL:    bl add https://site.com/file.js -r
    8. Direct archive:     bl add https://site.com/stack.zip -r
    9. Custom domain file: bl add site.com:path/file.js -r

   10. One-time registry:  bl add express@1 -r --registry https://github.com/other/boiler

```
bl add [resource] [flags]
```

### Examples

```
  # Add snippet (auto-detects if single version)
  bl add errorHandler

  # Add snippet with template variables
  bl add apiClient
  # Prompts: bl__API_URL [http://localhost:3000]: https://api.example.com
  #          bl__API_KEY [your-key]: abc123xyz
  # Output: Clean file with variables replaced, no metadata comments

  # Add specific version
  bl add logger@2.js

  # Add to specific directory
  bl add config --to ./src/utils

  # Add stack
  bl add express-api@1

  # Force overwrite
  bl add middleware --force

  # Remote: from configured registry
  bl add express@1 -r

  # Remote: GitHub short format
  bl add rishiyaduwanshi/boiler-express -r

  # Remote: GitLab
  bl add https://gitlab.com/alice/my-stack -r

  # Remote: Bitbucket
  bl add https://bitbucket.org/alice/my-stack -r

  # Remote: file inside GitHub repo
  bl add rishiyaduwanshi/boiler-snippets:js/errorHandler.js -r

  # Remote: direct file URL
  bl add https://mysite.com/snippets/helper.js -r

  # Remote: direct archive URL
  bl add https://mysite.com/stack.zip -r

  # Remote: one-time registry override
  bl add express@1 -r --registry https://github.com/myorg/boiler

  # One-shot fetch without saving to store (no -r needed)
  bl use alice/my-stack
```

### Options

```
  -f, --force             Force operation without confirmation
  -h, --help              help for add
      --registry string   Custom registry URL (overrides config)
  -r, --remote            Fetch from remote registry
  -t, --to string         Destination path (default ".")
```

