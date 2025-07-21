/**
 * Regression test for type conversion and wire type bugs in generated codec functions
 * 
 * This test ensures that:
 * 1. Repeated fields don't cause type conversion errors (use TODO instead of incorrect assignment)
 * 2. Fixed32 fields are properly implemented (not TODO)
 * 3. Single fields work correctly
 * 4. Data location specifiers are correct
 */

const fs = require('fs');
const path = require('path');

function testTypeConversionFixes() {
    console.log('Testing type conversion and wire type fixes...');
    
    const solFilePath = path.join(__dirname, 'type_conversion_test', 'type_conversion_test.sol');
    
    if (!fs.existsSync(solFilePath)) {
        console.error('ERROR: Generated Solidity file not found:', solFilePath);
        process.exit(1);
    }
    
    const solContent = fs.readFileSync(solFilePath, 'utf8');
    
    // Test cases: Required patterns (should be present)
    const requiredPatterns = [
        // Data location specifiers should be present
        /string memory value;/,
        /bytes memory value;/,
        
        // Fixed32 field should be properly implemented (not TODO)
        /instance\.fixed_value = value;/,
        /decode_fixed32\(pos, buf\);/,
        
        // Single field assignments should work
        /instance\.text_data = value;/,
        /instance\.binary_data = value;/,
        
        // Wire type checking should include Fixed32
        /wire_type == ProtobufLib\.WireType\.Fixed32;/,
        
        // Library name should be correct
        /library Type_conversion_test \{/,
        
        // PlaceholderType should be defined
        /struct PlaceholderType \{/,
        /bytes corrupted_data;/,
    ];
    
    // Test cases: Forbidden patterns (should NOT be present)
    const forbiddenPatterns = [
        // Should NOT have incorrect type assignments for repeated fields
        /instance\.string_array = value;/,
        /instance\.bytes_array = value;/,
        
        // Should NOT have data location specifiers missing 
        /string\s+value;/,  // Should have 'memory'
    ];
    
    // Test cases: TODO patterns (acceptable temporary implementations)  
    const expectedTodoPatterns = [
        // Repeated field handling should have TODO comments
        /TODO: Implement repeated field appending/,
        /instance\.string_array\.push\(value\); \/\/ This syntax doesn't exist in Solidity/,
        /instance\.bytes_array\.push\(value\); \/\/ This syntax doesn't exist in Solidity/,
    ];
    
    let success = true;
    
    // Check that all required patterns are present
    console.log('\nChecking for required patterns...');
    requiredPatterns.forEach((pattern, index) => {
        if (pattern.test(solContent)) {
            console.log(`✓ Required pattern ${index + 1}: Found`);
        } else {
            console.error(`✗ Required pattern ${index + 1}: Missing`);
            console.error(`  Pattern: ${pattern.toString()}`);
            success = false;
        }
    });
    
    // Check that forbidden patterns are NOT present
    console.log('\nChecking for forbidden patterns...');
    forbiddenPatterns.forEach((pattern, index) => {
        const matches = solContent.match(pattern);
        if (matches) {
            console.error(`✗ Forbidden pattern ${index + 1}: Found (should not exist)`);
            console.error(`  Pattern: ${pattern.toString()}`);
            console.error(`  Match: ${matches[0]}`);
            success = false;
        } else {
            console.log(`✓ Forbidden pattern ${index + 1}: Not found (correct)`);
        }
    });
    
    // Check that expected TODO patterns are present
    console.log('\nChecking for expected TODO patterns...');
    expectedTodoPatterns.forEach((pattern, index) => {
        if (pattern.test(solContent)) {
            console.log(`✓ Expected TODO pattern ${index + 1}: Found (acceptable temporary implementation)`);
        } else {
            console.error(`✗ Expected TODO pattern ${index + 1}: Missing (repeated fields should have TODO comments)`);
            console.error(`  Pattern: ${pattern.toString()}`);
            success = false;
        }
    });
    
    if (success) {
        console.log('\n✅ All type conversion fix tests PASSED');
        console.log('✅ Regression tests confirm the fixes are working correctly');
        return true;
    } else {
        console.error('\n❌ Type conversion fix tests FAILED');
        console.error('❌ Some bugs may have regressed or new issues introduced');
        return false;
    }
}

// Run the test
if (require.main === module) {
    const success = testTypeConversionFixes();
    process.exit(success ? 0 : 1);
}

module.exports = { testTypeConversionFixes };