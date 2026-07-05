<div align="center">

<img src="web/src/assets/logo.svg" alt="Boiler Logo" width="120" />

# Boiler

**Code Once. Reuse Forever.**

Store reusable code snippets and project templates locally. Automatic versioning, template variables, zero config.

[![License: Apache-2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://golang.org)
[![Release](https://img.shields.io/github/v/release/rishiyaduwanshi/boiler)](https://github.com/rishiyaduwanshi/boiler/releases)
[![Sponsor](https://img.shields.io/badge/Sponsor-Boiler-2ea043?logo=githubsponsors&logoColor=white)](https://github.com/sponsors/rishiyaduwanshi)
[![Use Cases](https://img.shields.io/badge/Use%20Cases-Explore-f97316)](https://boiler.iamabhinav.dev/guides/usecases/)

[Documentation](https://boiler.iamabhinav.dev) • [Use Cases](https://boiler.iamabhinav.dev/guides/usecases/)

</div>

**Pre-release - v0.x.x** - Feel free to use it, break it, Every idea and bug report shapes what we build next.

> Open an [issue](https://github.com/rishiyaduwanshi/boiler/issues) or start a [discussion](https://github.com/rishiyaduwanshi/boiler/discussions) anytime.

---

## Why Boiler?

Stop installing entire packages for one function. Stop copy-pasting code between projects.

```bash
# Store once
bl store ./utils/errorHandler.js

# Reuse anywhere
bl add errorHandler
```

**Perfect for:** Utility functions, config files, project boilerplates, API templates.

---

## Installation

**Windows:**
```powershell
iwr -useb https://boiler.iamabhinav.dev/install | iex
```

**Linux / macOS:**
```bash
curl -fsSL https://boiler.iamabhinav.dev/install | bash
```

**npm (all platforms):**
```bash
npm install -g @boilercli/core
```

**Verify:**
```bash
bl version
```

---

## Quick Start

```bash
# 1. Store a file as snippet
bl store ./middleware/auth.js
# ✓ Stored snippet 'auth@1.js' at ~/.boiler/store/...

# 2. Add to any project
bl add auth
# ✓ Added snippet 'auth@1.js' to boiler/auth.js

# 3. Store a directory as stack (bl init required first)
cd ./my-template
bl init        # creates boiler.stack.json
bl store       # stores it

# 4. Add a stack
bl add my-template
# ✓ Added stack 'my-template@1' to boiler/my-template

# 5. List everything
bl ls
```

---

## Template Variables

Create configurable snippets:

```js
// errorHandler.js
// __var bl__LOG_LEVEL = error
// __var bl__EMAIL = admin@example.com

function handleError(err) {
  console[bl__LOG_LEVEL](err.message);
  sendEmail('bl__EMAIL', err);
}
```

```bash
bl add errorHandler
#   bl__LOG_LEVEL [error]: warn
#   bl__EMAIL [admin@example.com]: dev@app.com
# ✓ Added snippet 'errorHandler@1.js' to boiler/errorHandler.js
# (metadata stripped, variables replaced)
```

---

## Features

- ✅ **Automatic Versioning** - `@1`, `@2`, etc.
- ✅ **Template Variables** - `bl__VAR_NAME` syntax with prompts
- ✅ **Reusable Config Vars** - `bl var KEY=value` and inline `bl__KEY` references in commands
- ✅ **Powerful Aliases** - Support for positional arguments (`bl__1`, `bl__2`) and inline variables
- ✅ **Remote Fetching** - GitHub, GitLab, Bitbucket, custom registries, direct URLs
- ✅ **Language Agnostic** - JS, Python, Go, Java, TS, Rust, C++, etc.
- ✅ **Stack Templates** - Store entire project folders
- ✅ **Zero Config** - Works immediately after install
- ✅ **Cross-Platform** - Windows, Linux, macOS
- ✅ **Self-Updating** - `bl self update`

---

## Remote Fetching

Pull snippets and stacks from anywhere — provider is auto-detected from the URL:

```bash
# GitHub (short format)
bl add rishiyaduwanshi/boiler-express -r

# GitHub (full URL)
bl add https://github.com/alice/my-stack -r

# GitLab
bl add https://gitlab.com/alice/my-stack -r

# Bitbucket
bl add https://bitbucket.org/alice/my-stack -r

# File from GitHub repo (snippet)
bl add alice/snippets:js/errorHandler.js -r

# Direct archive URL
bl add https://mysite.com/template.zip -r

# From configured registry
bl conf --set-registry https://github.com/myteam/boiler
bl add express@1 -r
```

**One-shot fetch (no local store):**
```bash
bl use alice/my-stack
bl use https://gitlab.com/alice/my-stack
bl use https://mysite.com/template.zip
```

---

## Commands

```bash
bl init              # Initialize stack/snippet config
bl store [path]      # Store file/folder
bl add <name> [path] # Add snippet/stack (default destination: ./boiler)
bl add <name> --spread # Spread stack contents into destination
bl alias [k|k=v]     # List, get, or set command aliases (supports bl__N args)
bl unalias <k>       # Remove a command alias
bl var [k|k=v]       # List, get, or set reusable variables
bl unvar <k>         # Remove a reusable variable
bl use <resource> [path]  # One-shot remote fetch (e.g., bl use bl__ORG/repo)
bl ls                # List all resources
bl search <query>    # Search by name (-r to search remote)
bl info <name>       # Show resource details
bl clean             # Remove resources
bl path              # Show installation paths
bl conf              # View/edit configuration
bl self update       # Update Boiler to latest
bl self uninstall    # Uninstall Boiler
bl version           # Show version
bl --help            # Full command list
```

**Full docs:** [boiler.iamabhinav.dev](https://boiler.iamabhinav.dev)

---

## Contributing

```bash
git clone https://github.com/rishiyaduwanshi/boiler.git
cd boiler
mise run setup   # Install deps and hooks
mise run dev     # Build development binary
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.
For security reporting, see [SECURITY.md](SECURITY.md).

---

## License

Apache-2.0 © [Abhinav Prakash](https://github.com/rishiyaduwanshi)

---

<div align="center">

[⭐ Star on GitHub](https://github.com/rishiyaduwanshi/boiler) • [📖 Docs](https://boiler.iamabhinav.dev)

</div>