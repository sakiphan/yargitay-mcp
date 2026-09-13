// SPDX-License-Identifier: AGPL-3.0-only
import assert from 'node:assert/strict';
import { test } from 'node:test';
import { createHash } from 'node:crypto';
import { EventEmitter } from 'node:events';
import { mkdtemp, writeFile, readFile, rm, symlink, readdir } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import path from 'node:path';
import { Readable } from 'node:stream';
import { selectAsset, cacheRoot, ensureBinary, runBinary, downloadFile, validateDownloadURL, MAX_BINARY_BYTES } from '../lib/launcher.mjs';

const fixture = Buffer.from('synthetic executable fixture, no court data');
const hash = createHash('sha256').update(fixture).digest('hex');
function manifest() {
  const m = { version: '0.1.0', repository: 'sakiphan/yargitay-mcp', assets: {} };
  for (const os of ['darwin', 'linux', 'windows']) for (const arch of ['amd64', 'arm64']) m.assets[`yargitay-mcp_${os}_${arch}${os === 'windows' ? '.exe' : ''}`] = { sha256: hash, size: fixture.length };
  return m;
}
async function temporary(t) { const dir = await mkdtemp(path.join(tmpdir(), 'yargitay-npm-test-')); t.after(() => rm(dir, { recursive: true, force: true })); return dir; }

test('selects six supported platform binaries with exact version URLs', () => {
  for (const [platform, os] of [['darwin', 'darwin'], ['linux', 'linux'], ['win32', 'windows']]) {
    for (const [architecture, arch] of [['x64', 'amd64'], ['arm64', 'arm64']]) {
      const selected = selectAsset(manifest(), platform, architecture);
      assert.equal(selected.name, `yargitay-mcp_${os}_${arch}${os === 'windows' ? '.exe' : ''}`);
      assert.equal(selected.url, `https://github.com/sakiphan/yargitay-mcp/releases/download/v0.1.0/${selected.name}`);
    }
  }
});
test('rejects unsupported platform, altered repository, version and hashes', () => {
  assert.throws(() => selectAsset(manifest(), 'linux', 'ia32'), /Desteklenmeyen/);
  assert.throws(() => selectAsset(manifest(), 'freebsd', 'x64'), /Desteklenmeyen/);
  for (const m of [{ ...manifest(), version: '../evil' }, { ...manifest(), repository: 'someone/else' }, { ...manifest(), assets: {} }]) assert.throws(() => selectAsset(m, 'linux', 'x64'));
  const invalidHash = manifest(); invalidHash.assets['yargitay-mcp_linux_amd64'].sha256 = 'wrong';
  assert.throws(() => selectAsset(invalidHash, 'linux', 'x64'));
  const oversized = manifest(); oversized.assets['yargitay-mcp_linux_amd64'].size = MAX_BINARY_BYTES + 1;
  assert.throws(() => selectAsset(oversized, 'linux', 'x64'));
});
test('cache paths respect platform defaults and reject relative explicit overrides', () => {
  assert.equal(cacheRoot('linux', {}, '/users/example'), path.join('/users/example', '.cache', 'sakiphan-yargitay-mcp'));
  assert.equal(cacheRoot('darwin', {}, '/users/example'), path.join('/users/example', 'Library', 'Caches', 'sakiphan-yargitay-mcp'));
  assert.throws(() => cacheRoot('linux', { YARGITAY_MCP_CACHE_DIR: '../relative' }, '/users/example'));
});
test('downloads once, reuses verified cache, and repairs corrupt bytes', async t => {
  const cacheDirectory = await temporary(t); let calls = 0;
  const options = { cacheDirectory, platform: 'linux', architecture: 'x64', download: async (url, file) => { calls++; assert.match(url, /releases\/download\/v0\.1\.0/); await writeFile(file, fixture); } };
  const binary = await ensureBinary(manifest(), options);
  assert.deepEqual(await readFile(binary), fixture);
  assert.equal(await ensureBinary(manifest(), options), binary); assert.equal(calls, 1);
  await writeFile(binary, Buffer.alloc(fixture.length, 1));
  await ensureBinary(manifest(), options); assert.equal(calls, 2);
});
test('hash mismatch and failed downloads cannot install a binary; temp files are cleaned', async t => {
  for (const kind of ['hash', 'network']) {
    const cacheDirectory = await temporary(t);
    await assert.rejects(ensureBinary(manifest(), { cacheDirectory, platform: 'linux', architecture: 'x64', download: async (_, file) => { await writeFile(file, Buffer.alloc(fixture.length)); if (kind === 'network') throw new Error('network'); } }));
    assert.deepEqual(await readdir(path.join(cacheDirectory, '0.1.0', hash)), []);
  }
});
test('simultaneous cold launches converge on one verified cache entry', async t => {
  const cacheDirectory = await temporary(t);
  const options = { cacheDirectory, platform: process.platform, architecture: process.arch, download: async (_, file) => writeFile(file, fixture) };
  const results = await Promise.all(Array.from({ length: 4 }, () => ensureBinary(manifest(), options)));
  assert.equal(new Set(results).size, 1);
  assert.deepEqual(await readFile(results[0]), fixture);
  assert.deepEqual(await readdir(path.dirname(results[0])), [path.basename(results[0])]);
});
test('never executes a cache symlink', async t => {
  if (process.platform === 'win32') return t.skip('symlink privileges vary on Windows');
  const cacheDirectory = await temporary(t);
  const options = { cacheDirectory, platform: 'linux', architecture: 'x64', download: async (_, file) => writeFile(file, fixture) };
  const binary = await ensureBinary(manifest(), options); const other = path.join(cacheDirectory, 'other');
  await writeFile(other, fixture); await rm(binary); await symlink(other, binary);
  await assert.rejects(ensureBinary(manifest(), options), /normal dosya/);
});
test('runs native process without shell, forwards arguments, exit status and signals', async () => {
  const child = new EventEmitter(); const signals = new EventEmitter(); const forwarded = []; child.kill = signal => forwarded.push(signal);
  let observed;
  const result = runBinary('/tmp/native file', ['serve', '--viewer'], { signals, spawnProcess: (...args) => { observed = args; return child; } });
  assert.deepEqual(observed, ['/tmp/native file', ['serve', '--viewer'], { stdio: 'inherit', shell: false, windowsHide: true }]);
  signals.emit('SIGINT'); signals.emit('SIGTERM'); assert.deepEqual(forwarded, ['SIGINT', 'SIGTERM']);
  child.emit('exit', 7, null); assert.equal(await result, 7); assert.equal(signals.listenerCount('SIGINT'), 0);
});
test('default command is serve, spawn failures remain stderr-safe', async () => {
  const child = new EventEmitter(); child.kill = () => {}; const signals = new EventEmitter();
  const result = runBinary('/tmp/native', [], { signals, spawnProcess: (_, args) => { assert.deepEqual(args, ['serve']); return child; } });
  child.emit('error', new Error('PRIVATE detail')); await assert.rejects(result, error => !error.message.includes('PRIVATE'));
});

function mockRequest(responses, visited = []) {
  return (url, options, callback) => {
    assert.equal(options.rejectUnauthorized, true);
    visited.push(String(url)); const req = new EventEmitter(); req.setTimeout = () => req;
    req.destroy = err => req.emit('error', err);
    queueMicrotask(() => {
      const next = responses.shift();
      if (next instanceof Error) return req.emit('error', next);
      const res = Readable.from([next.body ?? fixture]); res.statusCode = next.status ?? 200; res.headers = next.headers ?? {};
      callback(res);
    });
    return req;
  };
}
test('downloads over HTTPS, follows allowed asset redirect, streams to disk', async t => {
  const dir = await temporary(t); const destination = path.join(dir, 'download'); const visited = [];
  await downloadFile('https://github.com/sakiphan/yargitay-mcp/releases/download/v0.1.0/file', destination, { request: mockRequest([{ status: 302, headers: { location: 'https://release-assets.githubusercontent.com/asset' } }, {}], visited) });
  assert.equal(visited.length, 2); assert.deepEqual(await readFile(destination), fixture);
});
test('rejects insecure URLs, redirects to unrelated hosts and excessive redirects', async t => {
  for (const url of ['http://github.com/a', 'https://evil.example/a', 'https://user:pass@github.com/a', 'https://github.com:8080/a']) assert.throws(() => validateDownloadURL(url));
  const dir = await temporary(t);
  await assert.rejects(downloadFile('https://github.com/a', path.join(dir, 'one'), { request: mockRequest([{ status: 302, headers: { location: 'https://evil.example/a' } }]) }));
  await assert.rejects(downloadFile('https://github.com/a', path.join(dir, 'two'), { request: mockRequest(Array.from({ length: 6 }, () => ({ status: 302, headers: { location: 'https://github.com/again' } }))) }), /yönlendirme sınırı/);
});
test('404, advertised oversize and network errors are safe failures', async t => {
  const dir = await temporary(t);
  for (const response of [{ status: 404 }, { headers: { 'content-length': String(MAX_BINARY_BYTES + 1) } }, new Error('PRIVATE signed URL')]) {
    await assert.rejects(downloadFile('https://github.com/a', path.join(dir, 'download'), { request: mockRequest([response]) }), error => !error.message.includes('PRIVATE'));
  }
});
test('total download deadline aborts a stalled request', async t => {
  const dir = await temporary(t);
  const request = (_, options) => { const req = new EventEmitter(); req.setTimeout = () => req; options.signal.addEventListener('abort', () => req.emit('error', new Error('aborted')), { once: true }); return req; };
  await assert.rejects(downloadFile('https://github.com/a', path.join(dir, 'download'), { request, deadlineMs: 20 }), /süre sınırı/);
});
