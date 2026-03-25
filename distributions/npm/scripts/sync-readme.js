#!/usr/bin/env node

const fs = require("node:fs");
const path = require("node:path");

const packageRoot = path.resolve(__dirname, "..");
const npmReadmePath = path.join(packageRoot, "README.md");
const repoReadmePath = process.env.BOILER_ROOT_README_PATH
  ? path.resolve(process.env.BOILER_ROOT_README_PATH)
  : path.resolve(packageRoot, "..", "..", "README.md");

if (!fs.existsSync(repoReadmePath)) {
  console.log("sync-readme: root README.md not found, keeping existing npm README.md");
  process.exit(0);
}

const rootReadme = fs.readFileSync(repoReadmePath, "utf8");
const npmReadme = fs.existsSync(npmReadmePath) ? fs.readFileSync(npmReadmePath, "utf8") : "";

if (npmReadme === rootReadme) {
  console.log("sync-readme: npm README.md already up to date");
  process.exit(0);
}

fs.writeFileSync(npmReadmePath, rootReadme, "utf8");
console.log("sync-readme: npm README.md synced from repository README.md");
