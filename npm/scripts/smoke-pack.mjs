// SPDX-License-Identifier: AGPL-3.0-only
// Offline POSIX integration: packed npx -> launcher -> real native MCP binary.
import assert from 'node:assert/strict';
import { mkdtemp, readFile, copyFile, chmod, rm } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import path from 'node:path';
import { spawn } from 'node:child_process';
import { createInterface } from 'node:readline';
import { ensureBinary } from '../lib/launcher.mjs';

if (process.platform === 'win32') throw new Error('Run packed integration on macOS or Linux; Windows launcher unit tests run separately.');
const pkg = JSON.parse(await readFile('package.json', 'utf8'));
const manifest = JSON.parse(await readFile('npm/release-manifest.json', 'utf8'));
const temporary = await mkdtemp(path.join(tmpdir(), 'yargitay-npx-smoke-'));
const binaryCache = path.join(temporary, 'binaries');
let child;
let deadline;
try {
  await ensureBinary(manifest, { cacheDirectory: binaryCache, download: async (url, target) => {
    await copyFile(path.join('dist', new URL(url).pathname.split('/').at(-1)), target);
    await chmod(target, 0o755);
  } });
  // Every npm/native write stays in this temporary directory; no personal client settings.
  child = spawn('npx', ['--offline', '--yes', '--cache', path.join(temporary, 'npm-cache'), '--package', path.resolve(`dist/${pkg.name}-${pkg.version}.tgz`), pkg.name], {
    env: { ...process.env, YARGITAY_MCP_CACHE_DIR: binaryCache }, stdio: ['pipe', 'pipe', 'pipe'], shell: false,
  });
  let stderr = '';
  child.stderr.setEncoding('utf8').on('data', chunk => { stderr += chunk; });
  const pending = new Map();
  let protocolError;
  const lines = createInterface({ input: child.stdout });
  lines.on('line', line => {
    try {
      const message = JSON.parse(line); // Download/npm noise on stdout fails the test.
      assert.equal(message.jsonrpc, '2.0');
      if (pending.has(message.id)) {
        const { resolve, reject } = pending.get(message.id);
        pending.delete(message.id);
        if (message.error) reject(new Error(JSON.stringify(message.error)));
        else resolve(message.result);
      }
    } catch (error) { protocolError = error; child.stdin.end(); }
  });
  const exited = new Promise((resolve, reject) => {
    child.once('error', reject);
    child.once('exit', (code, signal) => {
      for (const waiter of pending.values()) waiter.reject(new Error(`Premature exit: ${code}/${signal}: ${stderr}`));
      resolve(code);
    });
  });
  const timeout = new Promise((_, reject) => {
    deadline = setTimeout(() => { child.stdin.end(); child.kill('SIGTERM'); reject(new Error('Packed MCP smoke timed out.')); }, 30_000);
  });
  const request = (id, method, params = {}) => new Promise((resolve, reject) => {
    pending.set(id, { resolve, reject });
    child.stdin.write(JSON.stringify({ jsonrpc: '2.0', id, method, params }) + '\n');
  });
  await Promise.race([timeout, (async () => {
    const init = await request(1, 'initialize', { protocolVersion: '2025-03-26', capabilities: {}, clientInfo: { name: 'offline-npx-smoke', version: '1.0.0' } });
    assert.ok(init.serverInfo);
    child.stdin.write(JSON.stringify({ jsonrpc: '2.0', method: 'notifications/initialized' }) + '\n');
    const listing = await request(2, 'tools/list');
    assert.deepEqual(listing.tools.map(tool => tool.name).sort(), ['get_aihm_decision', 'get_aym_decision', 'get_danistay_decision', 'get_yargitay_decision', 'list_yargitay_units', 'search_aihm_decisions', 'search_aym_decisions', 'search_danistay_decisions', 'search_yargitay_decisions']);
    const units = await request(3, 'tools/call', { name: 'list_yargitay_units', arguments: {} });
    assert.ok(!units.isError);
    const data = units.structuredContent ?? JSON.parse(units.content.find(item => item.type === 'text').text);
    assert.equal(data.units.length, 51);
    child.stdin.end();
    assert.equal(await exited, 0, stderr);
    if (protocolError) throw protocolError;
  })()]);
  console.log('Offline packed npx smoke passed: initialize, nine tools, 51 units, clean stdout and exit.');
} finally {
  clearTimeout(deadline);
  if (child && child.exitCode === null) { child.stdin.end(); child.kill('SIGTERM'); }
  await rm(temporary, { recursive: true, force: true });
}
