/**
 * Regression test for qualified struct names in generated codec functions
 * 
 * This test ensures that all codec function signatures use fully-qualified
 * struct names (e.g., "LibraryName.StructName memory") instead of 
 * unqualified names (e.g., "StructName memory").
 * 
 * This prevents the critical Solidity compilation error:
 * "Error (7920): Identifier not found or not unique."
 */

const fs = require('fs');
const path = require('path');

function testQualifiedStructNames() {
    console.log('Testing qualified struct names in codec functions...');
    
    const solFilePath = path.join(__dirname, 'qualified', 'test', 'qualified_struct_names.sol');
    
    if (!fs.existsSync(solFilePath)) {
        console.error('ERROR: Generated Solidity file not found:', solFilePath);
        process.exit(1);
    }
    
    const solContent = fs.readFileSync(solFilePath, 'utf8');
    
    // Test cases: These patterns MUST be present (qualified names)
    const requiredPatterns = [
        // TestMessage codec functions must use qualified names
        /function decode_field\(.*Qualified_Test\.TestMessage memory instance\)/,
        /function decode\(.*returns \(bool, uint64, Qualified_Test\.TestMessage memory\)/,
        /Qualified_Test\.TestMessage memory instance;/,
        
        // NestedMessage codec functions must use qualified names  
        /function decode_field\(.*Qualified_Test\.NestedMessage memory instance\)/,
        /function decode\(.*returns \(bool, uint64, Qualified_Test\.NestedMessage memory\)/,
        /Qualified_Test\.NestedMessage memory instance;/,
    ];
    
    // Test cases: These patterns MUST NOT be present (unqualified names in codec functions)
    const forbiddenPatterns = [
        // Unqualified struct names in codec function signatures are forbidden
        /function decode_field\(.*\s+TestMessage memory instance\)/, // Missing Qualified_Test.
        /function decode\(.*returns \(bool, uint64, TestMessage memory\)/, // Missing Qualified_Test.
        /function decode_field\(.*\s+NestedMessage memory instance\)/, // Missing Qualified_Test.
        /function decode\(.*returns \(bool, uint64, NestedMessage memory\)/, // Missing Qualified_Test.
    ];
    
    let success = true;
    
    // Check that all required patterns are present
    console.log('\nChecking for required qualified struct names...');
    requiredPatterns.forEach((pattern, index) => {
        if (pattern.test(solContent)) {
            console.log(`✓ Required pattern ${index + 1}: Found qualified struct name`);
        } else {
            console.error(`✗ Required pattern ${index + 1}: Missing qualified struct name`);
            console.error(`  Pattern: ${pattern.toString()}`);
            success = false;
        }
    });
    
    // Check that forbidden patterns are NOT present
    console.log('\nChecking for forbidden unqualified struct names...');
    forbiddenPatterns.forEach((pattern, index) => {
        const matches = solContent.match(pattern);
        if (matches) {
            console.error(`✗ Forbidden pattern ${index + 1}: Found unqualified struct name in codec function`);
            console.error(`  Pattern: ${pattern.toString()}`);
            console.error(`  Match: ${matches[0]}`);
            success = false;
        } else {
            console.log(`✓ Forbidden pattern ${index + 1}: No unqualified struct names found in codec functions`);
        }
    });
    
    // Additional validation: Check that the library name follows the correct pattern
    console.log('\nChecking library name conversion...');
    if (/library Qualified_Test \{/.test(solContent)) {
        console.log('✓ Library name correctly converted: qualified.test → Qualified_Test');
    } else {
        console.error('✗ Library name conversion failed');
        success = false;
    }
    
    // Regression test for data location specifiers (fix for missing 'memory' bug)
    console.log('\nChecking data location specifiers for local variables...');
    const dataLocationPatterns = [
        // Must have 'memory' specifier for string local variables
        /string memory value;/,
    ];
    
    const forbiddenDataLocationPatterns = [
        // Must NOT have unqualified string declarations in codec functions
        /function\s+decode_field\([^)]*\)\s+internal\s+pure\s+returns\s*\([^)]*\)\s*\{[^}]*string\s+value;/s,
    ];
    
    dataLocationPatterns.forEach((pattern, index) => {
        if (pattern.test(solContent)) {
            console.log(`✓ Data location pattern ${index + 1}: Found correct 'memory' specifier`);
        } else {
            console.error(`✗ Data location pattern ${index + 1}: Missing 'memory' specifier`);
            console.error(`  Pattern: ${pattern.toString()}`);
            success = false;
        }
    });
    
    forbiddenDataLocationPatterns.forEach((pattern, index) => {
        const matches = solContent.match(pattern);
        if (matches) {
            console.error(`✗ Forbidden data location pattern ${index + 1}: Found local variable without data location specifier`);
            console.error(`  Pattern: ${pattern.toString()}`);
            console.error(`  Match: ${matches[0].substring(0, 100)}...`);
            success = false;
        } else {
            console.log(`✓ Forbidden data location pattern ${index + 1}: No local variables without data location specifiers found`);
        }
    });
    
    if (success) {
        console.log('\n✅ All qualified struct name tests PASSED');
        console.log('✅ Regression test confirms the fix is working correctly');
        return true;
    } else {
        console.error('\n❌ Qualified struct name tests FAILED');
        console.error('❌ The unqualified struct names bug may have regressed');
        return false;
    }
}

// Run the test
if (require.main === module) {
    const success = testQualifiedStructNames();
    process.exit(success ? 0 : 1);
}

module.exports = { testQualifiedStructNames };