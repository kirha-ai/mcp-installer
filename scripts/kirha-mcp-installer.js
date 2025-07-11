#!/usr/bin/env node

const { spawn } = require('child_process');
const path = require('path');
const os = require('os');
// Try to load binary-utils from the correct location
let getBinaryName;
try {
  // When in bin/ folder (after copying)
  getBinaryName = require('../scripts/binary-utils').getBinaryName;
} catch (e) {
  // When in scripts/ folder (development)
  getBinaryName = require('./binary-utils').getBinaryName;
}

function main() {
  const binaryName = getBinaryName();
  const binaryPath = path.join(__dirname, binaryName);
  
  // Pass all arguments to the Go binary
  const args = process.argv.slice(2);
  
  const child = spawn(binaryPath, args, {
    stdio: 'inherit',
    env: process.env
  });
  
  child.on('error', (error) => {
    if (error.code === 'ENOENT') {
      console.error(`Binary not found: ${binaryPath}`);
      console.error('Make sure the binary is built for your platform.');
      console.error(`Platform: ${os.platform()}, Architecture: ${os.arch()}`);
    } else {
      console.error('Failed to start kirha-mcp-installer:', error.message);
    }
    process.exit(1);
  });
  
  child.on('close', (code) => {
    process.exit(code);
  });
  
  // Handle signals
  process.on('SIGINT', () => {
    child.kill('SIGINT');
  });
  
  process.on('SIGTERM', () => {
    child.kill('SIGTERM');
  });
}

if (require.main === module) {
  main();
}