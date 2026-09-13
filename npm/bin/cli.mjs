#!/usr/bin/env node
// SPDX-License-Identifier: AGPL-3.0-only
import { readFile } from 'node:fs/promises';
import { ensureBinary, runBinary } from '../lib/launcher.mjs';

try {
  const manifest = JSON.parse(await readFile(new URL('../release-manifest.json', import.meta.url), 'utf8'));
  const packageInfo = JSON.parse(await readFile(new URL('../../package.json', import.meta.url), 'utf8'));
  if (manifest.version !== packageInfo.version) throw new Error('Paket ve binary sürümü uyuşmuyor.');
  const binary = await ensureBinary(manifest);
  process.exitCode = await runBinary(binary, process.argv.slice(2));
} catch (error) {
  // No protocol output on stdout; downloads/errors are diagnostics on stderr only.
  console.error(`sakiphan-yargitay-mcp: ${error.message}`);
  process.exitCode = 1;
}
