package generator

// GoogleProtobufTypes provides shared generation functions for Google protobuf type definitions
type GoogleProtobufTypes struct{}

// NewGoogleProtobufTypes creates a new Google protobuf types helper
func NewGoogleProtobufTypes() *GoogleProtobufTypes {
	return &GoogleProtobufTypes{}
}

// GenerateAllTypes generates all common Google protobuf type definitions
func (gpt *GoogleProtobufTypes) GenerateAllTypes(b *WriteableBuffer) {
	gpt.generateStructDefinition(b)
	gpt.generateTimestampDefinition(b)
	gpt.generateEmptyDefinition(b)
}

// generateStructDefinition generates the Struct type definition
func (gpt *GoogleProtobufTypes) generateStructDefinition(b *WriteableBuffer) {
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
func (gpt *GoogleProtobufTypes) generateTimestampDefinition(b *WriteableBuffer) {
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
func (gpt *GoogleProtobufTypes) generateEmptyDefinition(b *WriteableBuffer) {
	b.P("// google.protobuf.Empty - represents an empty message")
	b.P("// Note: Empty structs are not allowed in Solidity, using placeholder")
	b.P("struct Empty {")
	b.Indent()
	b.P("bool _placeholder; // Placeholder field to avoid empty struct compilation error")
	b.Unindent()
	b.P("}")
	b.P0()
}