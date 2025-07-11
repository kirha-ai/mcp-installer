#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const os = require('os');
const { exec } = require('child_process');

const CACHE_DIR = path.join(os.tmpdir(), 'kirha-mcp-installer-cache');
const VERSION_CACHE_FILE = path.join(CACHE_DIR, 'version.json');
const binDir = path.join(__dirname, '..', 'bin');

async function cleanCache() {
  if (fs.existsSync(VERSION_CACHE_FILE)) {
    fs.unlinkSync(VERSION_CACHE_FILE);
  }
  if (fs.existsSync(CACHE_DIR)) {
    fs.rmdirSync(CACHE_DIR);
  }
}

async function removeBinary() {
  const binaryPath = path.join(binDir, 'kirha-mcp-installer-darwin-arm64');
  if (fs.existsSync(binaryPath)) {
    fs.unlinkSync(binaryPath);
  }
}

async function timeCommand(command, description) {
  return new Promise((resolve, reject) => {
    const start = Date.now();
    exec(command, (error, stdout, stderr) => {
      const duration = Date.now() - start;
      console.log(`${description}: ${duration}ms`);
      if (error) {
        reject(error);
      } else {
        resolve({ duration, stdout, stderr });
      }
    });
  });
}

async function runPerformanceTest() {
  console.log('🚀 Performance Test - NPX MCP Installer Optimizations\n');
  
  // Test 1: Cold start (no cache, no binary)
  console.log('📦 Test 1: Cold start (first install)');
  await cleanCache();
  await removeBinary();
  const coldResult = await timeCommand('node scripts/install.js', 'Cold start');
  
  // Test 2: Warm start (cache exists, binary exists)
  console.log('\n📦 Test 2: Warm start (binary exists)');
  const warmResult = await timeCommand('node scripts/install.js', 'Warm start');
  
  // Test 3: Cache hit (binary removed, cache exists)
  console.log('\n📦 Test 3: Cache hit (binary removed, cache exists)');
  await removeBinary();
  const cacheResult = await timeCommand('node scripts/install.js', 'Cache hit');
  
  // Summary
  console.log('\n📊 Performance Summary:');
  console.log('─'.repeat(40));
  console.log(`Cold start (full download): ${coldResult.duration}ms`);
  console.log(`Warm start (skip all):      ${warmResult.duration}ms`);
  console.log(`Cache hit (skip API):       ${cacheResult.duration}ms`);
  console.log('─'.repeat(40));
  
  const speedup = Math.round((coldResult.duration / warmResult.duration) * 100) / 100;
  console.log(`🎯 Speed improvement: ${speedup}x faster on subsequent installs`);
  
  const cacheSpeedup = Math.round((coldResult.duration / cacheResult.duration) * 100) / 100;
  console.log(`📈 Cache benefit: ${cacheSpeedup}x faster with API cache`);
}

if (require.main === module) {
  runPerformanceTest().catch(console.error);
}