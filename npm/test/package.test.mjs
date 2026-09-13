// SPDX-License-Identifier: AGPL-3.0-only
import assert from 'node:assert/strict';
import { test } from 'node:test';
import { mkdtemp, mkdir, writeFile, readFile, copyFile, rm } from 'node:fs/promises';
import { createHash } from 'node:crypto';
import { spawnSync } from 'node:child_process';
import { tmpdir } from 'node:os';
import path from 'node:path';

async function fixture(t) {
  const root = await mkdtemp(path.join(tmpdir(), 'yargitay-pack-test-'));
  t.after(() => rm(root, { recursive: true, force: true }));
  await mkdir(path.join(root, 'npm/scripts'), { recursive: true });
  await mkdir(path.join(root, 'dist'));
  await copyFile(new URL('../scripts/prepare-package.mjs', import.meta.url), path.join(root, 'npm/scripts/prepare-package.mjs'));
  await writeFile(path.join(root, 'package.json'), JSON.stringify({ version: '0.1.0' }));
  await writeFile(path.join(root, 'dist/release.json'), JSON.stringify({ version: '0.1.0', commit: 'unknown' }));
  await writeFile(path.join(root, 'dist/THIRD_PARTY_NOTICES.txt'), 'Synthetic license fixture');
  const checksums = [];
  for (const os of ['darwin', 'linux', 'windows']) for (const arch of ['amd64', 'arm64']) {
    const name = `yargitay-mcp_${os}_${arch}${os === 'windows' ? '.exe' : ''}`;
    const body = Buffer.from(`synthetic ${name}`);
    await writeFile(path.join(root, 'dist', name), body);
    checksums.push(`${createHash('sha256').update(body).digest('hex')}  ${name}`);
  }
  await writeFile(path.join(root, 'dist/checksums.txt'), checksums.join('\n') + '\n');
  return root;
}
function prepare(root) {
  return spawnSync(process.execPath, ['npm/scripts/prepare-package.mjs'], { cwd: root, encoding: 'utf8', timeout: 10_000 });
}
test('prepack pins six build hashes and includes dependency license notices', async t => {
  const root = await fixture(t);
  const result = prepare(root);
  assert.equal(result.status, 0, result.stderr);
  assert.equal(result.stdout, '');
  const manifest = JSON.parse(await readFile(path.join(root, 'npm/release-manifest.json'), 'utf8'));
  assert.equal(manifest.repository, 'sakiphan/yargitay-mcp');
  assert.equal(manifest.version, '0.1.0');
  assert.equal(Object.keys(manifest.assets).length, 6);
  assert.equal(await readFile(path.join(root, 'npm/THIRD_PARTY_NOTICES.txt'), 'utf8'), 'Synthetic license fixture');
});
test('prepack rejects mismatched versions, corrupted or missing builds and duplicate checksums', async t => {
  for (const failure of ['version', 'hash', 'missing', 'duplicate']) {
    const root = await fixture(t);
    if (failure === 'version') await writeFile(path.join(root, 'dist/release.json'), '{"version":"0.2.0"}');
    if (failure === 'hash') await writeFile(path.join(root, 'dist/yargitay-mcp_linux_arm64'), 'corrupted');
    if (failure === 'missing') await rm(path.join(root, 'dist/yargitay-mcp_windows_arm64.exe'));
    if (failure === 'duplicate') {
      const file = path.join(root, 'dist/checksums.txt');
      const content = await readFile(file, 'utf8');
      await writeFile(file, content + content);
    }
    assert.notEqual(prepare(root).status, 0, failure);
    await assert.rejects(readFile(path.join(root, 'npm/release-manifest.json')), { code: 'ENOENT' });
  }
});
