const fs = require('fs');
const path = require('path');

// Test: Check that Google protobuf types are properly generated
function testGoogleProtobufTypes() {
  const solFile = path.join(__dirname, 'google_protobuf_test/google_protobuf_test.sol');
  
  if (!fs.existsSync(solFile)) {
    console.error('❌ Test file not generated');
    process.exit(1);
  }
  
  const solContent = fs.readFileSync(solFile, 'utf8');
  
  // Check for proper Google protobuf library definition
  if (!/library\s+Google_Protobuf\s*{/.test(solContent)) {
    console.error('❌ Missing Google_Protobuf library definition');
    process.exit(1);
  }
  
  // Check for struct definitions
  if (!/struct\s+Struct\s*{/.test(solContent)) {
    console.error('❌ Missing Google_Protobuf.Struct definition');
    process.exit(1);
  }
  
  if (!/struct\s+Timestamp\s*{/.test(solContent)) {
    console.error('❌ Missing Google_Protobuf.Timestamp definition');
    process.exit(1);
  }
  
  if (!/struct\s+Empty\s*{/.test(solContent)) {
    console.error('❌ Missing Google_Protobuf.Empty definition');
    process.exit(1);
  }
  
  // Check that the types are being used correctly
  if (!/Google_Protobuf\.Struct/.test(solContent)) {
    console.error('❌ Google_Protobuf.Struct not being used');
    process.exit(1);
  }
  
  if (!/Google_Protobuf\.Timestamp/.test(solContent)) {
    console.error('❌ Google_Protobuf.Timestamp not being used');
    process.exit(1);
  }
  
  if (!/Google_Protobuf\.Empty/.test(solContent)) {
    console.error('❌ Google_Protobuf.Empty not being used');
    process.exit(1);
  }
  
  console.log('✅ Google protobuf types properly generated');
}

// Run the test
testGoogleProtobufTypes(); 