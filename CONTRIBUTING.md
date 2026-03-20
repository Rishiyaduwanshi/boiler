# Contributing to Boiler

> **⚠️ IMPORTANT: Please Discuss Before Starting Work**
>
> **DO NOT start working on features or major changes without discussing first!** Open an issue or comment on an existing one to discuss your proposed changes before writing code. This ensures:
> - Your effort aligns with the project's direction
> - No duplicate work happens
> - You understand the context and requirements
>
> **I am not responsible if your PR is rejected without prior discussion.** Save yourself time and frustration by talking first, coding second.

---

## 📢 Current State of the Codebase

**Please be aware:** This project is currently focused on **functionality over code quality**. Known issues include:

- 🔧 **Code is cluttered** - Refactoring is planned but not prioritized yet
- 📦 **Duplication exists** - Some code is repeated, will be DRY-ed later
- 🏗️ **Architecture needs cleanup** - Current structure works but isn't optimal
- 📝 **Incomplete documentation** - Docs are work-in-progress
- 🧪 **Limited test coverage** - Tests will be added incrementally

**These are known and will be addressed in future iterations.**  **please discuss first** before submitting large refactoring PRs.

---

Thank you for your interest in contributing to Boiler! This guide will help you get started.

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

# Run tests (if available)
go test ./...
```

---

## Development Workflow

### 1. **Fork & Branch**
```bash
# Fork the repo on GitHub, then:
git clone https://github.com/YOUR_USERNAME/boiler.git
cd boiler
git checkout -b feature/your-feature-name
```

### 2. **Make Changes**
- Write clean, idiomatic Go code
- Follow existing code style and patterns
- Add comments for complex logic
- Update documentation if needed

### 3. **Test Your Changes**
```bash
# Build and test locally
go build -o bl main.go
./bl init
./bl store test.js
./bl add test
```

### 4. **Commit & Push**
```bash
git add .
git commit -m "feat: add your feature description"
git push origin feature/your-feature-name
```

### 5. **Open Pull Request**
- Go to GitHub and create a PR from your fork
- Describe what you changed and why
- Link any related issues

---

## Commit Message Convention

Use conventional commit format:

```
<type>: <description>

[optional body]
[optional footer]
```

**Types:**
- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation changes
- `refactor:` - Code refactoring
- `test:` - Adding tests
- `chore:` - Build/tooling changes
- `ci:` - CI/CD changes

**Examples:**
```
feat: add template variable support
fix: handle file lock during self-update on Windows
docs: update README with new examples
refactor: simplify version parsing logic
```

---

## Project Structure

```
boiler/
├── main.go                    # Entry point
├── cmd/boiler/                # CLI binary
├── internal/
│   ├── cli/                   # Command implementations
│   │   ├── add.go            # bl add command
│   │   ├── store.go          # bl store command
│   │   ├── list.go           # bl ls command
│   │   ├── search.go         # bl search command
│   │   └── ...
│   ├── config/               # Configuration management
│   │   └── config.go         # Load/save settings
│   ├── store/                # Storage operations
│   │   └── store.go          # File/metadata handling
│   └── utils/                # Utilities
│       ├── fs.go             # File system helpers
│       ├── logger.go         # Logging
│       └── prompt.go         # User prompts
├── scripts/                   # Install/uninstall scripts
│   ├── install.ps1           # Windows installer
│   └── install.sh            # Linux/macOS installer
├── store/                     # Default storage location
│   ├── snippets/             # Code snippets
│   └── stacks/               # Project templates
└── web/                       # Documentation website (Starlight)
    └── src/content/docs/     # Markdown docs
```

---

## Adding a New Command

To add a new command (e.g., `bl export`):

### 1. Create command file: `internal/cli/export.go`
```go
package cli

import (
    "github.com/spf13/cobra"
)

func newExportCmd(cfg *config.Config, logger *utils.Logger) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "export",
        Short: "Export snippets to a file",
        RunE: func(cmd *cobra.Command, args []string) error {
            // Implementation here
            return nil
        },
    }
    return cmd
}
```

### 2. Register in `internal/cli/root.go`
```go
func Execute(cfg *config.Config, logger *utils.Logger) error {
    rootCmd := &cobra.Command{...}
    
    // Add your command
    rootCmd.AddCommand(newExportCmd(cfg, logger))
    
    return rootCmd.Execute()
}
```

### 3. Update documentation
Add a new file: `web/src/content/docs/commands/export.md`

---

## Code Style Guidelines

### Go Conventions
- Use `gofmt` for formatting
- Run `go vet` before committing
- Keep functions small and focused
- Use descriptive variable names
- Add error handling for all operations

### Example: Good vs Bad
```go
// ❌ Bad
func d(f string) error {
    _, err := os.Stat(f)
    return err
}

// ✅ Good
func fileExists(filePath string) (bool, error) {
    _, err := os.Stat(filePath)
    if os.IsNotExist(err) {
        return false, nil
    }
    return err == nil, err
}
```

---

## Areas to Contribute

### 🐛 Bug Fixes
- Check [GitHub Issues](https://github.com/rishiyaduwanshi/boiler/issues) for bugs
- Reproduce the issue locally
- Fix and test thoroughly
- Add tests if possible

### 📝 Documentation
- Improve README examples
- Add tutorials to website
- Fix typos or unclear sections
- Add code comments

### 🧪 Testing
- Add unit tests for utilities
- Add integration tests for commands
- Test on different platforms (Windows, Linux, macOS)

### 🎨 UI/UX
- Improve CLI output formatting
- Better error messages
- Progress bars for long operations
- Colored output

---

## Running Locally

### Build for development
```bash
go build -o bl main.go
```

### Build for production
```bash
# Current platform
go build -ldflags="-s -w" -o bl main.go

# Cross-compile for Windows
GOOS=windows GOARCH=amd64 go build -o bl.exe main.go

# Cross-compile for Linux
GOOS=linux GOARCH=amd64 go build -o bl main.go

# Cross-compile for macOS
GOOS=darwin GOARCH=amd64 go build -o bl main.go
```

### Test installer script locally
```bash
# Windows (PowerShell)
.\scripts\install.ps1

# Linux/macOS
bash scripts/install.sh
```

---

## Documentation Website

The documentation is built with [Starlight](https://starlight.astro.build/).

### Local development
```bash
cd web
npm install
npm run dev
# Visit http://localhost:4321
```

### Build documentation
```bash
cd web
npm run build
```

### Add a new doc page
1. Create: `web/src/content/docs/your-page.md`
2. Update: `web/astro.config.mjs` (add to sidebar)

---

## Pull Request Guidelines

### Before submitting:
- ✅ Code builds without errors
- ✅ Tested locally on your platform
- ✅ Follows existing code style
- ✅ Commit messages follow convention
- ✅ Documentation updated (if needed)

### PR Description Template:
```markdown
## What changed?
Brief description of your changes.

## Why?
Explain the motivation for this change.

## How to test?
Steps to test your changes:
1. Run `bl store test.js`
2. Run `bl add test`
3. Verify output

## Related Issues
Closes #123
```

---

## Questions or Help?

- 💬 [GitHub Discussions](https://github.com/rishiyaduwanshi/boiler/discussions)
- 🐛 [Report Issues](https://github.com/rishiyaduwanshi/boiler/issues)
- 📖 [Read the Docs](https://boiler.iamabhinav.dev)

---

## License

By contributing, you agree that your contributions will be licensed under the Apache License 2.0.

---

**Thank you for contributing to Boiler!** 🚀
