// Drives the js/wasm build of internal/sim under Node so the parity test can
// compare its output against the native build byte for byte.
//
// Reads a JSON array of {seed, decisions} cases on stdin and writes one
// result JSON per line to stdout, in the same order.

import { readFile } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';
import vm from 'node:vm';
import fs from 'node:fs';

const here = dirname(fileURLToPath(import.meta.url));
const pub = join(here, '..', 'web', 'public');

// wasm_exec.js is a classic script that defines a global `Go`. Run it in the
// current global context rather than importing it as a module.
const execSrc = await readFile(join(pub, 'wasm_exec.js'), 'utf8');
vm.runInThisContext(execSrc, { filename: 'wasm_exec.js' });

const go = new globalThis.Go();
const bytes = await readFile(join(pub, 'sim.wasm'));
const { instance } = await WebAssembly.instantiate(bytes, go.importObject);

// main() ends in select{} and never returns, so this promise never settles.
// The globals are installed during the synchronous part of startup.
go.run(instance).catch((err) => {
  process.stderr.write(`wasm exited: ${err}\n`);
  process.exit(1);
});

if (typeof globalThis.lightsOutRunSeason !== 'function') {
  process.stderr.write('wasm module did not install lightsOutRunSeason\n');
  process.exit(1);
}

const stdin = await new Promise((resolve, reject) => {
  const chunks = [];
  process.stdin.on('data', (c) => chunks.push(c));
  process.stdin.on('end', () => resolve(Buffer.concat(chunks).toString('utf8')));
  process.stdin.on('error', reject);
});

const cases = JSON.parse(stdin);
const out = [];
for (const c of cases) {
  // The seed crosses as a string: a JS number is float64 and would lose
  // precision above 2^53.
  out.push(globalThis.lightsOutRunSeason(String(c.seed), JSON.stringify(c.picks)));
}

// Write synchronously and handle partial writes. process.stdout.write
// followed by process.exit truncates on a pipe, because exit does not wait
// for an async flush -- which silently drops most of a multi-megabyte
// result set.
const buf = Buffer.from(out.join('\n') + '\n', 'utf8');
let written = 0;
while (written < buf.length) {
  written += fs.writeSync(1, buf, written, buf.length - written);
}
process.exit(0);
