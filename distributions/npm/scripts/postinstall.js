#!/usr/bin/env node

const fs = require("node:fs");
const fsp = require("node:fs/promises");
const os = require("node:os");
const path = require("node:path");
const https = require("node:https");
const crypto = require("node:crypto");
const tar = require("tar");
const AdmZip = require("adm-zip");

const PACKAGE_JSON_PATH = path.resolve(__dirname, "..", "package.json");
const RUNTIME_DIR = path.resolve(__dirname, "..", "runtime");
const REPO = "rishiyaduwanshi/boiler";

function platformName(platform) {
  if (platform === "win32") return "Windows";
  if (platform === "linux") return "Linux";
  if (platform === "darwin") return "Darwin";
  throw new Error(`Unsupported platform: ${platform}`);
}

function archName(arch) {
  if (arch === "x64") return "x86_64";
  if (arch === "arm64") return "arm64";
  throw new Error(`Unsupported architecture: ${arch}`);
}

function readPackageVersion() {
  const raw = fs.readFileSync(PACKAGE_JSON_PATH, "utf8");
  const pkg = JSON.parse(raw);
  return pkg.version;
}

function download(url, destinationPath) {
  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(destinationPath);
    https
      .get(
        url,
        {
          headers: {
            "User-Agent": "boiler-npm-installer",
          },
        },
        (response) => {
          if (response.statusCode >= 300 && response.statusCode < 400 && response.headers.location) {
            file.close();
            fs.unlink(destinationPath, () => {
              download(response.headers.location, destinationPath).then(resolve).catch(reject);
            });
            return;
          }

          if (response.statusCode !== 200) {
            file.close();
            fs.unlink(destinationPath, () => {
              reject(new Error(`Download failed (${response.statusCode}) for ${url}`));
            });
            return;
          }

          response.pipe(file);
          file.on("finish", () => {
            file.close();
            resolve();
          });
        }
      )
      .on("error", (error) => {
        file.close();
        fs.unlink(destinationPath, () => reject(error));
      });
  });
}

function sha256(filePath) {
  const hash = crypto.createHash("sha256");
  const data = fs.readFileSync(filePath);
  hash.update(data);
  return hash.digest("hex");
}

function parseChecksum(checksumText, fileName) {
  const line = checksumText
    .split(/\r?\n/)
    .find((entry) => entry.trim().endsWith(` ${fileName}`));

  if (!line) {
    throw new Error(`Checksum for ${fileName} not found`);
  }

  const parts = line.trim().split(/\s+/);
  return parts[0];
}

async function findBinary(rootDir, binaryName) {
  const entries = await fsp.readdir(rootDir, { withFileTypes: true });

  for (const entry of entries) {
    const fullPath = path.join(rootDir, entry.name);
    if (entry.isDirectory()) {
      const nested = await findBinary(fullPath, binaryName);
      if (nested) {
        return nested;
      }
    } else if (entry.isFile() && entry.name === binaryName) {
      return fullPath;
    }
  }

  return "";
}

async function main() {
  const version = readPackageVersion();
  const tag = `v${version}`;

  if (version.includes("development")) {
    throw new Error("Package version is not set for release publishing");
  }

  const osName = platformName(process.platform);
  const architecture = archName(process.arch);
  const extension = process.platform === "win32" ? "zip" : "tar.gz";
  const binaryFileName = process.platform === "win32" ? "bl.exe" : "bl";
  const assetName = `boiler_${osName}_${architecture}.${extension}`;

  const releaseAssetUrl = `https://github.com/${REPO}/releases/download/${tag}/${assetName}`;
  const checksumsUrl = `https://github.com/${REPO}/releases/download/${tag}/checksums.txt`;

  const tempRoot = await fsp.mkdtemp(path.join(os.tmpdir(), "boiler-npm-"));
  const archivePath = path.join(tempRoot, assetName);
  const checksumPath = path.join(tempRoot, "checksums.txt");
  const extractDir = path.join(tempRoot, "extract");

  await fsp.mkdir(extractDir, { recursive: true });

  console.log(`Downloading Boiler ${tag} for ${process.platform}/${process.arch}`);
  await download(releaseAssetUrl, archivePath);
  await download(checksumsUrl, checksumPath);

  const checksumText = await fsp.readFile(checksumPath, "utf8");
  const expectedHash = parseChecksum(checksumText, assetName).toLowerCase();
  const actualHash = sha256(archivePath).toLowerCase();

  if (expectedHash !== actualHash) {
    throw new Error(`Checksum mismatch for ${assetName}`);
  }

  if (extension === "zip") {
    const zip = new AdmZip(archivePath);
    zip.extractAllTo(extractDir, true);
  } else {
    await tar.x({ file: archivePath, cwd: extractDir });
  }

  const extractedBinary = await findBinary(extractDir, binaryFileName);
  if (!extractedBinary) {
    throw new Error(`${binaryFileName} not found in ${assetName}`);
  }

  await fsp.mkdir(RUNTIME_DIR, { recursive: true });
  const targetBinary = path.join(RUNTIME_DIR, binaryFileName);
  await fsp.copyFile(extractedBinary, targetBinary);

  if (process.platform !== "win32") {
    await fsp.chmod(targetBinary, 0o755);
  }

  await fsp.rm(tempRoot, { recursive: true, force: true });
}

main().catch(async (error) => {
  console.error(`Boiler install failed: ${error.message}`);
  process.exit(1);
});
