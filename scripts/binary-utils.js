const os = require('os');

function getBinaryInfo() {
  const platform = os.platform();
  const arch = os.arch();
  
  let binaryName = 'kirha-mcp-installer';
  let ext = '';
  
  // Map Node.js arch to Go arch
  let goArch = arch;
  if (arch === 'x64') goArch = 'amd64';
  
  if (platform === 'win32') {
    binaryName += `-windows-${goArch}`;
    ext = '.exe';
  } else if (platform === 'darwin') {
    binaryName += `-darwin-${goArch}`;
  } else if (platform === 'linux') {
    binaryName += `-linux-${goArch}`;
  } else {
    console.error(`Unsupported platform: ${platform}`);
    process.exit(1);
  }
  
  return {
    name: binaryName + ext,
    checksumName: binaryName + ext + '.sha256'
  };
}

function getBinaryName() {
  return getBinaryInfo().name;
}

module.exports = {
  getBinaryInfo,
  getBinaryName
};