// SPDX-License-Identifier: AGPL-3.0-only
import { createHash } from 'node:crypto';
import { readFile, writeFile, copyFile } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';
import path from 'node:path';

const root = fileURLToPath(new URL('../../', import.meta.url));
const pkg = JSON.parse(await readFile(path.join(root, 'package.json'), 'utf8'));
if (!/^\d+\.\d+\.\d+$/.test(pkg.version)) throw new Error('Stable x.y.z package version required.');
const release = JSON.parse(await readFile(path.join(root, 'dist/release.json'), 'utf8'));
if (release.version !== pkg.version) throw new Error('package.json version must match dist/release.json; rebuild the release.');
const checksumLines = (await readFile(path.join(root, 'dist/checksums.txt'), 'utf8')).trim().split(/\r?\n/);
const checksums = new Map();
for (const line of checksumLines) {
  const match = /^([a-f0-9]{64})\s+(yargitay-mcp_[a-z0-9_]+(?:\.exe)?)$/.exec(line);
  if (!match || checksums.has(match[2])) throw new Error('Malformed or duplicate release checksum.');
  checksums.set(match[2], match[1]);
}
const manifest = { version: pkg.version, repository: 'sakiphan/yargitay-mcp', assets: {} };
for (const os of ['darwin', 'linux', 'windows']) {
  for (const arch of ['amd64', 'arm64']) {
    const name = `yargitay-mcp_${os}_${arch}${os === 'windows' ? '.exe' : ''}`;
    const body = await readFile(path.join(root, 'dist', name));
    const sha256 = createHash('sha256').update(body).digest('hex');
    if (checksums.get(name) !== sha256) throw new Error(`Release checksum mismatch: ${name}`);
    manifest.assets[name] = { sha256, size: body.length };
  }
}
// Pin the exact release build, not a moving latest URL or a downloaded checksum list.
await writeFile(path.join(root, 'npm/release-manifest.json'), JSON.stringify(manifest, null, 2) + '\n');
await copyFile(path.join(root, 'dist/THIRD_PARTY_NOTICES.txt'), path.join(root, 'npm/THIRD_PARTY_NOTICES.txt'));
console.error(`Prepared npm manifest for v${pkg.version}: six verified assets.`);
