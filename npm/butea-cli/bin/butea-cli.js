#!/usr/bin/env node
'use strict';

const { spawnSync } = require('child_process');
const path = require('path');
const os = require('os');

const PLATFORM = os.platform(); // darwin, linux, win32
const ARCH = os.arch();         // arm64, x64

const PACKAGE_MAP = {
  'darwin-arm64':  { pkg: 'butea-cli-darwin-arm64',  bin: 'butea-cli' },
  'darwin-x64':    { pkg: 'butea-cli-darwin-x64',    bin: 'butea-cli' },
  'linux-arm64':   { pkg: 'butea-cli-linux-arm64',   bin: 'butea-cli' },
  'linux-x64':     { pkg: 'butea-cli-linux-x64',     bin: 'butea-cli' },
  'win32-arm64':   { pkg: 'butea-cli-windows-arm64', bin: 'butea-cli.exe' },
  'win32-x64':     { pkg: 'butea-cli-windows-x64',   bin: 'butea-cli.exe' },
};

const key = `${PLATFORM}-${ARCH}`;
const entry = PACKAGE_MAP[key];

if (!entry) {
  console.error(
    `butea-cli: unsupported platform/architecture: ${PLATFORM}/${ARCH}\n` +
    `Supported: ${Object.keys(PACKAGE_MAP).join(', ')}`
  );
  process.exit(1);
}

let binaryPath;
try {
  binaryPath = require.resolve(`${entry.pkg}/bin/${entry.bin}`);
} catch {
  console.error(
    `butea-cli: could not find binary for ${PLATFORM}/${ARCH}.\n` +
    `Try reinstalling: npm install -g butea-cli`
  );
  process.exit(1);
}

const result = spawnSync(binaryPath, process.argv.slice(2), { stdio: 'inherit' });

if (result.error) {
  console.error(`butea-cli: failed to run binary: ${result.error.message}`);
  process.exit(1);
}

process.exit(result.status ?? 1);
