const fs = require('fs');
const path = require('path');
const assert = require('assert');
const { execSync } = require('child_process');

describe('Import Path Tests', () => {
    const generatedFile = path.join(__dirname, 'import_test', 'import_paths.sol');
    const packageLevelFile1 = path.join(__dirname, 'test/v1.sol');
    const packageLevelFile2 = path.join(__dirname, 'test/v2.sol');

    it('should generate correct import paths', () => {
        // Read the generated Solidity file
        const content = fs.readFileSync(generatedFile, 'utf8');

        // Split into lines for easier testing
        const lines = content.split('\n');

        // Find all import statements
        const imports = lines.filter(line => line.trim().startsWith('import'));

        // Verify we have exactly 3 unique imports (ProtobufLib.sol, helper.sol, nested.sol)
        assert.strictEqual(imports.length, 3, 'Should have exactly 3 unique imports');

        // Verify ProtobufLib is imported correctly (local path)
        assert.ok(
            imports.some(line => line.includes('import "ProtobufLib.sol";')),
            'ProtobufLib should be imported using local path'
        );

        // Verify no scoped package imports exist
        assert.ok(
            !imports.some(line => line.includes('@')),
            'No imports should contain @ symbol'
        );

        // Verify no node_modules imports exist
        assert.ok(
            !imports.some(line => line.includes('node_modules')),
            'No imports should contain node_modules'
        );

        // Verify helper.proto is imported correctly
        assert.ok(
            imports.some(line => line.includes('import "helper.sol";')),
            'helper.proto should be imported as helper.sol'
        );

        // Verify nested.proto is imported correctly
        assert.ok(
            imports.some(line => line.includes('import "subfolder/nested.sol";')),
            'nested.proto should be imported as subfolder/nested.sol'
        );
    });

    it('should handle package-level imports correctly', () => {
        // Read the package-level Solidity files
        const content1 = fs.readFileSync(packageLevelFile1, 'utf8');
        const content2 = fs.readFileSync(packageLevelFile2, 'utf8');

        // Split into lines for easier testing
        const lines1 = content1.split('\n');
        const lines2 = content2.split('\n');

        // Find all import statements
        const imports1 = lines1.filter(line => line.trim().startsWith('import'));
        const imports2 = lines2.filter(line => line.trim().startsWith('import'));

        // Verify each file has exactly 1 import (ProtobufLib.sol)
        assert.strictEqual(imports1.length, 1, 'Package-level file 1 should have exactly 1 import');
        assert.strictEqual(imports2.length, 1, 'Package-level file 2 should have exactly 1 import');

        // Verify ProtobufLib is imported correctly (local path) in both files
        assert.ok(
            imports1.some(line => line.includes('import "ProtobufLib.sol";')),
            'Package-level file 1 should import ProtobufLib using local path'
        );
        assert.ok(
            imports2.some(line => line.includes('import "ProtobufLib.sol";')),
            'Package-level file 2 should import ProtobufLib using local path'
        );

        // Verify no scoped package imports exist in either file
        assert.ok(
            !imports1.some(line => line.includes('@')),
            'Package-level file 1 should not contain @ symbol'
        );
        assert.ok(
            !imports2.some(line => line.includes('@')),
            'Package-level file 2 should not contain @ symbol'
        );

        // Verify no node_modules imports exist in either file
        assert.ok(
            !imports1.some(line => line.includes('node_modules')),
            'Package-level file 1 should not contain node_modules'
        );
        assert.ok(
            !imports2.some(line => line.includes('node_modules')),
            'Package-level file 2 should not contain node_modules'
        );
    });

    it('should handle import path parameter correctly', () => {
        // Test with scoped package path
        const scopedPath = '@lazyledger/protobuf3-solidity-lib/contracts/ProtobufLib.sol';
        const nodeModulesPath = 'node_modules/@lazyledger/protobuf3-solidity-lib/contracts/ProtobufLib.sol';
        
        // Run protoc with different import paths
        const result1 = execSync(`protoc --plugin bin/protoc-gen-sol --sol_out "protobuf_lib_import=${scopedPath}:test/pass/import_test" -I test/pass/import_test test/pass/import_test/package_level/test1.proto`);
        const result2 = execSync(`protoc --plugin bin/protoc-gen-sol --sol_out "protobuf_lib_import=${nodeModulesPath}:test/pass/import_test" -I test/pass/import_test test/pass/import_test/package_level/test2.proto`);

        // Read the generated files
        const content1 = fs.readFileSync(packageLevelFile1, 'utf8');
        const content2 = fs.readFileSync(packageLevelFile2, 'utf8');

        // Verify both files use local import
        assert.ok(
            content1.includes('import "ProtobufLib.sol";'),
            'Should use local import regardless of scoped package path'
        );
        assert.ok(
            content2.includes('import "ProtobufLib.sol";'),
            'Should use local import regardless of node_modules path'
        );
    });
}); 