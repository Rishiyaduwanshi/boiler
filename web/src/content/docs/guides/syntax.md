---
title: Boiler Syntax Reference
description: Complete guide to Boiler's metadata and template syntax
---

Boiler uses a simple comment-based syntax for defining metadata and template variables in your snippets. This guide covers all syntax features and conventions.

## Metadata Comments

Metadata is defined using special comment keywords that start with double underscores (`__`). Boiler automatically detects your file's comment style based on the extension.

### Required Metadata

#### `__author`

Identifies the snippet creator. Required for all snippets.

```javascript
// __author John Doe
```

### Optional Metadata

#### `__desc`

Brief description of what the snippet does.

```javascript
// __desc Database connection utility with retry logic
```

#### `__version`

Internal version identifier (optional, auto-managed by filename).

```javascript
// __version 1
```

**Note:** Version is primarily managed through filenames (`file@1.js`, `file@2.js`). This metadata field is optional.

## Template Variables

Variables allow you to create reusable templates that prompt for values when added to a project.

### Syntax

```
// __var VARIABLE_NAME = DefaultValue
```

- **Variable names**: Use `bl__` prefix followed by uppercase with underscores
- **Default values**: Shown to users during `bl add`, can be overridden
- **Replacement**: All occurrences of the variable name in your code are replaced

### Example: API Client

```javascript
// __author Jane Smith
// __desc REST API client with configurable endpoint
// __var bl__API_URL = http://localhost:3000
// __var bl__API_KEY = your-api-key-here
// __var bl__TIMEOUT = 5000

const apiClient = {
  baseURL: 'bl__API_URL',
  apiKey: 'bl__API_KEY',
  timeout: bl__TIMEOUT,
  
  async fetch(endpoint) {
    const response = await fetch(`${this.baseURL}${endpoint}`, {
      headers: { 'Authorization': `Bearer ${this.apiKey}` },
      timeout: this.timeout
    });
    return response.json();
  }
};

module.exports = apiClient;
```

### When User Adds This Snippet

```bash
bl add apiClient
```

**Interactive Prompt:**
```
Template variables found:
  bl__API_URL [http://localhost:3000]: https://api.myapp.com
  bl__API_KEY [your-api-key-here]: sk_live_abc123xyz
  bl__TIMEOUT [5000]: 10000
✓ Snippet added: apiClient@1.js → ./apiClient.js
```

**Final Output (`apiClient.js`):**
```javascript
const apiClient = {
  baseURL: 'https://api.myapp.com',
  apiKey: 'sk_live_abc123xyz',
  timeout: 10000,
  
  async fetch(endpoint) {
    const response = await fetch(`${this.baseURL}${endpoint}`, {
      headers: { 'Authorization': `Bearer ${this.apiKey}` },
      timeout: this.timeout
    });
    return response.json();
  }
};

module.exports = apiClient;
```

**Notice:**
- All `bl__VAR_NAME` replaced with user values
- All metadata comments (`__author`, `__desc`, `__var`) removed
- Clean, production-ready code

## Multiple Variables

You can use as many variables as needed in a single snippet:

```python
# __author DevTeam
# __desc Database configuration with multiple environments
# __var bl__DB_HOST = localhost
# __var bl__DB_PORT = 5432
# __var bl__DB_NAME = myapp_db
# __var bl__DB_USER = postgres
# __var bl__DB_PASS = password123
# __var bl__DB_POOL_SIZE = 10

DATABASE_CONFIG = {
    'host': 'bl__DB_HOST',
    'port': bl__DB_PORT,
    'database': 'bl__DB_NAME',
    'user': 'bl__DB_USER',
    'password': 'bl__DB_PASS',
    'pool_size': bl__DB_POOL_SIZE
}
```

### Custom Comment Styles

For unsupported file types, add them to your config:

```bash
# Edit config
bl conf -e
```

Add to the `artifacts` section:

```json
{
  "artifacts": {
    "default": "//  ",
    "custom": "#  ",
    "dockerfile": "#  ",
    "makefile": "#  ",
    "nginx": "#  "
  }
}
```

**Format:** `"extension": "comment_prefix  "` (note the two spaces after prefix)

## Variable Naming Conventions

### ✅ Good Variable Names

```javascript
// __var bl__API_URL = http://localhost
// __var bl__MAX_RETRIES = 3
// __var bl__TIMEOUT_MS = 5000
// __var bl__DB_HOST = localhost
```

**Conventions:**
- Always start with `bl__` prefix
- Use UPPERCASE with underscores
- Descriptive and clear purpose
- No spaces or special characters

### ❌ Bad Variable Names

```javascript
// __var apiUrl = http://localhost           // Missing bl__ prefix
// __var BL_TIMEOUT = 5000                   // Wrong prefix format
// __var bl__my-var = test                   // Hyphens not allowed
// __var bl__some var = value                // Spaces not allowed
```

##  Examples

### 1. Error Handler with Custom Logger

```javascript
// __author ErrorHandling Team
// __desc Centralized error handler with logging
// __var bl__LOG_LEVEL = error
// __var bl__ENABLE_STACK_TRACE = true
// __var bl__NOTIFY_EMAIL = admin@example.com

function handleError(error, context = {}) {
  const logLevel = 'bl__LOG_LEVEL';
  const showStack = bl__ENABLE_STACK_TRACE;
  const notifyEmail = 'bl__NOTIFY_EMAIL';
  
  console[logLevel]('Error occurred:', error.message);
  
  if (showStack && error.stack) {
    console.error(error.stack);
  }
  
  if (logLevel === 'error') {
    sendEmailNotification(notifyEmail, error, context);
  }
  
  return {
    success: false,
    error: error.message,
    timestamp: new Date().toISOString()
  };
}
```

### 2. Database Model Template

```python
# __author Backend Team
# __desc SQLAlchemy model template
# __var bl__TABLE_NAME = users
# __var bl__PRIMARY_KEY = id
# __var bl__TIMESTAMP_FIELDS = true

from sqlalchemy import Column, Integer, String, DateTime
from sqlalchemy.ext.declarative import declarative_base
from datetime import datetime

Base = declarative_base()

class bl__TABLE_NAME(Base):
    __tablename__ = 'bl__TABLE_NAME'
    
    bl__PRIMARY_KEY = Column(Integer, primary_key=True)
    name = Column(String(100), nullable=False)
    email = Column(String(255), unique=True)
    
    # Conditional timestamp fields
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, onupdate=datetime.utcnow)
```

### 3. Configuration File Template

```yaml
# __author DevOps Team
# __desc Kubernetes deployment config
# __var bl__APP_NAME = my-app
# __var bl__NAMESPACE = default
# __var bl__REPLICAS = 3
# __var bl__IMAGE = nginx:latest
# __var bl__PORT = 80

apiVersion: apps/v1
kind: Deployment
metadata:
  name: bl__APP_NAME
  namespace: bl__NAMESPACE
spec:
  replicas: bl__REPLICAS
  selector:
    matchLabels:
      app: bl__APP_NAME
  template:
    metadata:
      labels:
        app: bl__APP_NAME
    spec:
      containers:
      - name: bl__APP_NAME
        image: bl__IMAGE
        ports:
        - containerPort: bl__PORT
```

## Best Practices

### 1. Use Meaningful Defaults

Provide sensible defaults that work in common scenarios:

```javascript
// ✅ Good: Sensible defaults
// __var bl__RETRY_COUNT = 3
// __var bl__TIMEOUT = 5000

// ❌ Bad: No context
// __var bl__VALUE = 123
// __var bl__THING = abc
```

### 2. Document Complex Variables

For non-obvious variables, add inline comments:

```javascript
// __var bl__MAX_CONNECTIONS = 10
// Maximum concurrent database connections (recommended: 5-20)
const pool = createPool({ maxConnections: bl__MAX_CONNECTIONS });
```

### 3. Group Related Variables

Keep related configuration together:

```javascript
// Database configuration
// __var bl__DB_HOST = localhost
// __var bl__DB_PORT = 5432
// __var bl__DB_NAME = myapp

// Cache configuration  
// __var bl__CACHE_TTL = 3600
// __var bl__CACHE_MAX_SIZE = 1000
```

### 4. Validate Critical Values

Add validation for important variables:

```javascript
// __var bl__PORT = 3000

const PORT = parseInt(bl__PORT);
if (isNaN(PORT) || PORT < 1024 || PORT > 65535) {
  throw new Error('Invalid port number');
}
```

## Working with Stacks

For stack templates (entire project directories), use `boiler.stack.json` instead:

```json
{
  "id": "express-api",
  "version": "1",
  "author": "Backend Team",
  "description": "Express.js REST API boilerplate",
  "ignore": ["node_modules", ".git", "dist"]
}
```

Stacks don't support template variables (yet), but you can use environment variables or config files within the stack.

## Tips & Tricks

### 1. Use Variables in Strings and Code

Variables work anywhere in your file:

```javascript
// In strings
const message = "bl__APP_NAME is running on port bl__PORT";

// In object keys (as strings)
const config = {
  "bl__ENV": { /* ... */ }
};

// In calculations
const timeout = bl__TIMEOUT_MS * 1000;
```

### 2. Boolean Variables

Use string values for booleans, parse in code:

```javascript
// __var bl__ENABLE_DEBUG = true

const DEBUG = 'bl__ENABLE_DEBUG' === 'true';
```

### 3. List/Array Variables

For simple lists, use comma-separated strings:

```python
# __var bl__ALLOWED_ORIGINS = http://localhost:3000,http://localhost:8080

ALLOWED_ORIGINS = 'bl__ALLOWED_ORIGINS'.split(',')
```

## CLI Variables & Aliases

The magic of `bl__VAR_NAME` isn't limited to snippet files! Boiler unifies this syntax across its entire ecosystem. You can use the exact same variable format directly in your terminal commands and aliases.

### Inline Variables

If you configure a variable in Boiler (`bl var org=rishiyaduwanshi`), you can inject it inline into any CLI argument:

```bash
bl use github.com/bl__org/boiler-templates:auth
```
Boiler dynamically resolves `bl__org` to `rishiyaduwanshi` before executing the command, making it incredibly powerful for dynamic remote fetching.

### Alias Positional Arguments

Boiler aliases support dynamic positional arguments using the `bl__N` syntax (`bl__1`, `bl__2`, etc.):

```bash
# Define a powerful alias
bl alias get-ignore="use github/gitignore:bl__1.gitignore"

# Execute it with an argument
bl get-ignore Node -m .gitignore
```
1. `bl__1` is replaced with your first argument (`Node`).
2. Any remaining unconsumed arguments (like `-m .gitignore`) are safely appended to the end of the expanded command.
3. The final executed command is: `bl use github/gitignore:Node.gitignore -m .gitignore`.

## Summary

- **Metadata**: Use `__author` (required), `__desc`, `__version` (optional)
- **Variables**: Format `// __var bl__NAME = default`
- **Naming**: Always use `bl__` prefix, UPPERCASE_WITH_UNDERSCORES
- **Replacement**: All occurrences replaced, metadata stripped from final output
- **Comment styles**: Auto-detected by extension, customizable in config
- **Use cases**: API configs, database credentials, environment settings, templates

Boiler's syntax is simple yet powerful - treat it as a lightweight templating language for your code snippets!
