---
title: Use Cases
description: Real-world scenarios and practical examples of using Boiler
---

Boiler is designed to eliminate repetitive coding tasks by storing and reusing code across projects. Here are practical use cases to get the most out of it.

---

## 1. Avoid Installing Entire Packages for Small Utilities

**Have you ever thought:** Sometimes we add a package even for a small use case?

**The Problem:**
```bash
# Just need one function from lodash?
npm install lodash  # 1.4MB for one function!

# Need a date formatter?
npm install moment  # 289KB for basic formatting!
```

**Boiler Solution:**
```javascript
// debounce.js
// __author Your Name
// __desc Debounce utility without lodash

function debounce(func, wait) {
  let timeout;
  return function executedFunction(...args) {
    const later = () => {
      clearTimeout(timeout);
      func(...args);
    };
    clearTimeout(timeout);
    timeout = setTimeout(later, wait);
  };
}

module.exports = debounce;
```

```bash
# Store once
bl store debounce.js

# Reuse everywhere - no package needed!
bl add debounce
```

**Benefits:**
- **Zero dependencies** - No node_modules bloat
- **Lightweight** - Only the code you need
- **Custom tailored** - Modify to your exact needs
- **Bundle size** - Smaller production builds

**Perfect for:** Small utilities like debounce, throttle, formatters, validators, parsers, etc.

---

## 2. Store & Reuse API Middleware

**Scenario:** You use the same authentication middleware in every Express project.

```bash
# Store it once
cd my-express-project
bl store middleware/auth.js

# Use in any new project
cd new-project
bl add auth
```

**Result:** Never write `verifyToken()` or `requireAuth()` again.

---

## 3. Database Connection Configs

**Scenario:** MongoDB/PostgreSQL connection setup is identical across microservices.

```bash
# Store your database config
bl store config/db.js

# Add to any service instantly
cd user-service
bl add db
```

**Perfect for:** Microservice architectures where consistency matters.

---

## 4. Error Handlers & Logging Utilities

**Scenario:** You have a custom error handler and logger you always use.

```bash
# Store utilities
bl store utils/errorHandler.js
bl store utils/appLogger.js

# Add both to new project
bl add errorHandler
bl add appLogger
```

**Why it matters:** Consistent error handling across all projects.

---

## 5. Project Boilerplates (Stacks)

**Scenario:** You repeatedly scaffold Express APIs with the same structure.

```bash
# Create your perfect Express template once
cd my-express-template
bl init  # Creates boiler.stack.json

# Customize boiler.stack.json to exclude node_modules, .env
# Then store the entire structure
bl store

# Start new projects in seconds
mkdir new-api && cd new-api
bl add express-api@1
npm install
```

**Use for:**
- Express/Fastify APIs
- Next.js starters
- Django REST templates
- Microservice scaffolds

---

## 6. Team Snippet Library

**Scenario:** Your team needs standardized code patterns shared across machines.

```bash
# Store team-approved patterns in a shared GitHub/GitLab repo
# Then any team member can fetch them:

# One-time registry setup per machine
bl conf --set-registry https://github.com/myteam/boiler-snippets

# Fetch and save to local store (reusable offline after)
bl add jwtHelper -r
bl add validation -r
bl add rateLimiter -r

# Or one-shot without saving to local store
bl use https://github.com/myteam/boiler-snippets:js/jwtHelper.js
```

**Benefit:** Code consistency across the entire team, across machines — no manual file sharing needed.

---

## 7. Versioned Code Evolution

**Scenario:** You improve a utility and want to keep both versions.

```bash
# Store version 1
bl store utils/emailService.js
# → Saved as emailService@1.js

# Later, improve it and store v2
# Update __version to 2 in file comments
bl store utils/emailService.js
# → Saved as emailService@2.js

# Use specific versions
bl add emailService@1  # Old projects
bl add emailService@2  # New projects
```

**When to use:** Gradual migration without breaking old projects.

---

## 8. Cross-Language Snippets

**Scenario:** You work in multiple languages and want to store patterns for all.

```bash
# Python async wrapper
bl init -n
# → Create asyncWrapper.py with __author and __version
bl store asyncWrapper.py

# JavaScript promise handler
bl init -n
# → Create promiseHandler.js
bl store promiseHandler.js

# Go error wrapper
bl init -n
# → Create errorWrapper.go
bl store errorWrapper.go
```

**Perfect for:** Polyglot developers managing multiple tech stacks.

---

## 9. Configuration Files

**Scenario:** You use the same ESLint, Prettier, or Docker configs everywhere.

```bash
# Store configs
bl store .eslintrc.json
bl store .prettierrc
bl store Dockerfile

# Add to new projects
bl add eslintrc
bl add prettierrc
bl add Dockerfile
```

**Saves time on:** Project setup and maintaining consistency.

---

## 10. Testing Utilities

**Scenario:** You have custom test helpers and fixtures.

```bash
# Store test utilities
bl store tests/helpers/mockData.js
bl store tests/helpers/testSetup.js

# Add to new test suites
bl add mockData
bl add testSetup
```

**Why:** Faster test setup with proven patterns.

---

## 11. Quick Script Deployment

**Scenario:** You have shell scripts for deployment, backups, or automation.

```bash
# Store scripts
bl store scripts/deploy.sh
bl store scripts/db-backup.sh

# Use in CI/CD or other projects
bl add deploy
bl add db-backup
```

**Use for:** DevOps automation and deployment pipelines.

---

## Best Practices

### 1. **Initialize Before Storing**
Always run `bl init` to add metadata (author, version, description) before storing.

### 2. **Use Descriptive Names**
```bash
# Bad
bl store helper.js

# Good
bl store jwtTokenHelper.js
```

### 3. **Version Strategically**
- Increment version for breaking changes
- Keep old versions for legacy projects

### 4. **Store Stacks with Proper Ignore Patterns**
Edit `boiler.stack.json` to exclude:
```json
{
  "ignore": [
    "node_modules",
    ".env",
    "dist",
    ".git"
  ]
}
```

### 5. **Search Before Creating**
```bash
bl search jwt
# Check if similar utility already exists
```

---

## Common Workflows

### New Project Setup
```bash
# Add stack template
bl add express-api

# Add common utilities
bl add errorHandler
bl add logger
bl add dbConfig

# Start coding
npm install
npm start
```

### Microservice Development
```bash
# Store shared middleware once
bl store auth.middleware.js
bl store cors.middleware.js

# Use across all services
cd user-service && bl add auth && bl add cors
cd order-service && bl add auth && bl add cors
cd payment-service && bl add auth && bl add cors
```

---

### Personal Code Library
```bash
# Store everything you frequently use
bl store utils/*

# List your library
bl ls

# Add what you need
bl add <snippet-name>
```

---

## Tips & Tricks

1. **Alias frequently used snippets:**
   ```bash
   bl config  # Edit config
   # Add aliases for quick access
   ```

2. **Search with keywords:**
   ```bash
   bl search middleware
   bl search validation
   ```

3. **Check info before adding:**
   ```bash
   bl info auth@2
   # See author, version, description
   ```

4. **Clean unused resources:**
   ```bash
   bl clean --snippets
   bl clean --stacks
   ```

---

## Next Steps

- [Quick Start Guide](/guides/quickstart/) - Get started in 5 minutes
- [Commands Reference](/commands/bl/) - Complete command documentation
- [Installation](/guides/installation/) - Setup instructions
