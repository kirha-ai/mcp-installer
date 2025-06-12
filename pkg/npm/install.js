const fs = require('fs');
const path = require('path');
const os = require('os');

function checkBinary() {
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
    console.warn(`Warning: Unsupported platform: ${platform}/${arch}`);
    return false;
  }
  
  const binaryName = platform === 'win32' ? 'mcp-installer.exe' : 'mcp-installer';
  const binaryPath = path.join(__dirname, 'binaries', `${goPlatform}_${goArch}`, binaryName);
  
  if (!fs.existsSync(binaryPath)) {
    console.warn(`Warning: Binary not found at ${binaryPath}`);
    console.warn('This may indicate an incomplete installation.');
    return false;
  }
  
  // Check if binary is executable (Unix-like systems)
  if (platform !== 'win32') {
    try {
      fs.accessSync(binaryPath, fs.constants.F_OK | fs.constants.X_OK);
    } catch (error) {
      console.warn(`Warning: Binary exists but is not executable: ${binaryPath}`);
      return false;
    }
  }
  
  console.log(`✓ MCP Installer binary ready for ${platform}/${arch}`);
  return true;
}

function main() {
  console.log('Installing MCP Installer...');
  
  if (checkBinary()) {
    console.log('Installation complete!');
    console.log('');
    console.log('Usage:');
    console.log('  npx @kirha/mcp-installer <command> --client <client> --key <api-key>');
    console.log('');
    console.log('Commands:');
    console.log('  install  - Install MCP server (fails if already exists)');
    console.log('  update   - Update existing MCP server configuration');
    console.log('  remove   - Remove MCP server from configuration');
    console.log('');
    console.log('Examples:');
    console.log('  npx @kirha/mcp-installer install --client claude --key your-api-key');
    console.log('  npx @kirha/mcp-installer update --client docker --key your-new-api-key');
    console.log('  npx @kirha/mcp-installer remove --client cursor');
    console.log('');
    console.log('For more help: npx @kirha/mcp-installer --help');
  } else {
    console.error('Installation failed: Binary not found or not executable');
    process.exit(1);
  }
}

main();