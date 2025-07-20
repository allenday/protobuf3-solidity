const fs = require('fs');
const path = require('path');

// Test: Check that empty structs are properly handled for google.protobuf.Empty
function testEmptyStructGeneration() {
  const solFile = path.join(__dirname, 'empty_struct_test/empty_struct_test.sol');
  
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
  
  // Check that Empty struct is properly handled (should have placeholder field)
  if (!/struct\s+Empty\s*{/.test(solContent)) {
    console.error('❌ Empty struct not generated');
    process.exit(1);
  }
  
  // Check that Empty struct has a placeholder field to avoid compilation errors
  if (!/bool\s+_placeholder;/.test(solContent)) {
    console.error('❌ Empty struct missing placeholder field');
    process.exit(1);
  }
  
  // Check that the placeholder comment is present
  if (!/Placeholder field to avoid empty struct compilation error/.test(solContent)) {
    console.error('❌ Missing placeholder field comment');
    process.exit(1);
  }
  
  // Check that the valid message is generated correctly
  if (!/struct\s+TestMessage\s*{/.test(solContent)) {
    console.error('❌ Valid message TestMessage not generated');
    process.exit(1);
  }
  
  // Check that the TestMessage uses the Google protobuf type correctly
  if (!/Google_Protobuf\.Empty\s+empty_field;/.test(solContent)) {
    console.error('❌ TestMessage does not use Google_Protobuf.Empty correctly');
    process.exit(1);
  }
  
  console.log('✅ Empty structs properly handled');
}

// Run the test
testEmptyStructGeneration(); 