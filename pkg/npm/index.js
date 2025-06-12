#!/usr/bin/env node

const { execFileSync } = require('child_process');
const path = require('path');
const os = require('os');

function getBinaryPath() {
  const platform = os.platform();
  const arch = os.arch();
  
  // Map Node.js arch to Go arch
  const archMapping = {
    'x64': 'amd64',
    'arm64': 'arm64'
  };
  
  // Map Node.js platform to Go platform
  const platformMapping = {
    'darwin': 'darwin',
    'linux': 'linux',
    'win32': 'windows'
  };
  
  const goPlatform = platformMapping[platform];
  const goArch = archMapping[arch];
  
  if (!goPlatform || !goArch) {
    console.error(`Unsupported platform: ${platform}/${arch}`);
    process.exit(1);
  }
  
  const binaryName = platform === 'win32' ? 'mcp-installer.exe' : 'mcp-installer';
  const binaryPath = path.join(__dirname, 'binaries', `${goPlatform}_${goArch}`, binaryName);
  
  return binaryPath;
}

function main() {
  try {
    const binaryPath = getBinaryPath();
    const args = process.argv.slice(2);
    
    // Execute the binary with the provided arguments
    execFileSync(binaryPath, args, {
      stdio: 'inherit',
      cwd: process.cwd()
    });
  } catch (error) {
    if (error.code === 'ENOENT') {
      console.error('Error: Binary not found. Please ensure the correct binary is installed for your platform.');
      console.error(`Platform: ${os.platform()}/${os.arch()}`);
      process.exit(1);
    } else if (error.status !== undefined) {
      // Binary executed but returned non-zero exit code
      process.exit(error.status);
    } else {
      console.error('Error executing mcp-installer:', error.message);
      process.exit(1);
    }
  }
}

main();