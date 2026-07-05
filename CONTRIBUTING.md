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

```bash
# Clone the repository
git clone https://github.com/rishiyaduwanshi/boiler.git
cd boiler

# Install dependencies
go mod download

# Build the project
go build -o bl main.go

# Run locally
./bl --help

# Run tests
go test ./...
```

---

## Development Workflow

### 1. Fork & Branch
```bash
git clone https://github.com/YOUR_USERNAME/boiler.git
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
go build -o bl main.go
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

---

## Project Structure

```
boiler/
├── main.go                        # Entry point - loads config, starts CLI
├── gendocs.go                     # Auto-generates CLI docs
├── go.mod / go.sum
│
├── internal/
│   ├── cli/                       # One file per command
│   │   ├── root.go               # Root command, Execute(), alias expansion
│   │   ├── add.go                # bl add
│   │   ├── store.go              # bl store
│   │   ├── clean.go              # bl clean
│   │   ├── search.go             # bl search
│   │   ├── init.go               # bl init
│   │   ├── self.go               # bl self update / uninstall
│   │   └── ...
│   ├── config/
│   │   └── config.go             # Load/save boiler.conf.json
│   ├── models/                    # Shared data types (StackConfig, etc.)
│   ├── remote/                    # Remote fetch logic
│   │   ├── fetch.go              # FetchSnippet, FetchStack
│   │   ├── providers.go          # GitHub, GitLab, Bitbucket, Generic
│   │   └── remote.go             # RemoteStore, registry metadata
│   ├── store/
│   │   └── store.go              # boiler.meta.json read/write, ParseResourceName
│   └── utils/                    # Shared helpers
│       ├── fs.go                 # CopyFile, CopyDir, CopyFileWithVariables
│       ├── metadata.go           # ParseSnippetMetadata, ValidateSnippetMetadata
│       ├── vars.go               # Variable resolution (:KEY, bl__VAR)
│       ├── helpers.go            # LoadStore, PickFromList, FindMatchingResources
│       ├── logger.go
│       ├── prompt.go
│       └── messages.go           # Shared error/success message constants
│
├── pkg/
│   └── version/                   # Version string injected at build time
│
├── scripts/                       # Install and uninstall scripts
│   ├── install.sh
│   ├── install.ps1
│   ├── uninstall.sh
│   └── uninstall.ps1
│
└── web/                           # Documentation website (Astro Starlight)
    └── src/content/docs/          # Markdown doc pages
```

---

## Adding a New Command

Commands use package-level variables for flags and `func init()` for registration. Follow the existing pattern:

### 1. Create `internal/cli/export.go`

```go
package cli

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
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
    // cfg and logger are package-level, available to all commands
    logger.Info("exporting snippets")
    // implementation here
    return nil
}

var exportForce bool

func init() {
    exportCmd.Flags().BoolVarP(&exportForce, "force", "f", false, "Overwrite existing files")
    rootCmd.AddCommand(exportCmd)
}
```

### 2. Add docs: `web/src/content/docs/commands/export.md`

No changes to `root.go` needed - `init()` handles registration automatically.

---

## Code Style

- Run `gofmt` before committing
- Run `go vet ./...` - fix all warnings before submitting
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

- [ ] Code builds: `go build -o bl main.go`
- [ ] Tests pass: `go test ./...`
- [ ] `go vet ./...` is clean
- [ ] Related issue linked (`Closes #N`)
- [ ] Docs updated if behavior changed
- [ ] Commit messages follow the convention above

---

## Running Locally

```bash
# Development build
go build -o bl main.go

# Cross-compile
GOOS=windows GOARCH=amd64 go build -o bl.exe main.go
GOOS=linux   GOARCH=amd64 go build -o bl main.go
GOOS=darwin  GOARCH=amd64 go build -o bl main.go
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
