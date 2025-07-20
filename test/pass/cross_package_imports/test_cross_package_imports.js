#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

// Test cases for relative path verification
const testCases = [
  {
    name: 'postfiat/v3/messages.sol',
    expectedImports: [
      '@protobuf3-solidity-lib/contracts/ProtobufLib.sol',
      '../../a2a/v1/a2a.sol'
    ]
  },
  {
    name: 'a2a/v1/a2a.sol',
    expectedImports: [
      '@protobuf3-solidity-lib/contracts/ProtobufLib.sol'
    ]
  },
  {
    name: 'shared/common/common.sol',
    expectedImports: [
      '@protobuf3-solidity-lib/contracts/ProtobufLib.sol'
    ]
  },
  {
    name: 'deep/nested/package/test.sol',
    expectedImports: [
      '@protobuf3-solidity-lib/contracts/ProtobufLib.sol',
      '../../../a2a/v1/a2a.sol',
      '../../../shared/common.sol'
    ]
  }
];

function extractImports(solidityContent) {
  const importRegex = /import\s+"([^"]+)";/g;
  const imports = [];
  let match;
  
  while ((match = importRegex.exec(solidityContent)) !== null) {
    imports.push(match[1]);
  }
  
  return imports;
}

function testRelativePaths() {
  console.log('🧪 Testing cross-package relative import paths...\n');
  
  let allPassed = true;
  
  for (const testCase of testCases) {
    const filePath = path.join(__dirname, testCase.name);
    
    if (!fs.existsSync(filePath)) {
      console.log(`❌ FAIL: ${testCase.name} - File not found`);
      allPassed = false;
      continue;
    }
    
    const content = fs.readFileSync(filePath, 'utf8');
    const actualImports = extractImports(content);
    
    console.log(`📁 Testing ${testCase.name}:`);
    console.log(`   Expected imports: ${testCase.expectedImports.join(', ')}`);
    console.log(`   Actual imports:   ${actualImports.join(', ')}`);
    
    // Check if all expected imports are present
    const missingImports = testCase.expectedImports.filter(imp => !actualImports.includes(imp));
    const unexpectedImports = actualImports.filter(imp => !testCase.expectedImports.includes(imp));
    
    if (missingImports.length > 0) {
      console.log(`   ❌ Missing imports: ${missingImports.join(', ')}`);
      allPassed = false;
    }
    
    if (unexpectedImports.length > 0) {
      console.log(`   ❌ Unexpected imports: ${unexpectedImports.join(', ')}`);
      allPassed = false;
    }
    
    if (missingImports.length === 0 && unexpectedImports.length === 0) {
      console.log(`   ✅ PASS`);
    }
    
    console.log('');
  }
  
  if (allPassed) {
    console.log('🎉 All cross-package import tests passed!');
    process.exit(0);
  } else {
    console.log('💥 Some cross-package import tests failed!');
    process.exit(1);
  }
}

// Regression test: Ensure no codec library is generated for imported types
function testNoCodecForImportedTypes() {
  const fs = require('fs');
  const path = require('path');
  const solFile = path.join(__dirname, 'postfiat/v3/messages.sol');
  const solContent = fs.readFileSync(solFile, 'utf8');

  // Should NOT contain: library A2AMessageCodec {
  if (/library\s+A2AMessageCodec\s*{/.test(solContent)) {
    console.error('❌ Regression: Found codec library for imported type A2AMessage in postfiat/v3/messages.sol');
    process.exit(1);
  } else {
    console.log('✅ Regression: No codec library for imported type A2AMessage in postfiat/v3/messages.sol');
  }
}

// Regression test: Ensure all local message types have both struct definitions and codec libraries
function testStructAndCodecConsistency() {
  const fs = require('fs');
  const path = require('path');
  const solFile = path.join(__dirname, 'postfiat/v3/messages.sol');
  const solContent = fs.readFileSync(solFile, 'utf8');

  // Check for local message types that should have both structs and codecs
  const localMessageTypes = [
    'GetAgentCardRequest', 
    'GetAgentCardResponse'
  ];
  
  for (const messageType of localMessageTypes) {
    const hasStruct = new RegExp(`struct\\s+${messageType}\\s*{`).test(solContent);
    const hasCodec = new RegExp(`library\\s+${messageType}Codec\\s*{`).test(solContent);
    
    if (!hasStruct && hasCodec) {
      console.error(`❌ Regression: Missing struct definition for ${messageType} but codec library exists`);
      process.exit(1);
    } else if (hasStruct && !hasCodec) {
      console.error(`❌ Regression: Missing codec library for ${messageType} but struct definition exists`);
      process.exit(1);
    } else if (!hasStruct && !hasCodec) {
      console.error(`❌ Regression: Missing both struct definition and codec library for ${messageType}`);
      process.exit(1);
    } else {
      console.log(`✅ Regression: ${messageType} has both struct definition and codec library`);
    }
  }
}

testNoCodecForImportedTypes();
testStructAndCodecConsistency();

// Run the test
testRelativePaths(); 