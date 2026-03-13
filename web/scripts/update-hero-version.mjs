import { readFileSync, writeFileSync } from 'fs';
import { join, dirname } from 'path';
import { fileURLToPath } from 'url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const svgPath = join(__dirname, '../public/hero-illustration.svg');

function updateVersion() {
  try {
    const version = (
      process.argv[2] ||
      process.env.BOILER_HERO_VERSION ||
      process.env.RELEASE_TAG ||
      ''
    ).trim();

    if (!version) {
      console.log('[hero-version] No release tag provided, keeping current SVG version');
      return;
    }

    let svg = readFileSync(svgPath, 'utf-8');
    const updatedSvg = svg.replace(/v\d+\.\d+\.\d+(?:[-+][0-9A-Za-z.-]+)?/, version);

    if (updatedSvg === svg) {
      console.log(`[hero-version] Version already ${version}`);
      return;
    }

    svg = updatedSvg;
    writeFileSync(svgPath, svg);
    console.log(`[hero-version] Updated to ${version}`);
  } catch (e) {
    console.log(`[hero-version] Failed to update version (${e.message})`);
    process.exitCode = 1;
  }
}

updateVersion();
