import { readFileSync } from 'fs';
import { join, dirname } from 'path';
import { fileURLToPath } from 'url';
import { spawnSync } from 'child_process';

const __dirname = dirname(fileURLToPath(import.meta.url));
const updateScriptPath = join(__dirname, '../web/scripts/update-hero-version.mjs');
const tagRefRe = /^refs\/tags\/(v\d+\.\d+\.\d+(?:[-+][0-9A-Za-z.-]+)?)$/;

function readPushRefsFromStdin() {
  if (process.stdin.isTTY) {
    return [];
  }

  const input = readFileSync(0, 'utf8').trim();
  if (!input) {
    return [];
  }

  return input
    .split('\n')
    .map((line) => line.trim())
    .filter(Boolean)
    .map((line) => {
      const [localRef, localSha, remoteRef, remoteSha] = line.split(/\s+/);
      return { localRef, localSha, remoteRef, remoteSha };
    });
}

function extractReleaseTags(refLines) {
  const tags = [];
  for (const line of refLines) {
    const match = tagRefRe.exec(line.localRef || '');
    if (!match) {
      continue;
    }

    // Skip tag deletions.
    if (!line.localSha || /^0+$/.test(line.localSha)) {
      continue;
    }

    tags.push(match[1]);
  }

  return [...new Set(tags)];
}

function run(command, args) {
  const result = spawnSync(command, args, {
    stdio: 'inherit',
    shell: false,
  });

  if (result.error) {
    throw result.error;
  }

  return result.status ?? 1;
}

function releaseTagsOnHead() {
  const result = spawnSync('git', ['tag', '--points-at', 'HEAD'], {
    stdio: ['ignore', 'pipe', 'inherit'],
    shell: false,
    encoding: 'utf8',
  });

  if (result.error) {
    throw result.error;
  }

  if (result.status !== 0) {
    return [];
  }

  return (result.stdout || '')
    .split('\n')
    .map((line) => line.trim())
    .filter((tag) => tagRefRe.test(`refs/tags/${tag}`));
}

function main() {
  const refLines = readPushRefsFromStdin();
  let tags = extractReleaseTags(refLines);

  // Lefthook can invoke pre-push commands without ref lines on stdin.
  // Fallback to release tags currently pointing at HEAD.
  if (tags.length === 0) {
    tags = releaseTagsOnHead();
  }

  if (tags.length === 0) {
    process.exit(0);
  }

  for (const tag of tags) {
    const scriptExitCode = run(process.execPath, [updateScriptPath, tag]);
    if (scriptExitCode !== 0) {
      process.exit(scriptExitCode);
    }
  }

  process.exit(0);
}

main();
