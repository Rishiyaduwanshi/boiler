<div align="center">

<img src="web/src/assets/logo.svg" alt="Boiler Logo" width="120" />

# Boiler

**Code Once. Reuse Forever.**

Store reusable code snippets and project templates locally. Automatic versioning, template variables, zero config.

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://golang.org)
[![Release](https://img.shields.io/github/v/release/rishiyaduwanshi/boiler)](https://github.com/rishiyaduwanshi/boiler/releases)

[Documentation](https://boiler.iamabhinav.dev) • [Use Cases](https://boiler.iamabhinav.dev/guides/usecases/)

</div>

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

**Verify:**
```bash
bl version
```

---

## Quick Start

```bash
# Initialize
bl init

# Store a file
bl store ./middleware/auth.js
# → Saved as auth@1.js

# Add to any project
bl add auth
# → Copied to current directory

# Fetch from remote
bl search express -r
bl add express@1 -r
# → Downloaded and initialized

# List all resources
bl ls

# Search locally
bl search auth
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
# ✓ Created (metadata stripped, variables replaced)
```

---

## Features

- ✅ **Automatic Versioning** - `@1`, `@2`, etc.
- ✅ **Template Variables** - `bl__VAR_NAME` syntax with prompts
- ✅ **Remote Fetching** - Pull snippets from GitHub or custom registries
- ✅ **Language Agnostic** - JS, Python, Go, Java, TS, Rust, C++, etc.
- ✅ **Stack Templates** - Store entire project folders
- ✅ **Zero Config** - Works immediately after install
- ✅ **Cross-Platform** - Windows, Linux, macOS
- ✅ **Self-Updating** - `bl self update`

---

## Remote Fetching

Pull snippets from GitHub repositories or custom registries:

```bash
# Set default registry
bl conf --set-registry https://github.com/rishiyaduwanshi/boiler

# Search remote resources
bl search express -r

# Add from registry
bl add express@1 -r

# Direct GitHub repo
bl add username/repo -r
bl add username/repo:path/to/file.js -r
```

---

## Commands

```bash
bl init              # Initialize Boiler
bl store [path]      # Store file/folder
bl add <name>        # Add snippet/stack (use -r for remote)
bl ls                # List all resources
bl search <query>    # Search by name (use -r for remote)
bl info <name>       # Show resource details
bl clean             # Remove unused versions
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
go build -o bl main.go
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

---

## License

MIT © [Abhinav Prakash](https://github.com/rishiyaduwanshi)

---

<div align="center">

[⭐ Star on GitHub](https://github.com/rishiyaduwanshi/boiler) • [📖 Docs](https://boiler.iamabhinav.dev)

</div>