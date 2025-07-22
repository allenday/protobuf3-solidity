const fs = require('fs');
const path = require('path');

// Test: Check that empty type name warnings are handled properly
function testEmptyTypeNameWarnings() {
  const solFile = path.join(__dirname, 'empty_type_name_test/empty_type_name_test.sol');
  
  if (!fs.existsSync(solFile)) {
    console.error('❌ Test file not generated');
    process.exit(1);
  }
  
  const solContent = fs.readFileSync(solFile, 'utf8');
  
  // Check that the package library is generated
  if (!/library\s+Empty_type_name_test\s*{/.test(solContent)) {
    console.error('❌ Package library not generated');
    process.exit(1);
  }
  
  // Check that TestMessage struct is generated
  if (!/struct\s+TestMessage\s*{/.test(solContent)) {
    console.error('❌ TestMessage struct not generated');
    process.exit(1);
  }
  
  // Check that the fields are properly typed (not UnknownType or PlaceholderType)
  if (!/string\s+value;/.test(solContent)) {
    console.error('❌ String field "value" not properly typed');
    process.exit(1);
  }
  
  if (!/int32\s+count;/.test(solContent)) {
    console.error('❌ Int32 field "count" not properly typed');
    process.exit(1);
  }
  
  // Check that UnknownType is NOT generated (should not happen with valid proto)
  if (/struct\s+UnknownType\s*{/.test(solContent)) {
    console.error('❌ UnknownType struct generated - empty type name not handled properly');
    process.exit(1);
  }
  
  // Check that PlaceholderType is NOT generated (should not happen with valid proto)
  if (/struct\s+PlaceholderType\s*{/.test(solContent)) {
    console.error('❌ PlaceholderType struct generated - empty type name not handled properly');
    process.exit(1);
  }
  
  // Check that the codec library is generated
  if (!/library\s+TestMessageCodec\s*{/.test(solContent)) {
    console.error('❌ Codec library not generated');
    process.exit(1);
  }
  
  console.log('✅ Empty type name warnings handled properly');
}

// Run the test
testEmptyTypeNameWarnings(); 