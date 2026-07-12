---
title: Use Cases
description: Real-world scenarios showing what you can actually do with Boiler
---

Boiler is a **code automation engine** - not just a snippet manager. Here's what developers actually do with it.

---

## 1. Instant `.gitignore` for Any Language or Framework

Boiler ships with a built-in `gi` alias wired to the official [github/gitignore](https://github.com/github/gitignore) repository - **225+ popular templates** and **300+ community templates** (version-specific, specialized tools, editors, OS).

```bash
# Node.js
bl gi Node

# Python
bl gi Python

# Go
bl gi Go

# macOS (global)
bl gi Global/macOS

# JetBrains IDEs (community)
bl gi community/JetBrains+all
```

**That's it.** No copying from gitignore.io, no searching GitHub manually. The `.gitignore` appears instantly in your current directory.

> The `gi` alias works by combining **Alias + Variable** power under the hood:
> ```
> add github/gitignore:bl__1.gitignore . -m .gitignore -r
> ```
> `bl__1` is replaced with your argument - that's the entire magic.

---

## 2. Pull Any File or Repo Without `git clone`

Boiler can fetch **individual files**, **specific folders**, or **entire repositories** from GitHub, GitLab, or Bitbucket - without ever running `git clone` or downloading a ZIP manually.

```bash
# Grab a single file from a GitHub repo
bl use alice/snippets:js/errorHandler.js ./src/utils

# Grab a specific subfolder as a template
bl use vercel/next.js:examples/blog ./new-blog

# Grab an entire repo
bl use rishiyaduwanshi/express-starter ./my-api

# From GitLab
bl use https://gitlab.com/myorg/templates:docker/node.Dockerfile .

# From any direct URL
bl use https://raw.githubusercontent.com/user/repo/main/Dockerfile .
```

The resource lands exactly where you point it, fully processed. No extra cleanup needed.

---

## 3. Full Feature Scaffolding in One Command

Using `.bl` scripts with the `bl new` command, you can scaffold an entire feature - route, controller, model - and automatically wire them into your existing files, all in a single command.

**Example:** `bl new feat user`

This runs a script (`bl/commands/feat.bl`) that:
1. Creates `src/routes/user.route.js` (from a template)
2. Creates `src/controllers/user.controller.js` (from a template)
3. **Injects** `import userRouter from './routes/user.route.js'` into `src/api.js`
4. **Injects** `app.use('/user', userRouter)` into `src/api.js`

```bash
# feat.bl script
__desc = "Scaffold a new API feature"
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

**Result:** Your feature is fully scaffolded and connected - zero manual wiring.

---

## 4. Team-Wide Standards via Shared Registry

Your team's approved patterns - auth middleware, error handlers, configs - stored once in a shared GitHub repo, accessible by every developer on every machine.

```bash
# One-time setup (per developer)
bl conf --set-registry https://github.com/myteam/boiler-snippets

# Any developer pulls approved patterns
bl add jwtAuth -r          # Saves to local store, reusable offline
bl add rateLimiter -r
bl add mongoConfig -r

# Or one-shot (no local caching)
bl use myteam/boiler-snippets:middleware/auth.js ./src/middleware
```

**Benefit:** Every service uses the *exact same* battle-tested code. No more copy-paste drift between microservices.

---

## 5. Local Project Snippets & Automation

For project-specific templates you don't want in your global store, use **Local Scoped Stores**.

```bash
# Initialize local config (sets scope: local)
bl init -c
```

This creates a `boiler.local.json` at your project root with `"scope": "local"`. Now Boiler uses `bl/` directory instead of `~/.boiler/`:

```
my-project/
├── boiler.local.json      ← scope: local
└── bl/
    ├── commands/
    │   └── cmp.bl         ← bl new cmp <name>
    └── snippets/
        └── component.jsx  ← local template
```

```bash
# All commands automatically use local store
bl new cmp Button
# → Scaffolds src/components/Button.jsx using bl/snippets/component.jsx
```

---

## 6. Reuse Without Reinstalling Packages

Before adding a 1.4MB package for one utility function, check your store:

```bash
# Store once (anywhere you write it)
bl store utils/debounce.js
bl store utils/deepClone.js
bl store utils/formatDate.js

# Reuse in any project, no npm needed
bl add debounce ./src/utils
bl add deepClone ./src/utils
```

```javascript
// debounce.js - your template
// __var bl__WAIT = 300
function debounce(fn, wait = bl__WAIT) {
  let t;
  return (...args) => { clearTimeout(t); t = setTimeout(() => fn(...args), wait); };
}
export default debounce;
```

---

## 7. Config Files - Same Setup Every Project

```bash
# Store your preferred configs
bl store .eslintrc.json
bl store .prettierrc
bl store tsconfig.json
bl store Dockerfile

# New project setup in seconds
bl add eslintrc .
bl add prettierrc .
bl add tsconfig .
bl add Dockerfile .
```

---

## 8. Versioned Code Evolution

Store multiple versions, use the right one per project:

```bash
# Store v1
bl store utils/emailService.js
# → Saved as emailService@1.js

# Improve it later, update __version to 2
bl store utils/emailService.js
# → Saved as emailService@2.js

bl add emailService@1    # Legacy projects
bl add emailService@2    # New projects
```

---

## 9. DevOps Scripts & CI Templates

```bash
# Store your deployment scripts
bl store scripts/deploy.sh
bl store scripts/backup.sh
bl store .github/workflows/ci.yml

# Reuse across services
bl add deploy ./scripts
bl add ci .github/workflows
```

---

## Next Steps

- [Project Structure](/guides/project-structure/) - How `bl/` and `boiler/` work together
- [Boiler Scripts (.bl)](/guides/bl-scripts/) - Write full automation pipelines
- [Remote Fetching](/guides/remote-fetching/) - Fetch from anywhere
- [Commands Reference](/commands/bl/) - Complete reference
