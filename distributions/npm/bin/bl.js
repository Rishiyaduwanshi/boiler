#!/usr/bin/env node

const { existsSync } = require("node:fs");
const path = require("node:path");
const { spawnSync } = require("node:child_process");

const binaryName = process.platform === "win32" ? "bl.exe" : "bl";
const binaryPath = path.resolve(__dirname, "..", "runtime", binaryName);

if (!existsSync(binaryPath)) {
  console.error("Boiler runtime is not installed. Reinstall the package: npm i -g @boilercli/core");
  process.exit(1);
}

const result = spawnSync(binaryPath, process.argv.slice(2), {
  stdio: "inherit",
  windowsHide: false,
});

if (result.error) {
  console.error(result.error.message);
  process.exit(1);
}

process.exit(result.status ?? 0);
