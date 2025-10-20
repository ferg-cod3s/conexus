#!/usr/bin/env node

import { spawn } from 'child_process';
import { fileURLToPath } from 'url';
import { dirname, join } from 'path';
import { platform, arch } from 'os';
import { existsSync } from 'fs';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

// Determine platform-specific binary name
function getBinaryPath() {
  const plat = platform();
  const architecture = arch();
  
  // Map Node.js arch to Go arch
  const goArch = architecture === 'x64' ? 'amd64' : architecture;
  
  let binaryName;
  if (plat === 'win32') {
    binaryName = `conexus-windows-${goArch}.exe`;
  } else if (plat === 'darwin') {
    binaryName = `conexus-darwin-${goArch}`;
  } else {
    binaryName = `conexus-linux-${goArch}`;
  }
  
  const binaryPath = join(__dirname, binaryName);
  
  // Fallback to built binary if platform-specific one doesn't exist
  if (!existsSync(binaryPath)) {
    const fallback = join(__dirname, '..', 'conexus');
    if (existsSync(fallback)) {
      return fallback;
    }
    return null;
  }
  
  return binaryPath;
}

const conexusBinary = getBinaryPath();

if (!conexusBinary) {
  console.error(`Error: No conexus binary found for ${platform()}-${arch()}`);
  console.error('Supported platforms: darwin-amd64, darwin-arm64, linux-amd64, linux-arm64, windows-amd64');
  process.exit(1);
}

const child = spawn(conexusBinary, process.argv.slice(2), {
  stdio: 'inherit',
  env: {
    ...process.env,
  }
});

child.on('error', (err) => {
  if (err.code === 'ENOENT') {
    console.error(`Error: conexus binary not found at ${conexusBinary}`);
    console.error('Please check the installation or file an issue at https://github.com/ferg-cod3s/conexus');
    process.exit(1);
  }
  console.error('Error starting conexus:', err);
  process.exit(1);
});

child.on('exit', (code) => {
  process.exit(code || 0);
});
