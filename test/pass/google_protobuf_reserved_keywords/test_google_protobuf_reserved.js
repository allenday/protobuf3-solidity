const fs = require('fs');
const path = require('path');

// Test: Check that reserved keywords in inline Google protobuf types are properly sanitized
function testGoogleProtobufReservedKeywords() {
  const solFile = path.join(__dirname, 'google_protobuf_reserved_test/google_protobuf_reserved_test.sol');
  
  if (!fs.existsSync(solFile)) {
    console.error('❌ Test file not generated');
    process.exit(1);
  }
  
  const solContent = fs.readFileSync(solFile, 'utf8');
  
  // Check that Google protobuf library is generated
  if (!/library\s+Google_Protobuf\s*{/.test(solContent)) {
    console.error('❌ Google_Protobuf library not generated');
    process.exit(1);
  }
  
  // Check that Timestamp struct is generated
  if (!/struct\s+Timestamp\s*{/.test(solContent)) {
    console.error('❌ Google_Protobuf.Timestamp struct not generated');
    process.exit(1);
  }
  
  // Check that the reserved keyword 'seconds' is properly sanitized in the inline definition
  if (solContent.includes('int64 seconds;') && !solContent.includes('int64 _seconds;')) {
    console.error('❌ Reserved keyword "seconds" not properly sanitized in inline Google protobuf type');
    process.exit(1);
  }
  
  // Check that the sanitized field name is used in the struct definition
  if (!/int64\s+_seconds;/.test(solContent)) {
    console.error('❌ Sanitized field name "_seconds" not found in Google_Protobuf.Timestamp');
    process.exit(1);
  }
  
  // Check that the valid message is generated correctly
  if (!/struct\s+TestMessage\s*{/.test(solContent)) {
    console.error('❌ Valid message TestMessage not generated');
    process.exit(1);
  }
  
  // Check that the TestMessage uses the Google protobuf type correctly
  if (!/Google_Protobuf\.Timestamp\s+created_at;/.test(solContent)) {
    console.error('❌ TestMessage does not use Google_Protobuf.Timestamp correctly');
    process.exit(1);
  }
  
  console.log('✅ Reserved keywords in inline Google protobuf types properly sanitized');
}

// Run the test
testGoogleProtobufReservedKeywords(); 