import { readFileSync, writeFileSync } from 'fs';
import { join, dirname } from 'path';
import { fileURLToPath } from 'url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const svgPath = join(__dirname, '../public/hero-illustration.svg');

async function updateVersion() {
  try {
    const res = await fetch('https://api.github.com/repos/rishiyaduwanshi/boiler/releases/latest');
    if (!res.ok) throw new Error(`GitHub API returned ${res.status}`);
    const data = await res.json();
    const version = data.tag_name;
    if (!version) throw new Error('No tag_name in response');

    let svg = readFileSync(svgPath, 'utf-8');
    svg = svg.replace(/v\d+\.\d+\.\d+/, version);
    writeFileSync(svgPath, svg);
    console.log(`[hero-version] Updated to ${version}`);
  } catch (e) {
    console.log(`[hero-version] Could not fetch version (${e.message}), keeping default`);
  }
}

updateVersion();
