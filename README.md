# protobuf3-solidity

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/lazyledger/protobuf3-solidity)
[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/lazyledger/protobuf3-solidity)](https://github.com/lazyledger/protobuf3-solidity/releases)
[![Go and Node.js Tests](https://github.com/lazyledger/protobuf3-solidity/workflows/Go%20and%20Node.js%20Tests/badge.svg)](https://github.com/lazyledger/protobuf3-solidity/actions?query=workflow%3A%22Go+and+Node.js+Tests%22)
[![GitHub](https://img.shields.io/github/license/lazyledger/protobuf3-solidity)](https://github.com/lazyledger/protobuf3-solidity/blob/master/LICENSE)

A [protobuf3](https://developers.google.com/protocol-buffers) code generator plugin for [Solidity](https://github.com/ethereum/solidity). Decode and encode protobuf messages in your Solidity contract! Leverages the [protobuf3-solidity-lib](https://github.com/lazyledger/protobuf3-solidity-lib) codec library.

Serialization rules are stricter than default protobuf3 rules, and are specified in [ADR-027](https://github.com/cosmos/cosmos-sdk/blob/master/docs/architecture/adr-027-deterministic-protobuf-serialization.md). The resulting serialization is bijective (one-to-one), rather than the usual non-deterministic and malleable serialization used in most protobuf parsers. This makes it suitable for canonical serialization in blockchain applications.

## Usage

Use as a `protoc` plugin:
```sh
protoc \
--plugin protoc-gen-sol \
--sol_out [license=<license string>,compile=<link,inline>,generate=<all,decoder,encoder>:]<output directory> \
<proto files>
```

Examples:
```sh
# Output foo.proto.sol in current directory
protoc --plugin protoc-gen-sol --sol_out . foo.proto

# Generate Solidity file with Apache-2.0 license identifier
protoc --plugin protoc-gen-sol --sol_out license=Apache-2.0:. foo.proto
```

### Parameters

- `license`: default `CC0`
  - any string is accepted, and the generated license comment will use the string as-is
- `compile`: default `inline`
  - `inline`: the generated library's functions will be inlined (`JUMP`)
  - `link`: the generated library's functions will be linked (`DELEGATECALL`)
- `generate`: default `decoder`
  - `all`: both decoder and encoder will be generated
  - `decoder`: only decoder will be generated
  - `encoder`: only encoder will be generated (experimental!)

### Feature support

The below protobuf file shows all supported features of this plugin.
```protobuf
syntax = "proto3";

// import is supported but not shown here

enum OtherEnum {
  UNSPECIFIED = 0;
  ONE = 1;
  TWO = 2;
};

message OtherMessage {
  uint64 other_field = 1;
}

message Message {
  int32 optional_int32 = 1;
  int64 optional_int64 = 2;
  uint32 optional_uint32 = 3;
  uint64 optional_uint64 = 4;
  sint32 optional_sint32 = 5;
  sint64 optional_sint64 = 6;
  fixed32 optional_fixed32 = 7;
  fixed64 optional_fixed64 = 8;
  sfixed32 optional_sfixed32 = 9;
  sfixed64 optional_sfixed64 = 10;
  bool optional_bool = 11;
  string optional_string = 12;
  bytes optional_bytes = 13;
  OtherEnum optional_enum = 14;
  OtherMessage optional_message = 15;
  float optional_float = 16;
  double optional_double = 17;

  repeated int32 repeated_int32 = 18 [packed = true];
  repeated int64 repeated_int64 = 19 [packed = true];
  repeated uint32 repeated_uint32 = 20 [packed = true];
  repeated uint64 repeated_uint64 = 21 [packed = true];
  repeated sint32 repeated_sint32 = 22 [packed = true];
  repeated sint64 repeated_sint64 = 23 [packed = true];
  repeated fixed32 repeated_fixed32 = 24 [packed = true];
  repeated fixed64 repeated_fixed64 = 25 [packed = true];
  repeated sfixed32 repeated_sfixed32 = 26 [packed = true];
  repeated sfixed64 repeated_sfixed64 = 27 [packed = true];
  repeated bool repeated_bool = 28 [packed = true];
  repeated OtherEnum repeated_enum = 29 [packed = true];
  repeated OtherMessage repeated_message = 30;
  
  // Repeated strings are supported
  repeated string repeated_strings = 31;
  
  // Map fields are supported
  map<string, uint64> string_to_uint64_map = 32;
  map<uint32, string> uint32_to_string_map = 33;
  
  // Oneof fields are supported
  oneof one_of {
    uint64 field1 = 34;
    string field2 = 35;
  }
}

// gRPC services are supported
service ExampleService {
  rpc GetMessage(GetMessageRequest) returns (GetMessageResponse);
  rpc StreamMessages(StreamMessagesRequest) returns (stream GetMessageResponse);
}

message GetMessageRequest {
  string id = 1;
}

message GetMessageResponse {
  Message message = 1;
}

message StreamMessagesRequest {
  string filter = 1;
}
```

**Rules to keep in mind:**
1. Enum values must start at `0` and increment by `1` (unless `strict_enum_validation=false`).
1. Field numbers must start at `1` and increment by `1` (unless `strict_field_numbers=false`).
1. Repeated numeric types must explicitly specify `[packed = true]`.

**Supported features**:
1. ✅ **All primitive types** - int32, int64, uint32, uint64, sint32, sint64, fixed32, fixed64, sfixed32, sfixed64, bool, string, bytes, float, double
2. ✅ **Enums** - Top-level enum definitions with proper validation
3. ✅ **Messages** - Top-level message definitions with all field types
4. ✅ **Imports** - Cross-file imports with proper type resolution
5. ✅ **Package support** - Full package namespace support with automatic library generation
6. ✅ **Map fields** - Support for protobuf map fields with automatic wrapper message generation
7. ✅ **Repeated fields** - Support for repeated fields including strings (with automatic wrapper message generation)
8. ✅ **Oneof fields** - Support for protobuf oneof fields (mutual exclusivity at runtime)
9. ✅ **gRPC services** - Generation of Solidity interfaces for gRPC service definitions
10. ✅ **Float/Double types** - Automatic conversion to fixed-point integers with proper IEEE 754 scaling (float→int32 with 1e6 precision, double→int64 with 1e15 precision)
11. ✅ **Cross-package type resolution** - Proper handling of package-qualified type names
12. ✅ **Configurable validation** - Relaxed validation options for Google API compatibility
13. ✅ **Scoped helper messages** - Helper messages are properly scoped to package namespaces

**Currently unsupported features**:
1. ❌ **Nested enum definitions** - Enums must be defined at the top level, not inside messages
2. ❌ **Nested message definitions** - Messages must be defined at the top level, not inside other messages
3. ❌ **Repeated bytes fields** - Repeated bytes fields are not supported
4. ❌ **Repeated message fields with packed=true** - Packed encoding is only supported for numeric types
5. ❌ **Repeated numeric fields without packed=true** - All repeated numeric fields must be packed
6. ❌ **Empty enums** - Enums must contain at least one value
7. ❌ **Proto2 syntax** - Only proto3 is supported
8. ❌ **Group fields** - Legacy protobuf feature not supported in proto3
9. ❌ **Custom options** - Protobuf custom options are not processed
10. ❌ **Extensions** - Protobuf extensions are not supported

## Building from source

Requires [Go](https://golang.org/) `>= 1.14`.

Build:
```sh
make
```

Test (requires a [`protoc`](https://github.com/protocolbuffers/protobuf) binary in `PATH`):
```sh
make test
```
