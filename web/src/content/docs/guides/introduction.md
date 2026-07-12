---
title: Introduction
description: Boiler is a code automation engine - fetch, scaffold, and reuse code across every project
---

Boiler is a CLI tool that turns repetitive coding tasks into single commands. Fetch code from anywhere, scaffold full features, and reuse your best work across every project.

## What Can You Do With Boiler?

### Fetch From Anywhere - Without `git clone`

Pull individual files, specific folders, or entire repos from GitHub, GitLab, Bitbucket, or any direct URL:

```bash
bl use alice/snippets:js/errorHandler.js ./src/utils
bl use vercel/next.js:examples/blog ./my-blog
bl use https://gitlab.com/myorg/templates:Dockerfile .
```

### Instant Gitignores for Any Stack

The built-in `gi` alias connects to 225+ popular templates and 300+ community templates:

```bash
bl gi Node
bl gi Python
bl gi Go
bl gi Global/macOS
```

### Store & Reuse Your Own Code

```bash
bl store utils/debounce.js     # Store once
bl add debounce ./src/utils    # Reuse anywhere
```

Template variables let you customize snippets on the fly - `bl__VAR_NAME` placeholders are prompted interactively when you add them.

### Automate Full Feature Scaffolding

Write `.bl` scripts that create multiple files, inject code into existing ones, and run commands - triggered with a single `bl new` call:

```bash
bl new feat user
# Creates: src/routes/user.route.js, src/controllers/user.controller.js
# Injects: imports and app.use() into src/api.js automatically
```

---

## Core Concepts

<table>
  <thead>
    <tr>
      <th>Concept</th>
      <th>What it is</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td><strong>Snippet</strong></td>
      <td>A single file stored in your Boiler store</td>
    </tr>
    <tr>
      <td><strong>Stack</strong></td>
      <td>An entire directory/project template</td>
    </tr>
    <tr>
      <td><strong>.bl Script</strong></td>
      <td>An automation script run with <code>bl new</code></td>
    </tr>
    <tr>
      <td><strong>Global Store</strong></td>
      <td><code>~/.boiler/</code> - shared across all your projects</td>
    </tr>
    <tr>
      <td><strong>Local Store</strong></td>
      <td><code>bl/</code> inside your project - project-specific templates</td>
    </tr>
    <tr>
      <td><strong>Registry</strong></td>
      <td>A GitHub/GitLab repo used as a shared snippet library for teams</td>
    </tr>
  </tbody>
</table>

---

## Language & Platform Agnostic

Boiler works with **any file type** - JavaScript, TypeScript, Python, Go, Rust, Java, C++, Dockerfiles, shell scripts, config files, and more. It runs on **Windows, macOS, and Linux**.

---

## Next Steps

- [Installation](/guides/installation/) - Get Boiler running in 30 seconds
- [Quick Start](/guides/quickstart/) - Your first snippet in 5 minutes
- [Use Cases](/guides/usecases/) - Real-world workflows and examples
- [Boiler Scripts (.bl)](/guides/bl-scripts/) - Write full automation pipelines
