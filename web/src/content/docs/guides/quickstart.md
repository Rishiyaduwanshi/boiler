---
title: Quick Start
description: Get Boiler running and do something useful in 5 minutes
---

## Your First Snippet

Create a file with template variables and store it:

```javascript
// errorHandler.js
// __author Your Name
// __desc Custom error handler
// __var bl__LOG_LEVEL = error

function handleError(err) {
  console[bl__LOG_LEVEL](err.message);
}

module.exports = handleError;
```

```bash
bl store errorHandler.js
# ✓ Snippet stored: errorHandler@1.js
```

Use it in any project:

```bash
cd ~/another-project
bl add errorHandler ./src/utils
#   bl__LOG_LEVEL [error]: warn
# ✓ Snippet added: errorHandler@1.js → ./src/utils/errorHandler.js
```

The metadata comments are stripped, `bl__LOG_LEVEL` is replaced with `warn`. Clean code, ready to use.

---

## Instant `.gitignore`

The `gi` alias ships by default - wired to 500+ gitignore templates:

```bash
bl gi Node           # Node.js
bl gi Python         # Python
bl gi Go             # Go
bl gi Global/macOS   # macOS global
```

---

## Fetch From GitHub Without `git clone`

```bash
# Single file from any public repo
bl use alice/snippets:utils/debounce.js ./src/utils

# Entire subfolder as a template
bl use vercel/next.js:examples/with-tailwindcss ./my-project

# Save to local store for reuse later
bl add alice/snippets:utils/debounce.js -r
```

---

## Store a Project Template (Stack)

```bash
cd my-express-template
bl init          # Creates boiler.stack.json (name, ignore patterns)
bl store         # Stores the entire directory as a stack
```

Start new projects from it:

```bash
mkdir new-api && cd new-api
bl add express-template --spread    # Spreads contents into ./
npm install
```

---

## Essential Commands

```bash
bl store <file>              # Store a file as snippet
bl store <folder>            # Store a folder as stack (requires bl init first)
bl add <name> [path]         # Add snippet or stack to a path (default: ./boiler)
bl add <name> -r             # Fetch from remote registry and save locally
bl use <url-or-repo> [path]  # One-shot fetch (no local caching)
bl gi <language>             # Instant gitignore for any language
bl ls                        # List all stored resources
bl search <query>            # Search by name (add -r to search remote)
bl info <name>               # Show details about a resource
bl new <script>              # Run a .bl automation script
bl alias [k=v]               # Create a command shortcut
bl var [k=v]                 # Set a reusable variable
bl conf                      # View or edit your configuration
bl clean                     # Remove resources from your store
bl self update               # Update Boiler to the latest version
```

---

## Next Steps

- [Use Cases](/guides/usecases/) - Real-world workflows and examples
- [Project Structure](/guides/project-structure/) - How `bl/` and `boiler/` work
- [Boiler Scripts (.bl)](/guides/bl-scripts/) - Write full automation pipelines
- [Remote Fetching](/guides/remote-fetching/) - Fetch from GitHub, GitLab, anywhere
