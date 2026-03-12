---
title: bl init
description: Command reference for bl init
---

Initialize stack config in current directory

### Synopsis

Initialize a boiler configuration file in the current directory.

For stacks (directories): Creates boiler.stack.json
For snippets (files): Creates boiler.snippet.json with metadata

Stack config includes:
  - Stack name and description
  - Author information
  - Files/folders to ignore
  - Version metadata

Snippet config includes:
  - Name, description, author
  - Language and tags
  - Version for templating

Similar to 'npm init', this helps you prepare projects for storing.

```
bl init [flags]
```

### Examples

```
  # Interactive init (prompts for details)
  bl init

  # Quick init with defaults (stack)
  bl init -y

  # Initialize as stack explicitly
  bl init -k
  bl init -k -y

  # Initialize as snippet
  bl init -n
  bl init -n -y

  # After init, customize and store
  bl store
```

### Options

```
  -h, --help      help for init
  -n, --snippet   Initialize as snippet
  -k, --stack     Initialize as stack
  -y, --yes       Skip prompts and use defaults
```

