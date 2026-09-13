// SPDX-License-Identifier: AGPL-3.0-only
import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import { execFileSync } from 'node:child_process';
import path from 'node:path';
import { pathToFileURL } from 'node:url';
import { selectAsset, verifyBinary } from '../lib/launcher.mjs';

// Read only named JSON members; never extract or execute release package scripts.
function packedJSON(archive, member) {
  return JSON.parse(execFileSync('tar', ['-xOf', archive, `package/${member}`], {
    encoding: 'utf8', maxBuffer: 1024 * 1024, timeout: 10_000,
  }));
}

export async function verifyPublish(tag, directory, readPackedJSON = packedJSON) {
  assert.match(tag, /^v\d+\.\d+\.\d+$/, 'Stable vX.Y.Z release required');
  const version = tag.slice(1);
  const archive = path.resolve(directory, `sakiphan-yargitay-mcp-${version}.tgz`);
  const pkg = readPackedJSON(archive, 'package.json');
  assert.equal(pkg.name, 'sakiphan-yargitay-mcp', 'Unexpected npm package name');
  assert.equal(pkg.version, version, 'Package/tag version mismatch');
  assert.equal(pkg.license, 'AGPL-3.0-only', 'Unexpected package license');
  assert.equal(pkg.publishConfig?.registry, 'https://registry.npmjs.org/');
  assert.equal(pkg.publishConfig?.access, 'public');
  const manifest = readPackedJSON(archive, 'npm/release-manifest.json');
  assert.equal(manifest.version, version, 'Manifest/tag version mismatch');
  const release = JSON.parse(await readFile(path.join(directory, 'release.json'), 'utf8'));
  assert.equal(release.version, version, 'Build/tag version mismatch');
  const checksums = new Map();
  for (const line of (await readFile(path.join(directory, 'checksums.txt'), 'utf8')).trim().split(/\r?\n/)) {
    const match = /^([a-f0-9]{64})\s+(yargitay-mcp_[a-z0-9_]+(?:\.exe)?)$/.exec(line);
    assert.ok(match && !checksums.has(match[2]), 'Malformed or duplicate checksum');
    checksums.set(match[2], match[1]);
  }
  assert.equal(checksums.size, 6, 'Exactly six binary checksums required');
  for (const platform of ['darwin', 'linux', 'win32']) for (const arch of ['x64', 'arm64']) {
    const asset = selectAsset(manifest, platform, arch);
    assert.equal(checksums.get(asset.name), asset.sha256, `Release/manifest checksum mismatch: ${asset.name}`);
    assert.ok(await verifyBinary(path.join(directory, asset.name), asset), `Corrupt or missing binary: ${asset.name}`);
  }
  return archive;
}

if (process.argv[1] && import.meta.url === pathToFileURL(path.resolve(process.argv[1])).href) {
  await verifyPublish(process.argv[2] ?? '', process.argv[3] ?? 'dist');
  console.log('Publish verification passed: package identity, version and six release binaries match.');
}
