<div align="center">

<img src="web/src/assets/logo.svg" alt="Boiler Logo" width="120" />

# Boiler

**Build Less. Ship More.**

Fetch code from anywhere. Scaffold full features in one command. Reuse everything across every project.

[![License: Apache-2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://golang.org)
[![Release](https://img.shields.io/github/v/release/rishiyaduwanshi/boiler)](https://github.com/rishiyaduwanshi/boiler/releases)
[![Sponsor](https://img.shields.io/badge/Sponsor-Boiler-2ea043?logo=githubsponsors&logoColor=white)](https://github.com/sponsors/rishiyaduwanshi)
[![Use Cases](https://img.shields.io/badge/Use%20Cases-Explore-f97316)](https://boiler.iamabhinav.dev/guides/usecases/)

[Documentation](https://boiler.iamabhinav.dev) • [Use Cases](https://boiler.iamabhinav.dev/guides/usecases/) • [Guides](https://boiler.iamabhinav.dev/guides/introduction/)

</div>

**Pre-release - v0.x.x** - Feel free to use it, break it. Every idea and bug report shapes what we build next.

> Open an [issue](https://github.com/rishiyaduwanshi/boiler/issues) or start a [discussion](https://github.com/rishiyaduwanshi/boiler/discussions) anytime.

---

## Instant `.gitignore` - 500+ Templates, One Command

```bash
bl gi Node        # Node.js
bl gi Python      # Python
bl gi Go          # Go
bl gi Global/macOS  # macOS global
bl gi community/JetBrains+all  # JetBrains IDEs
```

No browser, no copy-paste. Powered by the official [github/gitignore](https://github.com/github/gitignore) repo with 225+ popular templates and 300+ community templates.

---

## Pull Files From GitHub - Without `git clone`

```bash
# Grab a single file from any public repo
bl use alice/snippets:js/errorHandler.js ./src/utils

# Grab a specific subfolder as a project template
bl use vercel/next.js:examples/blog ./my-blog

# Entire repo (no clone, no zip download)
bl use rishiyaduwanshi/express-starter ./my-api

# From GitLab or any direct URL
bl use https://gitlab.com/myorg/templates:Dockerfile .
```

---

## Scaffold Full Features in One Command

Write a `.bl` script that creates routes, controllers, and auto-wires them into your existing codebase:

```bash
bl new feat user
# ✓ Created src/routes/user.route.js
# ✓ Created src/controllers/user.controller.js
# ✓ Injected import into src/api.js
# ✓ Injected app.use('/user', userRouter) into src/api.js
```

```bash
# bl/commands/feat.bl
__var bl__feature = bl__1.lowercase()
__var bl__Feature = bl__1.capitalize()

add route.js ./src/routes/bl__feature.route.js --local
add controller.js ./src/controllers/bl__feature.controller.js --local

inject ./src/api.js -d imports -a -c `
import bl__featureRouter from './routes/bl__feature.route.js';
`
inject ./src/api.js -d routes -a -c `
app.use('/bl__feature', bl__featureRouter);
`
```

---

## Store Once. Reuse Everywhere.

```bash
# Store a snippet with template variables
bl store utils/dbConnect.js
# ✓ Stored snippet 'dbConnect@1.js'

# Reuse in any project
cd ~/new-service
bl add dbConnect ./src
#   bl__DB_URL [mongodb://localhost:27017]: ...
# ✓ Added snippet → ./src/dbConnect.js
```

Variables replaced. Metadata stripped. Clean code, ready to use.

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

## Features

- ✅ **Remote Fetching** - GitHub, GitLab, Bitbucket, direct URLs, custom registries
- ✅ **Zero-Clone** - Pull individual files or entire repos without `git clone`
- ✅ **Automation Scripts** - `.bl` scripts scaffold multiple files and inject code
- ✅ **Template Variables** - `bl__VAR_NAME` syntax with interactive prompts
- ✅ **Built-in `gi` Alias** - 500+ gitignore templates via `bl gi <language>`
- ✅ **Aliases + Vars** - Build your own one-word power commands
- ✅ **Local & Global Stores** - Per-project `bl/` directory or global `~/.boiler`
- ✅ **Automatic Versioning** - `@1`, `@2`, etc. per resource
- ✅ **Language Agnostic** - JS, Python, Go, Rust, Java, Docker, configs - anything
- ✅ **Framework Agnostic** - Express, Next.js, Django, Spring, whatever you use
- ✅ **OS Agnostic** - Windows, Linux, macOS
- ✅ **Self-Updating** - `bl self update`

---

## Commands

```bash
bl store <file>              # Store a file as snippet
bl store <folder>            # Store a directory as stack (bl init required first)
bl add <name> [path]         # Add snippet or stack (default: ./boiler)
bl add <name> -r             # Fetch from remote and save to local store
bl use <url-or-repo> [path]  # One-shot fetch - no local caching
bl gi <language>             # Instant gitignore (bl gi Node, bl gi Python, etc.)
bl new <script> [args]       # Run a .bl automation script
bl ls                        # List all stored resources
bl search <query> [-r]       # Search local or remote store
bl info <name>               # Show resource details
bl alias [k=v]               # Create a command shortcut
bl var [k=v]                 # Set a reusable variable
bl init                      # Initialize stack/snippet config
bl conf                      # View or edit configuration
bl clean                     # Remove resources from store
bl path                      # Show installation paths
bl self update               # Update Boiler to latest
bl self uninstall            # Uninstall Boiler
bl version                   # Show version
```

**Full docs:** [boiler.iamabhinav.dev](https://boiler.iamabhinav.dev)

---

## Project-Local Store

For project-specific templates that shouldn't be in your global store:

```bash
bl init -c          # Creates boiler.local.json with scope: local
```

```
my-project/
├── boiler.local.json          ← scope: local
└── bl/
    ├── commands/
    │   └── cmp.bl             ← bl new cmp <ComponentName>
    └── snippets/
        └── component.jsx      ← local template
```

All `bl` commands automatically use `bl/` instead of `~/.boiler/` when `scope: local` is set.

---

## Team Registry

Host a shared snippet library on GitHub. Any developer pulls standardized patterns in seconds:

```bash
# One-time setup
bl conf --set-registry https://github.com/myteam/boiler-snippets

# Pull approved patterns (cached locally for offline use)
bl add jwtAuth -r
bl add rateLimiter -r
bl add mongoConfig -r
```

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