#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const os = require('os');
const https = require('https');
const crypto = require('crypto');
const { getBinaryInfo } = require('./binary-utils');

const GITHUB_REPO = 'kirha-ai/mcp-installer';
const GITHUB_API_URL = `https://api.github.com/repos/${GITHUB_REPO}/releases/latest`;

// Cache directory for storing version info
const CACHE_DIR = path.join(os.tmpdir(), 'kirha-mcp-installer-cache');
const VERSION_CACHE_FILE = path.join(CACHE_DIR, 'version.json');
const CACHE_DURATION = 24 * 60 * 60 * 1000; // 24 hours in milliseconds

function downloadFile(url, destination) {
  return new Promise((resolve, reject) => {
    console.log(`Downloading ${url}`);
    const file = fs.createWriteStream(destination);
    
    const request = https.get(url, {
      headers: {
        'User-Agent': 'kirha-mcp-installer-npm-installer'
      }
    }, (response) => {
      if (response.statusCode === 302 || response.statusCode === 301) {
        // Follow redirect
        return downloadFile(response.headers.location, destination)
          .then(resolve)
          .catch(reject);
      }
      
      if (response.statusCode !== 200) {
        reject(new Error(`Failed to download: ${response.statusCode} ${response.statusMessage}`));
        return;
      }
      
      response.pipe(file);
      
      file.on('finish', () => {
        file.close();
        resolve();
      });
      
      file.on('error', (error) => {
        fs.unlink(destination, () => {}); // Delete partial file
        reject(error);
      });
    }).on('error', reject);
    
    // Ensure the request is properly closed
    request.on('error', reject);
  });
}

function getLatestRelease() {
  return new Promise((resolve, reject) => {
    https.get(GITHUB_API_URL, {
      headers: {
        'User-Agent': 'kirha-mcp-installer-npm-installer'
      }
    }, (response) => {
      let data = '';
      
      response.on('data', (chunk) => {
        data += chunk;
      });
      
      response.on('end', () => {
        if (response.statusCode !== 200) {
          reject(new Error(`GitHub API request failed: ${response.statusCode}`));
          return;
        }
        
        try {
          const release = JSON.parse(data);
          resolve(release);
        } catch (error) {
          reject(new Error(`Failed to parse GitHub API response: ${error.message}`));
        }
      });
    }).on('error', reject);
  });
}

function verifyChecksum(filePath, checksumPath) {
  if (!fs.existsSync(checksumPath)) {
    console.warn('Checksum file not found, skipping verification');
    return true;
  }
  
  const expectedChecksum = fs.readFileSync(checksumPath, 'utf8').trim().split(' ')[0];
  const fileBuffer = fs.readFileSync(filePath);
  const actualChecksum = crypto.createHash('sha256').update(fileBuffer).digest('hex');
  
  if (actualChecksum !== expectedChecksum) {
    console.error(`Checksum verification failed!`);
    console.error(`Expected: ${expectedChecksum}`);
    console.error(`Actual: ${actualChecksum}`);
    return false;
  }
  
  console.log('✅ Checksum verification passed');
  return true;
}

function getCachedVersionInfo() {
  try {
    if (!fs.existsSync(VERSION_CACHE_FILE)) {
      return null;
    }
    
    const cacheData = JSON.parse(fs.readFileSync(VERSION_CACHE_FILE, 'utf8'));
    const now = Date.now();
    
    if (now - cacheData.timestamp > CACHE_DURATION) {
      return null; // Cache expired
    }
    
    return cacheData.release;
  } catch (error) {
    return null;
  }
}

function setCachedVersionInfo(release) {
  try {
    if (!fs.existsSync(CACHE_DIR)) {
      fs.mkdirSync(CACHE_DIR, { recursive: true });
    }
    
    const cacheData = {
      timestamp: Date.now(),
      release: release
    };
    
    fs.writeFileSync(VERSION_CACHE_FILE, JSON.stringify(cacheData, null, 2));
  } catch (error) {
    console.warn('Failed to cache version info:', error.message);
  }
}

async function main() {
  const binaryInfo = getBinaryInfo();
  const binDir = path.join(__dirname, '..', 'bin');
  const binaryPath = path.join(binDir, binaryInfo.name);
  const checksumPath = path.join(binDir, binaryInfo.checksumName);
  
  console.log(`Installing kirha-mcp-installer for ${os.platform()} ${os.arch()}`);
  console.log(`Binary: ${binaryInfo.name}`);
  
  // Create bin directory if it doesn't exist
  if (!fs.existsSync(binDir)) {
    fs.mkdirSync(binDir, { recursive: true });
  }
  
  // Check if binary already exists and is executable
  if (fs.existsSync(binaryPath)) {
    console.log('Binary already exists, skipping download');
    
    // Make binary executable on Unix systems
    if (os.platform() !== 'win32') {
      try {
        fs.chmodSync(binaryPath, '755');
      } catch (error) {
        console.warn(`Failed to make binary executable: ${error.message}`);
      }
    }
    
    console.log(`✅ kirha-mcp-installer is ready for ${os.platform()} ${os.arch()}`);
    process.exit(0);
  }
  
  try {
    // Try to get release info from cache first
    let release = getCachedVersionInfo();
    
    if (!release) {
      console.log('Fetching latest release information...');
      release = await getLatestRelease();
      setCachedVersionInfo(release);
    } else {
      console.log('Using cached release information...');
    }
    
    // Find the binary asset
    const binaryAsset = release.assets.find(asset => asset.name === binaryInfo.name);
    const checksumAsset = release.assets.find(asset => asset.name === binaryInfo.checksumName);
    
    if (!binaryAsset) {
      console.error(`Binary not found for your platform: ${binaryInfo.name}`);
      console.error(`Platform: ${os.platform()}, Architecture: ${os.arch()}`);
      console.error('Available binaries:');
      release.assets.forEach(asset => {
        if (asset.name.startsWith('kirha-mcp-installer-')) {
          console.error(`  - ${asset.name}`);
        }
      });
      process.exit(1);
    }
    
    // Download binary
    await downloadFile(binaryAsset.browser_download_url, binaryPath);
    console.log(`✅ Downloaded ${binaryInfo.name}`);
    
    // Download and verify checksum if available (optional for faster installation)
    if (checksumAsset && process.env.KIRHA_VERIFY_CHECKSUM !== 'false') {
      console.log('Downloading and verifying checksum...');
      await downloadFile(checksumAsset.browser_download_url, checksumPath);
      
      if (!verifyChecksum(binaryPath, checksumPath)) {
        fs.unlinkSync(binaryPath);
        fs.unlinkSync(checksumPath);
        console.error('Installation failed due to checksum mismatch');
        process.exit(1);
      }
      
      // Clean up checksum file
      fs.unlinkSync(checksumPath);
    } else if (checksumAsset) {
      console.log('Skipping checksum verification for faster installation (set KIRHA_VERIFY_CHECKSUM=true to enable)');
    }
    
    // Make binary executable on Unix systems
    if (os.platform() !== 'win32') {
      try {
        fs.chmodSync(binaryPath, '755');
        console.log(`Made ${binaryInfo.name} executable`);
      } catch (error) {
        console.warn(`Failed to make binary executable: ${error.message}`);
      }
    }
    
    
    console.log(`✅ kirha-mcp-installer ${release.tag_name} is ready for ${os.platform()} ${os.arch()}`);
    process.exit(0);
    
  } catch (error) {
    console.error(`Installation failed: ${error.message}`);
    
    // Clean up partial downloads
    if (fs.existsSync(binaryPath)) {
      fs.unlinkSync(binaryPath);
    }
    if (fs.existsSync(checksumPath)) {
      fs.unlinkSync(checksumPath);
    }
    
    console.error('\nTroubleshooting:');
    console.error('1. Check your internet connection');
    console.error('2. Verify the GitHub repository exists and has releases');
    console.error('3. Try installing again later');
    console.error('4. Download manually from: https://go.kirha.ai/kirha-mcp-installer/releases');
    console.error(`5. Place the binary manually in: ${binaryPath}`);
    
    process.exit(1);
  }
}

if (require.main === module) {
  main().catch(console.error);
}