const fs = require('fs');
const path = require('path');

// Test: Check that duplicate library declarations are avoided
function testDuplicateLibraryDeclarations() {
  const solFile1 = path.join(__dirname, 'duplicate_library_test1/duplicate_library_test1.sol');
  const solFile2 = path.join(__dirname, 'duplicate_library_test2/duplicate_library_test2.sol');
  
  if (!fs.existsSync(solFile1)) {
    console.error('❌ Test file 1 not generated');
    process.exit(1);
  }
  
  if (!fs.existsSync(solFile2)) {
    console.error('❌ Test file 2 not generated');
    process.exit(1);
  }
  
  const solContent1 = fs.readFileSync(solFile1, 'utf8');
  const solContent2 = fs.readFileSync(solFile2, 'utf8');
  
  // Check that first file has Google protobuf library declaration
  if (!/library\s+Google_Protobuf\s*{/.test(solContent1)) {
    console.error('❌ Google_Protobuf library not found in file 1');
    process.exit(1);
  }
  
  // Check that second file does NOT have Google protobuf library declaration (avoiding duplicates)
  if (/library\s+Google_Protobuf\s*{/.test(solContent2)) {
    console.error('❌ Google_Protobuf library found in file 2 - duplicate declaration not avoided');
    process.exit(1);
  }
  
  // Check that first file has Google protobuf type definitions
  if (!/struct\s+Timestamp\s*{/.test(solContent1)) {
    console.error('❌ Timestamp struct not found in file 1');
    process.exit(1);
  }
  
  if (!/struct\s+Empty\s*{/.test(solContent1)) {
    console.error('❌ Empty struct not found in file 1');
    process.exit(1);
  }
  
  if (!/struct\s+Struct\s*{/.test(solContent1)) {
    console.error('❌ Struct definition not found in file 1');
    process.exit(1);
  }
  
  // Check that second file does NOT have Google protobuf type definitions
  if (/struct\s+Timestamp\s*{/.test(solContent2)) {
    console.error('❌ Timestamp struct found in file 2 - duplicate not avoided');
    process.exit(1);
  }
  
  if (/struct\s+Empty\s*{/.test(solContent2)) {
    console.error('❌ Empty struct found in file 2 - duplicate not avoided');
    process.exit(1);
  }
  
  // Check that both files have their own package libraries
  if (!/library\s+Duplicate_library_test1\s*{/.test(solContent1)) {
    console.error('❌ Package library not found in file 1');
    process.exit(1);
  }
  
  if (!/library\s+Duplicate_library_test2\s*{/.test(solContent2)) {
    console.error('❌ Package library not found in file 2');
    process.exit(1);
  }
  
  // Check that both files have their message structs
  if (!/struct\s+TestMessage1\s*{/.test(solContent1)) {
    console.error('❌ TestMessage1 struct not found in file 1');
    process.exit(1);
  }
  
  if (!/struct\s+TestMessage2\s*{/.test(solContent2)) {
    console.error('❌ TestMessage2 struct not found in file 2');
    process.exit(1);
  }
  
  // Check that both files can still reference Google_Protobuf types
  if (!/Google_Protobuf\.Timestamp/.test(solContent1)) {
    console.error('❌ Google_Protobuf.Timestamp reference not found in file 1');
    process.exit(1);
  }
  
  if (!/Google_Protobuf.Timestamp/.test(solContent2)) {
    console.error('❌ Google_Protobuf.Timestamp reference not found in file 2');
    process.exit(1);
  }
  
  console.log('✅ Duplicate library declarations successfully avoided');
}

// Run the test
testDuplicateLibraryDeclarations(); 