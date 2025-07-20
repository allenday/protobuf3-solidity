const fs = require('fs');
const path = require('path');

// Test: Check that reserved keywords are properly sanitized
function testReservedKeywordConflicts() {
  const solFile = path.join(__dirname, 'reserved_keyword_test/reserved_keyword_test.sol');
  
  if (!fs.existsSync(solFile)) {
    console.error('❌ Test file not generated');
    process.exit(1);
  }
  
  const solContent = fs.readFileSync(solFile, 'utf8');
  
  // Check that sanitized field names are used in struct definition (with underscore prefix)
  if (!/string\s+_seconds;/.test(solContent)) {
    console.error('❌ Sanitized field name "_seconds" not found in struct');
    process.exit(1);
  }
  
  if (!/uint32\s+_return;/.test(solContent)) {
    console.error('❌ Sanitized field name "_return" not found in struct');
    process.exit(1);
  }
  
  if (!/bool\s+_public;/.test(solContent)) {
    console.error('❌ Sanitized field name "_public" not found in struct');
    process.exit(1);
  }
  
  // Check that sanitized field names are used in codec functions
  if (!/instance\._seconds/.test(solContent)) {
    console.error('❌ Sanitized field name "_seconds" not used in codec');
    process.exit(1);
  }
  
  if (!/instance\._return/.test(solContent)) {
    console.error('❌ Sanitized field name "_return" not used in codec');
    process.exit(1);
  }
  
  if (!/instance\._public/.test(solContent)) {
    console.error('❌ Sanitized field name "_public" not used in codec');
    process.exit(1);
  }
  
  // Check that raw reserved keywords are NOT used (should be sanitized)
  if (solContent.includes('seconds;') && !solContent.includes('_seconds;')) {
    console.error('❌ Reserved keyword "seconds" not properly sanitized');
    process.exit(1);
  }
  
  if (solContent.includes('return;') && !solContent.includes('_return;')) {
    console.error('❌ Reserved keyword "return" not properly sanitized');
    process.exit(1);
  }
  
  if (solContent.includes('public;') && !solContent.includes('_public;')) {
    console.error('❌ Reserved keyword "public" not properly sanitized');
    process.exit(1);
  }
  
  console.log('✅ Reserved keywords properly sanitized');
}

// Run the test
testReservedKeywordConflicts(); 