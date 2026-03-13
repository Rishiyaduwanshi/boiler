---
title: Quick Start
description: Get started with Boiler in 5 minutes
---

Learn the basics of Boiler in just a few minutes.

## 1. Store Your First Snippet

Create and store a simple error handler:

```bash
# Create a file
echo "function handleError(err) { console.error(err); }" > errorHandler.js

# Store it
bl store errorHandler.js
```

Output: `✓ Snippet stored: errorHandler@1.js`

## 2. List Your Snippets

View all your stored snippets:

```bash
bl ls --snippets
```

Output:
```
📄 Snippets:
  • errorHandler@1.js

📦 Stacks:
  No stacks found
```

## 3. Use the Snippet in Another Project

Navigate to a different directory and add the snippet:

```bash
cd ../my-other-project
bl add errorHandler
```

Output: `✓ Snippet added: errorHandler@1.js → boiler/errorHandler.js`

The snippet is copied to the default `boiler/` directory.

## 4. Store a Project Stack

Store an entire project directory as a stack:

```bash
# First, initialize stack config (required before storing a directory)
cd my-express-app
bl init
# Prompts for stack name, description, author, and files to ignore

# Then store it
bl store
```

Output: `✓ Stored stack 'express-starter@1' at /path/to/store/stacks/express-starter@1`

## 5. Initialize a New Project from Stack

Start a new project using your stack:

```bash
mkdir new-project
cd new-project
bl add express-starter
```

Output: `✓ Stack added: express-starter@1 → boiler/express-starter`

Your stack is copied to `./boiler/express-starter` by default.

## Common Commands

### Store Resources
```bash
bl store <file>              # Store a snippet
bl store <folder> --stack    # Store a stack
```

### Add Resources
```bash
bl add <name>                # Add snippet/stack into ./boiler
bl add <name> <path>         # Add to a specific relative/absolute path
bl add <name> --spread       # Spread stack contents into destination
bl add <name@version.ext>    # Add specific version
bl add <name> -r             # Fetch from remote registry and save locally
bl use <url-or-repo>         # One-shot fetch from anywhere (no local store)
```

### List Resources
```bash
bl ls                        # List all
bl ls --snippets             # List only snippets
bl ls --stacks               # List only stacks
```

### Get Information
```bash
bl info <name>               # Show resource details
bl path                      # Show store location
```

### Search and Clean
```bash
bl search <query>            # Search by name
bl clean <name>              # Remove a resource
```

## What You've Learned

- ✓ Store files as snippets with `bl store`
- ✓ Store directories as stacks with `bl store --stack`
- ✓ List resources with `bl ls`
- ✓ Add resources with `bl add` (auto-detects single versions)
- ✓ Get info with `bl info`

You're now ready to use Boiler in your daily workflow!


