# Contributing to Boiler

> **Discuss before you code.**
> Open an issue or comment on an existing one before starting work on features or significant changes. This keeps everyone aligned and avoids wasted effort. Bug fixes and documentation improvements can go straight to a PR.

---

## Current State

Boiler is actively evolving toward v1.0. The codebase is being improved continuously - bugs get fixed, APIs get refined, and architecture decisions are still being made. This means:

- **Breaking changes are possible** between minor versions (we're on v0.x.x)
- **Architecture is open to change** - if you see a better way, open an issue and let's discuss it
- **Test coverage is growing** - new contributions should ideally include tests
- **All feedback is welcome** - feature ideas, UX improvements, and bug reports directly shape the roadmap

---

## Quick Start

We use [mise](https://mise.jdx.dev) to manage dependencies, tools, and scripts.

```bash
# Clone the repository
git clone https://github.com/rishiyaduwanshi/boiler.git
cd boiler

# Install mise (if not already installed)
curl https://mise.run | sh

# Setup the project (installs Go, downloads deps, sets up git hooks)
mise run setup

# Build for development
mise run dev

# Run locally
./bl --help

# Run tests
mise run test
```

---

## Development Workflow

### 1. Fork & Branch
```bash
git clone https://github.com/rishiyaduwanshi/boiler.git
cd boiler
git checkout -b fix/your-fix-name
# or
git checkout -b feat/your-feature-name
```

### 2. Make Changes
- Write clean, idiomatic Go code
- Follow existing patterns in the package you're editing
- Add comments for non-obvious logic
- Update documentation alongside the code change if needed

### 3. Test Your Changes
```bash
mise run dev
./bl store ./testfile.js
./bl add testfile
./bl ls
```

### 4. Commit & Push
```bash
git add .
git commit -m "fix: clear error on invalid bl clean choice"
git push origin fix/your-fix-name
```

### 5. Open Pull Request
- Describe what changed and why
- Link the related issue (`Closes #55`)
- Try to keep PRs focused - one concern per PR

---

## Commit Message Convention

```
<type>: <short description>

[optional body explaining why, not what]
[optional footer: Closes #issue]
```

**Types:**
- `feat:`  - new feature or behavior
- `fix:` - bug fix
- `docs:` - documentation only
- `refactor:` - restructuring without behavior change
- `test:` - adding or fixing tests
- `chore:` - build, tooling, CI changes
- `security:` - security-related fix

**Examples:**
```
fix: return non-zero exit code on invalid bl clean choice
feat: auto-fill __author from git config when missing
refactor: replace addCmd globals with addOptions struct
security: implement bl self uninstall natively without curl
docs: update CONTRIBUTING with correct project structure
```

## Adding a New Command

Commands are isolated into their own packages. Follow the existing pattern:

### 1. Create `internal/cli/export/cmd.go`

```go
package export

import (
    "fmt"
    "os"

    "github.com/rishiyaduwanshi/boiler/internal/config"
    "github.com/rishiyaduwanshi/boiler/internal/utils"
    "github.com/spf13/cobra"
)

var (
    cfg    *config.Config
    logger *utils.Logger
)

func Setup(c *config.Config, l *utils.Logger) {
    cfg = c
    logger = l
}

var Cmd = &cobra.Command{
    Use:   "export [destination]",
    Short: "Export all snippets to a directory",
    Args:  cobra.MaximumNArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        if err := runExport(args); err != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
            os.Exit(1)
        }
    },
}

func runExport(args []string) error {
    logger.Info("exporting snippets")
    return nil
}

var exportForce bool

func init() {
    Cmd.Flags().BoolVarP(&exportForce, "force", "f", false, "Overwrite existing files")
}
```

### 2. Register in `internal/cli/root.go`

Add it to `Execute()` and `init()` in `root.go`:

```go
import exportcmd "github.com/rishiyaduwanshi/boiler/internal/cli/export"

func Execute(...) {
    // ...
    exportcmd.Setup(cfg, logger)
}

func init() {
    // ...
    rootCmd.AddCommand(exportcmd.Cmd)
}
```

### 3. Auto-generate docs

Just run `mise build` or `lefthook run pre-commit`. The `gendocs.go` tool will automatically generate `web/src/content/docs/commands/export.md` for you!

---

## Code Style

- Formatting and linting happen automatically on `git commit` via git hooks.
- Keep functions small and focused
- Handle every error - never silently discard `err`
- Use the shared message constants in `utils/messages.go` for user-facing strings

---

## Areas to Contribute

### 🐛 Bug Fixes
Check [open issues](https://github.com/rishiyaduwanshi/boiler/issues?q=is%3Aissue+is%3Aopen+label%3Abug). Issues labeled `good first issue` are a great starting point.

### 🔒 Security
Found a vulnerability? See [SECURITY.md](SECURITY.md) before opening a public issue.

### 📝 Documentation
Docs live in `web/src/content/docs/`. Every command change should come with a matching doc update.

### 🧪 Testing
Tests live next to the code they test (`store_test.go`, `config_test.go`, etc.). New behavior should include tests.

### 💡 Ideas & Discussion
Not sure if something is worth building? Open a [Discussion](https://github.com/rishiyaduwanshi/boiler/discussions) first - architecture and design decisions are made collaboratively.

---

## Pull Request Checklist

- [ ] Code builds: `mise run build`
- [ ] Tests pass: `mise run test`
- [ ] `mise run lint` is clean
- [ ] Related issue linked (`Closes #N`)
- [ ] Docs updated if behavior changed
- [ ] Commit messages follow the convention above

---

## Running Locally

Our cross-platform build script automatically detects your OS and builds the correct binary (`bl` or `bl.exe`).

```bash
# Development build (unoptimized, good for debugging)
mise run dev

# Standard build (optimized binary)
mise run build

# Release build (injects Git version tag)
mise run release
```

## Documentation Website

```bash
cd web
npm install
npm run dev        # http://localhost:4321
npm run build      # production build
```

---

## Questions?

- 💬 [GitHub Discussions](https://github.com/rishiyaduwanshi/boiler/discussions)
- 🐛 [Open an Issue](https://github.com/rishiyaduwanshi/boiler/issues)
- 📖 [Documentation](https://boiler.iamabhinav.dev)

---

By contributing, you agree that your contributions will be licensed under the [Apache License 2.0](LICENSE).
