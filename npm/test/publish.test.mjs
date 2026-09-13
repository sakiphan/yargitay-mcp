// SPDX-License-Identifier: AGPL-3.0-only
import assert from 'node:assert/strict';
import { test } from 'node:test';
import { mkdtemp, writeFile, rm } from 'node:fs/promises';
import { createHash } from 'node:crypto';
import { tmpdir } from 'node:os';
import path from 'node:path';
import { verifyPublish } from '../scripts/verify-publish.mjs';

async function fixture(t) {
  const directory = await mkdtemp(path.join(tmpdir(), 'yargitay-publish-test-'));
  t.after(() => rm(directory, { recursive: true, force: true }));
  const pkg = { name: 'sakiphan-yargitay-mcp', version: '0.1.0', license: 'AGPL-3.0-only', publishConfig: { registry: 'https://registry.npmjs.org/', access: 'public' } };
  const manifest = { version: '0.1.0', repository: 'sakiphan/yargitay-mcp', assets: {} };
  const lines = [];
  for (const os of ['darwin', 'linux', 'windows']) for (const arch of ['amd64', 'arm64']) {
    const name = `yargitay-mcp_${os}_${arch}${os === 'windows' ? '.exe' : ''}`;
    const body = Buffer.from(`synthetic ${name}`);
    const sha256 = createHash('sha256').update(body).digest('hex');
    manifest.assets[name] = { size: body.length, sha256 };
    lines.push(`${sha256}  ${name}`);
    await writeFile(path.join(directory, name), body);
  }
  await writeFile(path.join(directory, 'release.json'), '{"version":"0.1.0"}');
  await writeFile(path.join(directory, 'checksums.txt'), lines.join('\n'));
  return { directory, pkg, manifest, readPacked: (_, member) => member === 'package.json' ? pkg : manifest };
}

test('publish verifier accepts only matching package, release and six assets', async t => {
  const f = await fixture(t);
  assert.equal(await verifyPublish('v0.1.0', f.directory, f.readPacked), path.join(f.directory, 'sakiphan-yargitay-mcp-0.1.0.tgz'));
});
test('publish verifier rejects identity, version, registry, hash and missing-asset failures', async t => {
  for (const failure of ['name', 'version', 'manifest', 'repository', 'registry', 'license', 'build', 'checksum', 'corrupt', 'missing']) {
    const f = await fixture(t);
    if (failure === 'name') f.pkg.name = 'wrong-package';
    if (failure === 'version') f.pkg.version = '0.2.0';
    if (failure === 'manifest') f.manifest.version = '0.2.0';
    if (failure === 'repository') f.manifest.repository = 'someone/else';
    if (failure === 'registry') f.pkg.publishConfig.registry = 'https://other.example/';
    if (failure === 'license') f.pkg.license = 'MIT';
    if (failure === 'build') await writeFile(path.join(f.directory, 'release.json'), '{"version":"0.2.0"}');
    if (failure === 'checksum') f.manifest.assets['yargitay-mcp_linux_amd64'].sha256 = '0'.repeat(64);
    if (failure === 'corrupt') await writeFile(path.join(f.directory, 'yargitay-mcp_linux_amd64'), 'corrupt');
    if (failure === 'missing') await rm(path.join(f.directory, 'yargitay-mcp_linux_amd64'));
    await assert.rejects(verifyPublish('v0.1.0', f.directory, f.readPacked), undefined, failure);
  }
});
test('publish verifier rejects prerelease and unsafe tags before opening assets', async () => {
  for (const tag of ['v0.1.0-beta.1', '../../x', '0.1.0', 'v1.0.0; echo x']) {
    await assert.rejects(verifyPublish(tag, 'unused', () => assert.fail('Must not read assets')));
  }
});
