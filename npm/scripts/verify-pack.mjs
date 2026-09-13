// SPDX-License-Identifier: AGPL-3.0-only
import { readFile } from 'node:fs/promises';

const report = JSON.parse(await readFile(process.argv[2] ?? 'dist/npm-pack.json', 'utf8'));
const files = report[0]?.files?.map(item => item.path) ?? [];
for (const required of ['package.json', 'npm/bin/cli.mjs', 'npm/lib/launcher.mjs', 'npm/release-manifest.json', 'npm/THIRD_PARTY_NOTICES.txt', 'LICENSE', 'NOTICE', 'README.md', 'third_party/Apache-2.0.txt']) {
  if (!files.includes(required)) throw new Error(`Package missing ${required}`);
}
for (const file of files) {
  if (!(file.startsWith('npm/bin/') || file.startsWith('npm/lib/') || ['package.json', 'npm/release-manifest.json', 'npm/THIRD_PARTY_NOTICES.txt', 'LICENSE', 'NOTICE', 'README.md', 'third_party/Apache-2.0.txt', 'third_party/README.md'].includes(file))) throw new Error(`Unexpected packed file: ${file}`);
}
console.log(`Package verified: ${files.length} files; no binaries, fixtures or personal configuration.`);
