const fs = require('fs');
const path = require('path');

// Test: Check that edge cases with type names are handled properly
function testEmptyTypeNames() {
  const solFile = path.join(__dirname, 'empty_type_test/empty_type_test.sol');
  
  if (!fs.existsSync(solFile)) {
    console.error('❌ Test file not generated');
    process.exit(1);
  }
  
  const solContent = fs.readFileSync(solFile, 'utf8');
  
  // Check that the valid message is generated correctly
  if (!/struct\s+TestMessage\s*{/.test(solContent)) {
    console.error('❌ Valid message TestMessage not generated');
    process.exit(1);
  }
  
  // Check that short message name is handled correctly
  if (!/struct\s+A\s*{/.test(solContent)) {
    console.error('❌ Short message name A not generated correctly');
    process.exit(1);
  }
  
  // Check that the valid fields are present
  if (!/string\s+field1;/.test(solContent)) {
    console.error('❌ Valid field field1 not found');
    process.exit(1);
  }
  
  if (!/uint32\s+field2;/.test(solContent)) {
    console.error('❌ Valid field field2 not found');
    process.exit(1);
  }
  
  if (!/string\s+short_name_field;/.test(solContent)) {
    console.error('❌ Valid field short_name_field not found');
    process.exit(1);
  }
  
  // Check that codec libraries are generated for both messages
  if (!/library\s+TestMessageCodec\s*{/.test(solContent)) {
    console.error('❌ TestMessageCodec library not generated');
    process.exit(1);
  }
  
  if (!/library\s+ACodec\s*{/.test(solContent)) {
    console.error('❌ ACodec library not generated');
    process.exit(1);
  }
  
  console.log('✅ Edge cases with type names handled properly');
}

// Run the test
testEmptyTypeNames(); 