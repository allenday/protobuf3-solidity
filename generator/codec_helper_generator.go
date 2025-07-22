package generator

import (
	"fmt"

	"google.golang.org/protobuf/types/descriptorpb"
)

// CodecHelperGenerator handles generation of codec helper functions
type CodecHelperGenerator struct{}

// NewCodecHelperGenerator creates a new codec helper generator
func NewCodecHelperGenerator() *CodecHelperGenerator {
	return &CodecHelperGenerator{}
}

// GenerateCodecHelpers generates helper functions for codec libraries
func (chg *CodecHelperGenerator) GenerateCodecHelpers(structName string, fields []*descriptorpb.FieldDescriptorProto, fieldNameMap map[int32]string, b *WriteableBuffer) error {
	// Generate check_key function
	chg.generateCheckKeyFunction(structName, fields, b)

	// Generate decode_field function
	chg.generateDecodeFieldFunction(structName, fields, fieldNameMap, b)

	return nil
}

// generateCheckKeyFunction generates the check_key function for wire type validation
func (chg *CodecHelperGenerator) generateCheckKeyFunction(structName string, fields []*descriptorpb.FieldDescriptorProto, b *WriteableBuffer) {
	b.P("function check_key(uint64 field_number, ProtobufLib.WireType wire_type) internal pure returns (bool) {")
	b.Indent()

	// Generate wire type checks for each field
	for _, field := range fields {
		fieldNumber := field.GetNumber()
		fieldType := field.GetType()

		b.P(fmt.Sprintf("if (field_number == %d) {", fieldNumber))
		b.Indent()

		// Check wire type based on field type
		switch fieldType {
		case descriptorpb.FieldDescriptorProto_TYPE_INT32,
			descriptorpb.FieldDescriptorProto_TYPE_INT64,
			descriptorpb.FieldDescriptorProto_TYPE_UINT32,
			descriptorpb.FieldDescriptorProto_TYPE_UINT64,
			descriptorpb.FieldDescriptorProto_TYPE_SINT32,
			descriptorpb.FieldDescriptorProto_TYPE_SINT64,
			descriptorpb.FieldDescriptorProto_TYPE_BOOL,
			descriptorpb.FieldDescriptorProto_TYPE_ENUM:
			b.P("return wire_type == ProtobufLib.WireType.Varint;")
		case descriptorpb.FieldDescriptorProto_TYPE_FIXED64,
			descriptorpb.FieldDescriptorProto_TYPE_SFIXED64,
			descriptorpb.FieldDescriptorProto_TYPE_DOUBLE:
			b.P("return wire_type == ProtobufLib.WireType.Bits64;")
		case descriptorpb.FieldDescriptorProto_TYPE_FIXED32,
			descriptorpb.FieldDescriptorProto_TYPE_SFIXED32,
			descriptorpb.FieldDescriptorProto_TYPE_FLOAT:
			b.P("return wire_type == ProtobufLib.WireType.Bits32;")
		case descriptorpb.FieldDescriptorProto_TYPE_STRING,
			descriptorpb.FieldDescriptorProto_TYPE_BYTES,
			descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
			b.P("return wire_type == ProtobufLib.WireType.LengthDelimited;")
		default:
			b.P("return false; // Unknown field type")
		}

		b.Unindent()
		b.P("}")
	}

	b.P("return false; // Unknown field number")
	b.Unindent()
	b.P("}")
	b.P0()
}

// generateDecodeFieldFunction generates the decode_field function for field decoding
func (chg *CodecHelperGenerator) generateDecodeFieldFunction(structName string, fields []*descriptorpb.FieldDescriptorProto, fieldNameMap map[int32]string, b *WriteableBuffer) {
	b.P(fmt.Sprintf("function decode_field(uint64 pos, bytes memory buf, uint64 len, uint64 field_number, %s memory instance) internal pure returns (bool, uint64) {", structName))
	b.Indent()

	// Generate field decoding for each field
	for _, field := range fields {
		fieldNumber := field.GetNumber()
		fieldName := fieldNameMap[fieldNumber]

		b.P(fmt.Sprintf("if (field_number == %d) {", fieldNumber))
		b.Indent()

		// Generate decoding logic based on field type
		chg.generateFieldDecoding(field, fieldName, structName, b)

		b.Unindent()
		b.P("}")
	}

	b.P("return (false, pos); // Unknown field number")
	b.Unindent()
	b.P("}")
	b.P0()
}

// generateFieldDecoding generates the decoding logic for a specific field
func (chg *CodecHelperGenerator) generateFieldDecoding(field *descriptorpb.FieldDescriptorProto, fieldName string, structName string, b *WriteableBuffer) {
	fieldType := field.GetType()
	isRepeated := isFieldRepeated(field)

	switch fieldType {
	case descriptorpb.FieldDescriptorProto_TYPE_STRING:
		b.P("bool success;")
		b.P("uint64 new_pos;")
		b.P("string memory value;")
		b.P(fmt.Sprintf("(success, new_pos, value) = ProtobufLib.decode_string(pos, buf);"))
		b.P("if (!success) {")
		b.Indent()
		b.P("return (false, pos);")
		b.Unindent()
		b.P("}")
		if isRepeated {
			// For repeated fields, we need to append to the array
			// TODO: Implement proper repeated field handling - for now, just a placeholder
			b.P("// TODO: Implement repeated field appending")
			b.P(fmt.Sprintf("// instance.%s.push(value); // This syntax doesn't exist in Solidity", fieldName))
		} else {
			b.P(fmt.Sprintf("instance.%s = value;", fieldName))
		}
		b.P("pos = new_pos;")
		b.P("return (true, pos);")

	case descriptorpb.FieldDescriptorProto_TYPE_UINT32:
		b.P("bool success;")
		b.P("uint64 new_pos;")
		b.P("uint32 value;")
		b.P(fmt.Sprintf("(success, new_pos, value) = ProtobufLib.decode_uint32(pos, buf);"))
		b.P("if (!success) {")
		b.Indent()
		b.P("return (false, pos);")
		b.Unindent()
		b.P("}")
		b.P(fmt.Sprintf("instance.%s = value;", fieldName))
		b.P("pos = new_pos;")
		b.P("return (true, pos);")

	case descriptorpb.FieldDescriptorProto_TYPE_INT32:
		b.P("bool success;")
		b.P("uint64 new_pos;")
		b.P("int32 value;")
		b.P(fmt.Sprintf("(success, new_pos, value) = ProtobufLib.decode_int32(pos, buf);"))
		b.P("if (!success) {")
		b.Indent()
		b.P("return (false, pos);")
		b.Unindent()
		b.P("}")
		b.P(fmt.Sprintf("instance.%s = value;", fieldName))
		b.P("pos = new_pos;")
		b.P("return (true, pos);")

	case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
		b.P("bool success;")
		b.P("uint64 new_pos;")
		b.P("bool value;")
		b.P(fmt.Sprintf("(success, new_pos, value) = ProtobufLib.decode_bool(pos, buf);"))
		b.P("if (!success) {")
		b.Indent()
		b.P("return (false, pos);")
		b.Unindent()
		b.P("}")
		b.P(fmt.Sprintf("instance.%s = value;", fieldName))
		b.P("pos = new_pos;")
		b.P("return (true, pos);")

	case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
		b.P("bool success;")
		b.P("uint64 new_pos;")
		b.P("uint64 length;")
		b.P(fmt.Sprintf("(success, new_pos, length) = ProtobufLib.decode_bytes(pos, buf);"))
		b.P("if (!success) {")
		b.Indent()
		b.P("return (false, pos);")
		b.Unindent()
		b.P("}")
		b.P("bytes memory value = new bytes(length);")
		b.P("for (uint64 i = 0; i < length; i++) {")
		b.Indent()
		b.P("value[i] = buf[new_pos + i];")
		b.Unindent()
		b.P("}")
		if isRepeated {
			// For repeated fields, we need to append to the array
			// TODO: Implement proper repeated field handling - for now, just a placeholder
			b.P("// TODO: Implement repeated field appending")
			b.P(fmt.Sprintf("// instance.%s.push(value); // This syntax doesn't exist in Solidity", fieldName))
		} else {
			b.P(fmt.Sprintf("instance.%s = value;", fieldName))
		}
		b.P("pos = new_pos + length;")
		b.P("return (true, pos);")

	case descriptorpb.FieldDescriptorProto_TYPE_FIXED32:
		b.P("bool success;")
		b.P("uint64 new_pos;")
		b.P("uint32 value;")
		b.P(fmt.Sprintf("(success, new_pos, value) = ProtobufLib.decode_fixed32(pos, buf);"))
		b.P("if (!success) {")
		b.Indent()
		b.P("return (false, pos);")
		b.Unindent()
		b.P("}")
		b.P(fmt.Sprintf("instance.%s = value;", fieldName))
		b.P("pos = new_pos;")
		b.P("return (true, pos);")

	default:
		// For other types, use a generic approach
		b.P("// TODO: Implement decoding for field type")
		b.P("return (false, pos);")
	}
}
