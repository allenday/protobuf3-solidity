# Release Notes - protobuf3-solidity v0.3.0

## Overview
This release represents a major enhancement to the protobuf3-solidity plugin, introducing comprehensive support for advanced protobuf features, improved code generation architecture, and critical bug fixes. The plugin has been completely refactored with a modular architecture for better maintainability and extensibility.

## 🚀 Major New Features

### Enhanced Protobuf Feature Support
- **Float and Double Types**: Full support for `float` and `double` protobuf types
- **Packed Arrays**: Support for packed repeated fields with `[packed = true]` option
- **Maps**: Complete support for protobuf map fields (`map<key, value>`)
- **Oneof Fields**: Support for protobuf oneof fields
- **gRPC Services**: Support for service definitions (structs generated for request/response messages)
- **Nested Types**: Support for nested enum and message definitions
- **Repeated Strings**: Full support for repeated string fields
- **Repeated Bytes**: Enhanced support for repeated bytes fields

### Modular Architecture
- **Field Processor**: New modular field processing system
- **Import Manager**: Dedicated import path handling and resolution
- **Library Generator**: Separate codec library generation
- **Message Generator**: Specialized message struct generation
- **Type Utils**: Centralized type conversion utilities
- **Validation**: Comprehensive validation system
- **Enhanced Features**: Configurable feature flags

### Configuration Options
- **ProtobufLib Import Path**: Configurable import path for ProtobufLib dependency
- **Strict Field Numbers**: Optional enforcement of monotonic field ordering
- **Strict Enum Validation**: Optional enforcement of enum value constraints
- **Empty Packed Arrays**: Configurable handling of empty packed arrays
- **Non-Monotonic Fields**: Support for relaxed field ordering

## 🔧 Bug Fixes

### Critical Fixes
- **Missing Struct Definitions**: Fixed bug where codec libraries were generated without corresponding struct definitions
- **Relative Import Paths**: Fixed incorrect relative import path calculation for cross-package imports
- **Assembly Syntax**: Resolved critical assembly syntax errors in generated code
- **Import Path Handling**: Fixed HH1006 errors and improved import path consistency
- **Out-of-Bounds Errors**: Fixed array indexing issues in encoding functions

### Import System Improvements
- **Cross-Package Imports**: Proper handling of imports between different protobuf packages
- **Local vs Package Imports**: Configurable import path handling for different deployment scenarios
- **Duplicate Import Prevention**: Avoid duplicate ProtobufLib imports
- **Directory Structure**: Preserved directory structure in generated import paths

### Validation Enhancements
- **Field Number Validation**: Strict validation of field number ordering
- **Enum Validation**: Strict validation of enum value constraints
- **Empty Message Support**: Proper handling of empty message definitions
- **Repeated Field Validation**: Enhanced validation for repeated fields

## 📁 Test Suite Enhancements

### New Test Categories
- **Cross-Package Imports**: Comprehensive testing of import path resolution
- **Empty Message Validation**: Testing of empty message handling
- **Import Path Handling**: Testing of various import scenarios
- **Package-Level Imports**: Testing of package-level import resolution
- **Scoped Package Imports**: Testing of scoped package import handling

### Moved Tests (Fail → Pass)
The following test categories now pass, indicating new feature support:
- `double` - Float/double type support
- `empty_message` - Empty message handling
- `float` - Float type support
- `map` - Map field support
- `nested_enum_definition` - Nested enum support
- `nested_message_definition` - Nested message support
- `oneof` - Oneof field support
- `repeated_bytes` - Repeated bytes support
- `repeated_string` - Repeated string support

## 🏗️ Architecture Changes

### Code Organization
- **Modular Components**: Split monolithic generator into specialized components
- **Field Processing**: Dedicated field processor for type conversion
- **Import Management**: Centralized import path handling
- **Library Generation**: Separate codec library generation
- **Message Generation**: Specialized message struct generation

### File Structure
```
generator/
├── enhanced_features.go    # Configurable feature flags
├── field_generator.go      # Field-specific generation logic
├── field_processor.go      # Field type processing
├── file_header.go          # File header generation
├── file_naming.go          # File naming utilities
├── generator.go            # Main generator orchestration
├── import_manager.go       # Import path management
├── library_generator.go    # Codec library generation
├── message_generator.go    # Message struct generation
├── type_utils.go           # Type conversion utilities
├── validation.go           # Validation logic
└── writeable_buffer.go     # Buffer management
```

## 📋 Usage Examples

### Basic Usage
```bash
# Generate Solidity file
protoc --plugin protoc-gen-sol --sol_out . foo.proto

# Generate with Apache-2.0 license
protoc --plugin protoc-gen-sol --sol_out license=Apache-2.0:. foo.proto
```

### Advanced Configuration
```bash
# Use local ProtobufLib import (for local development)
protoc --plugin protoc-gen-sol --sol_out protobuf_lib_import=ProtobufLib.sol:. foo.proto

# Use package ProtobufLib import (for npm packages - default)
protoc --plugin protoc-gen-sol --sol_out protobuf_lib_import=@protobuf3-solidity-lib/contracts/ProtobufLib.sol:. foo.proto

# Relaxed validation for compatibility
protoc --plugin protoc-gen-sol --sol_out strict_field_numbers=false,allow_empty_packed_arrays=true:. foo.proto
```

## 🔄 Migration Guide

### From v0.2.x
- **Import Paths**: ProtobufLib import path is now configurable, defaults to package path
- **Validation**: Stricter validation by default, use configuration options to relax
- **Generated Code**: Improved code structure with better error handling
- **Cross-Package Imports**: Now properly supported with correct relative paths

### Breaking Changes
- **Default Import Path**: ProtobufLib now defaults to package import path instead of local
- **Validation**: Stricter field number and enum validation by default
- **Generated Structure**: Codec libraries are now generated at top level for Solidity compliance

## 📊 Statistics
- **60 files changed**
- **11,274 insertions, 7,651 deletions**
- **9 test categories moved from fail to pass**
- **15+ new configuration options**
- **8 new generator modules**

## 🎯 Compatibility
- **Protobuf**: Full proto3 support with enhanced features
- **Solidity**: Compatible with Solidity ^0.8.19
- **Hardhat**: Fixed HH1006 import errors
- **Truffle**: Improved compatibility with Truffle projects

## 🔮 Future Roadmap
- Enhanced error reporting and diagnostics
- Support for proto2 features
- Performance optimizations
- Additional configuration options
- Extended test coverage

---

**Release Date**: July 2024  
**Previous Version**: v0.2.x  
**Commit Range**: a2ec9c62c3c9c59c6f334b446d5b79beb9ebcdca → dc9d0ce 