const fs = require('fs');
const path = require('path');

// Test: Check that codec libraries are complete and functional
function testIncompleteCodecLibraries() {
  const solFile = path.join(__dirname, 'incomplete_codec_test/incomplete_codec_test.sol');
  
  if (!fs.existsSync(solFile)) {
    console.error('❌ Test file not generated');
    process.exit(1);
  }
  
  const solContent = fs.readFileSync(solFile, 'utf8');
  
  // Check that codec library exists
  if (!/library\s+TestMessageCodec\s*{/.test(solContent)) {
    console.error('❌ Missing TestMessageCodec library');
    process.exit(1);
  }
  
  // Check that helper functions are properly defined
  if (!/function\s+check_key\s*\(/.test(solContent)) {
    console.error('❌ Missing check_key function definition');
    process.exit(1);
  }
  
  if (!/function\s+decode_field\s*\(/.test(solContent)) {
    console.error('❌ Missing decode_field function definition');
    process.exit(1);
  }
  
  // Check that decode function exists and is complete
  if (!/function\s+decode\s*\(/.test(solContent)) {
    console.error('❌ Missing decode function in codec library');
    process.exit(1);
  }
  
  // Check that the decode function has proper implementation
  if (!/ProtobufLib\.decode_key/.test(solContent)) {
    console.error('❌ Missing proper decode_key call in decode function');
    process.exit(1);
  }
  
  // Check that the helper functions are being called correctly
  if (!/check_key\(field_number, wire_type\)/.test(solContent)) {
    console.error('❌ Missing proper check_key call in decode function');
    process.exit(1);
  }
  
  if (!/decode_field\(pos, buf, len, field_number, instance\)/.test(solContent)) {
    console.error('❌ Missing proper decode_field call in decode function');
    process.exit(1);
  }
  
  console.log('✅ Codec libraries are complete and functional');
}

// Run the test
testIncompleteCodecLibraries(); 