const fs = require('fs');
const path = require('path');
const assert = require('assert');
const { execSync } = require('child_process');

describe('Import Path Tests', () => {
    const generatedFile = path.join(__dirname, 'import_test', 'import_paths.sol');

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

    it('should handle import path parameter correctly', () => {
        // Test with scoped package path
        const scopedPath = '@lazyledger/protobuf3-solidity-lib/contracts/ProtobufLib.sol';
        const nodeModulesPath = 'node_modules/@lazyledger/protobuf3-solidity-lib/contracts/ProtobufLib.sol';
        
        // Run protoc with different import paths
        const result1 = execSync(`protoc --plugin bin/protoc-gen-sol --sol_out "protobuf_lib_import=${scopedPath}:test/pass/import_test" -I test/pass/import_test test/pass/import_test/import_paths.proto`);
        const result2 = execSync(`protoc --plugin bin/protoc-gen-sol --sol_out "protobuf_lib_import=${nodeModulesPath}:test/pass/import_test" -I test/pass/import_test test/pass/import_test/import_paths.proto`);

        // Read the generated files
        const content1 = fs.readFileSync(generatedFile, 'utf8');
        const content2 = fs.readFileSync(generatedFile, 'utf8');

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