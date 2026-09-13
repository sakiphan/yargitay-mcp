// SPDX-License-Identifier: AGPL-3.0-only
import { createHash, randomUUID } from 'node:crypto';
import { createReadStream, createWriteStream } from 'node:fs';
import { chmod, lstat, mkdir, mkdtemp, rename, rm, unlink } from 'node:fs/promises';
import https from 'node:https';
import { homedir } from 'node:os';
import path from 'node:path';
import { spawn } from 'node:child_process';
import { Transform } from 'node:stream';
import { pipeline } from 'node:stream/promises';

const REPOSITORY = 'sakiphan/yargitay-mcp';
export const MAX_BINARY_BYTES = 128 * 1024 * 1024;
const DOWNLOAD_HOSTS = new Set(['github.com', 'release-assets.githubusercontent.com', 'objects.githubusercontent.com']);
const PLATFORMS = { darwin: 'darwin', linux: 'linux', win32: 'windows' };
const ARCHITECTURES = { x64: 'amd64', arm64: 'arm64' };

export function selectAsset(manifest, platform = process.platform, architecture = process.arch) {
  const os = PLATFORMS[platform];
  const arch = ARCHITECTURES[architecture];
  if (!os || !arch) throw new Error(`Desteklenmeyen platform: ${platform}/${architecture}. macOS/Linux/Windows, x64/arm64 gerekir.`);
  if (manifest.repository !== REPOSITORY || !/^\d+\.\d+\.\d+$/.test(manifest.version ?? '')) {
    throw new Error('Geçersiz release manifesti.');
  }
  const name = `yargitay-mcp_${os}_${arch}${os === 'windows' ? '.exe' : ''}`;
  const entry = manifest.assets?.[name];
  if (!entry || !/^[a-f0-9]{64}$/.test(entry.sha256) || !Number.isSafeInteger(entry.size) || entry.size <= 0 || entry.size > MAX_BINARY_BYTES) {
    throw new Error('Bu platform için doğrulanmış binary manifestte yok.');
  }
  return { name, sha256: entry.sha256, size: entry.size, url: `https://github.com/${REPOSITORY}/releases/download/v${manifest.version}/${name}` };
}

export function cacheRoot(platform = process.platform, env = process.env, home = homedir()) {
  const configured = env.YARGITAY_MCP_CACHE_DIR;
  if (configured) {
    if (!path.isAbsolute(configured)) throw new Error('YARGITAY_MCP_CACHE_DIR mutlak bir yol olmalı.');
    return configured;
  }
  if (platform === 'win32') return path.join(env.LOCALAPPDATA || path.join(home, 'AppData', 'Local'), 'sakiphan-yargitay-mcp', 'Cache');
  if (platform === 'darwin') return path.join(home, 'Library', 'Caches', 'sakiphan-yargitay-mcp');
  return path.join(env.XDG_CACHE_HOME && path.isAbsolute(env.XDG_CACHE_HOME) ? env.XDG_CACHE_HOME : path.join(home, '.cache'), 'sakiphan-yargitay-mcp');
}

export async function verifyBinary(file, asset) {
  let stat;
  try { stat = await lstat(file); } catch (error) { if (error.code === 'ENOENT') return false; throw error; }
  if (!stat.isFile() || stat.isSymbolicLink()) throw new Error('Binary önbelleği normal dosya değil; çalıştırılmadı.');
  if (stat.size !== asset.size || stat.size > MAX_BINARY_BYTES) return false;
  const hash = createHash('sha256');
  for await (const chunk of createReadStream(file)) hash.update(chunk);
  return hash.digest('hex') === asset.sha256;
}

export function validateDownloadURL(input) {
  const url = new URL(input);
  if (url.protocol !== 'https:' || url.username || url.password || (url.port && url.port !== '443') || !DOWNLOAD_HOSTS.has(url.hostname)) {
    throw new Error('İndirme adresi güvenilir GitHub HTTPS kaynağı değil.');
  }
  return url;
}

export async function downloadFile(input, destination, { request = https.get, deadlineMs = 120_000 } = {}) {
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), deadlineMs);
  try {
    let url = validateDownloadURL(input);
    for (let redirects = 0; redirects <= 5; redirects++) {
      const response = await new Promise((resolve, reject) => {
        const req = request(url, { signal: controller.signal, rejectUnauthorized: true, headers: { 'User-Agent': 'sakiphan-yargitay-mcp', Accept: 'application/octet-stream' } }, resolve);
        req.setTimeout(30_000, () => req.destroy(new Error('timeout')));
        req.on('error', reject);
      });
      if ([301, 302, 303, 307, 308].includes(response.statusCode)) {
        const location = response.headers.location;
        response.destroy();
        if (!location || redirects === 5) throw new Error('Release indirmesinde yönlendirme sınırı aşıldı.');
        url = validateDownloadURL(new URL(location, url));
        continue;
      }
      if (response.statusCode !== 200) {
        response.destroy();
        throw new Error(response.statusCode === 404 ? 'GitHub release binary’si bulunamadı; bu sürüm henüz yayımlanmamış olabilir.' : 'GitHub binary indirmesi başarısız.');
      }
      const length = response.headers['content-length'];
      if (length && (!/^\d+$/.test(length) || Number(length) > MAX_BINARY_BYTES)) {
        response.destroy();
        throw new Error('İndirilen binary boyut sınırını aşıyor.');
      }
      let received = 0;
      const limit = new Transform({ transform(chunk, encoding, done) {
        received += chunk.length;
        done(received > MAX_BINARY_BYTES ? new Error('Binary boyut sınırı aşıldı.') : null, chunk);
      } });
      await pipeline(response, limit, createWriteStream(destination, { flags: 'wx', mode: 0o700 }), { signal: controller.signal });
      return;
    }
  } catch (error) {
    if (controller.signal.aborted) throw new Error('Binary indirme süre sınırı aşıldı.');
    // Do not print response bodies, signed redirect URLs or raw network errors.
    if (/^(GitHub|Release|İndirme|İndirilen|Binary)/.test(error.message)) throw error;
    throw new Error('GitHub binary indirilemedi; bağlantıyı kontrol edip yeniden deneyin.');
  } finally { clearTimeout(timer); }
}

export async function ensureBinary(manifest, options = {}) {
  const platform = options.platform ?? process.platform;
  const asset = selectAsset(manifest, platform, options.architecture ?? process.arch);
  const root = options.cacheDirectory ?? cacheRoot();
  const directory = path.join(root, manifest.version, asset.sha256);
  const binary = path.join(directory, asset.name);
  if (await verifyBinary(binary, asset)) return binary;
  await mkdir(directory, { recursive: true, mode: 0o700 });
  const temporary = await mkdtemp(path.join(directory, '.download-'));
  const staged = path.join(temporary, asset.name);
  try {
    await (options.download ?? downloadFile)(asset.url, staged);
    if (!await verifyBinary(staged, asset)) throw new Error('SHA-256 veya boyut uyuşmuyor; binary çalıştırılmadı.');
    await chmod(staged, 0o755);
    // Concurrent launches each stage privately. An already-verified winner can be reused.
    if (await verifyBinary(binary, asset)) return binary;
    try { await rename(staged, binary); } catch (error) {
      if (await verifyBinary(binary, asset)) return binary;
      if (platform === 'win32' && ['EEXIST', 'EPERM'].includes(error.code)) {
        // Quarantine only this content-addressed cache entry, never a general directory.
        const quarantine = path.join(directory, `.invalid-${randomUUID()}`);
        await rename(binary, quarantine);
        await rename(staged, binary);
        await unlink(quarantine);
      } else { throw error; }
    }
    return binary;
  } finally {
    await rm(temporary, { recursive: true, force: true });
  }
}

export function runBinary(binary, args = [], { spawnProcess = spawn, signals = process } = {}) {
  return new Promise((resolve, reject) => {
    // Inherit stdin/stdout/stderr byte-for-byte. Never invoke a shell.
    const child = spawnProcess(binary, args.length ? args : ['serve'], { stdio: 'inherit', shell: false, windowsHide: true });
    const forwardInt = () => child.kill('SIGINT');
    const forwardTerm = () => child.kill('SIGTERM');
    signals.on('SIGINT', forwardInt);
    signals.on('SIGTERM', forwardTerm);
    const cleanup = () => { signals.removeListener('SIGINT', forwardInt); signals.removeListener('SIGTERM', forwardTerm); };
    child.once('error', () => { cleanup(); reject(new Error('Go binary başlatılamadı; dosya izinlerini kontrol edin.')); });
    child.once('exit', (code, signal) => { cleanup(); resolve(code ?? (signal === 'SIGINT' ? 130 : signal === 'SIGTERM' ? 143 : 1)); });
  });
}
