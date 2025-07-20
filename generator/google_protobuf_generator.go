package generator

import (
	"strings"

	"google.golang.org/protobuf/types/descriptorpb"
)

// GoogleProtobufGenerator handles generation of Google protobuf type definitions
type GoogleProtobufGenerator struct{}

// NewGoogleProtobufGenerator creates a new Google protobuf generator
func NewGoogleProtobufGenerator() *GoogleProtobufGenerator {
	return &GoogleProtobufGenerator{}
}

// GenerateGoogleProtobufTypes generates Solidity definitions for Google protobuf types
func (gpg *GoogleProtobufGenerator) GenerateGoogleProtobufTypes(protoFile *descriptorpb.FileDescriptorProto, b *WriteableBuffer) error {
	// Check if this file uses Google protobuf types
	usesGoogleTypes := false
	for _, dependency := range protoFile.GetDependency() {
		if strings.HasPrefix(dependency, "google/protobuf/") {
			usesGoogleTypes = true
			break
		}
	}

	if !usesGoogleTypes {
		return nil
	}

	// Generate Google protobuf library
	b.P("// Google protobuf type definitions")
	b.P("library Google_Protobuf {")
	b.Indent()

	// Generate struct definitions for commonly used Google protobuf types
	gpg.generateStructDefinition(b)
	gpg.generateTimestampDefinition(b)
	gpg.generateEmptyDefinition(b)

	b.Unindent()
	b.P("}")
	b.P0()

	return nil
}

// generateStructDefinition generates the Struct type definition
func (gpg *GoogleProtobufGenerator) generateStructDefinition(b *WriteableBuffer) {
	b.P("// google.protobuf.Struct - represents a structured data value")
	b.P("struct Struct {")
	b.Indent()
	b.P("// Fields map - in practice this would need proper implementation")
	b.P("// For now, this is a placeholder that allows compilation")
	b.P("// In a real implementation, this would be a map<string, Value>")
	b.P("bytes data; // Placeholder for structured data")
	b.Unindent()
	b.P("}")
	b.P0()
}

// generateTimestampDefinition generates the Timestamp type definition
func (gpg *GoogleProtobufGenerator) generateTimestampDefinition(b *WriteableBuffer) {
	b.P("// google.protobuf.Timestamp - represents a point in time")
	b.P("struct Timestamp {")
	b.Indent()
	b.P("int64 _seconds; // Seconds since Unix epoch")
	b.P("int32 nanos;   // Nanoseconds within the second")
	b.Unindent()
	b.P("}")
	b.P0()
}

// generateEmptyDefinition generates the Empty type definition
func (gpg *GoogleProtobufGenerator) generateEmptyDefinition(b *WriteableBuffer) {
	b.P("// google.protobuf.Empty - represents an empty message")
	b.P("struct Empty {")
	b.Indent()
	b.P("// Empty struct - no fields")
	b.Unindent()
	b.P("}")
	b.P0()
}
