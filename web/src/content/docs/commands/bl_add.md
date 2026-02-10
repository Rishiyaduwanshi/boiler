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
  Use -r flag to fetch from remote registry or directly from GitHub/URLs:
    1. From registry: bl add express@1 -r
       (Uses registry configured in config or --registry flag)
    
    2. Direct from GitHub: bl add owner/repo -r
       (Fetches entire repo as stack)
    
    3. Direct snippet: bl add owner/repo:path/to/file.js -r
       (Fetches single file)
    
    4. Direct URL: bl add https://yourdomain.com/path/file.js -r
       (Downloads from any URL)
    
    5. Custom domain: bl add yourdomain.com:path/file.js -r
       (Assumes HTTPS)
    
    6. Custom registry: bl add express@1 -r --registry https://github.com/other/boiler
       (One-time registry override)

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

  # Remote: From registry
  bl add express@1 -r

  # Remote: Direct from GitHub (entire repo)
  bl add rishiyaduwanshi/boiler-express -r

  # Remote: Direct snippet from GitHub
  bl add rishiyaduwanshi/boiler-snippets:js/errorHandler.js -r

  # Remote: From custom website (direct URL)
  bl add https://iamabhinav.dev/snippets/helper.js -r

  # Remote: From custom domain
  bl add iamabhinav.dev:snippets/validator.js -r

  # Remote: Custom registry
  bl add express@1 -r --registry https://github.com/myorg/boiler
```

### Options

```
  -b, --both              Add to both local and global
  -f, --force             Force operation without confirmation
  -g, --global            Add to global store
  -h, --help              help for add
      --registry string   Custom registry URL (overrides config)
  -r, --remote            Fetch from remote registry
  -t, --to string         Destination path (default ".")
```

