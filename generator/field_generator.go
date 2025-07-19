package generator

import (
	"errors"
	"fmt"
	"strings"

	"google.golang.org/protobuf/types/descriptorpb"
)

// generateMessageDecoder generates the decoder functions for a message
func (g *Generator) generateMessageDecoder(structName string, fields []*descriptorpb.FieldDescriptorProto, fieldNameMap map[int32]string, b *WriteableBuffer) error {
	// Top-level decoder function
	b.P(fmt.Sprintf("function decode(uint64 initial_pos, bytes memory buf, uint64 len) internal pure returns (bool, uint64, %s memory) {", structName))
	b.Indent()

	b.P("// Message instance")
	b.P(fmt.Sprintf("%s memory instance;", structName))
	b.P("// Previous field number")
	b.P("uint64 previous_field_number = 0;")
	b.P("// Current position in the buffer")
	b.P("uint64 pos = initial_pos;")
	b.P("")

	b.P("// Sanity checks")
	b.P("if (pos + len < pos) {")
	b.Indent()
	b.P("return (false, pos, instance);")
	b.Unindent()
	b.P("}")
	b.P("")

	b.P("while (pos - initial_pos < len) {")
	b.Indent()
	b.P("// Decode the key (field number and wire type)")
	b.P("bool success;")
	b.P("uint64 field_number;")
	b.P("ProtobufLib.WireType wire_type;")
	b.P("(success, pos, field_number, wire_type) = ProtobufLib.decode_key(pos, buf);")
	b.P("if (!success) {")
	b.Indent()
	b.P("return (false, pos, instance);")
	b.Unindent()
	b.P("}")
	b.P("")

	b.P("// Check that the field number is within bounds")
	b.P(fmt.Sprintf("if (field_number > %d) {", len(fields)))
	b.Indent()
	b.P("return (false, pos, instance);")
	b.Unindent()
	b.P("}")
	b.P("")

	b.P("// Check that the field number is monotonically increasing")
	if !g.allowNonMonotonicFields {
		b.P("if (field_number <= previous_field_number) {")
		b.Indent()
		b.P("return (false, pos, instance);")
		b.Unindent()
		b.P("}")
	}
	b.P("")

	b.P("// Check that the wire type is correct")
	b.P("success = check_key(field_number, wire_type);")
	b.P("if (!success) {")
	b.Indent()
	b.P("return (false, pos, instance);")
	b.Unindent()
	b.P("}")
	b.P("")

	b.P("// Actually decode the field")
	b.P("(success, pos) = decode_field(pos, buf, len, field_number, instance);")
	b.P("if (!success) {")
	b.Indent()
	b.P("return (false, pos, instance);")
	b.Unindent()
	b.P("}")
	b.P("")

	b.P("previous_field_number = field_number;")
	b.Unindent()
	b.P("}")
	b.P("")

	b.P("return (true, pos, instance);")
	b.Unindent()
	b.P("}")
	b.P("")

	return nil
}

// generateMessageEncoder generates the encoder functions for a message
func (g *Generator) generateMessageEncoder(structName string, fields []*descriptorpb.FieldDescriptorProto, fieldNameMap map[int32]string, b *WriteableBuffer) error {
	// Top-level encoder function
	b.P(fmt.Sprintf("function encode(uint64 pos, bytes memory buf, %s memory instance) internal pure returns (uint64) {", structName))
	b.Indent()

	// Encode each field
	for _, field := range fields {
		fieldNumber := field.GetNumber()
		b.P(fmt.Sprintf("pos = encode_%d(pos, buf, instance);", fieldNumber))
	}

	b.P("return pos;")
	b.Unindent()
	b.P("}")
	b.P("")

	// Individual field encoders
	for _, field := range fields {
		fieldName := fieldNameMap[field.GetNumber()]
		fieldDescriptorType := field.GetType()
		fieldNumber := field.GetNumber()

		b.P(fmt.Sprintf("// %s.%s", structName, fieldName))
		b.P(fmt.Sprintf("function encode_%d(uint64 pos, bytes memory buf, %s memory instance) internal pure returns (uint64) {", fieldNumber, structName))
		b.Indent()

		if isFieldRepeated(field) {
			// Repeated field

			if isFieldPacked(field) {
				// Packed repeated field

				switch fieldDescriptorType {
				case descriptorpb.FieldDescriptorProto_TYPE_ENUM:
					// Packed repeated enum

					_, err := g.getSolTypeName(field)
					if err != nil {
						return err
					}

					b.P(fmt.Sprintf("if (instance.%s.length > 0) {", fieldName))
					b.Indent()
					b.P("// Encode key")
					b.P(fmt.Sprintf("pos = ProtobufLib.encode_key(%d, ProtobufLib.WireType.LengthDelimited, pos, buf);", fieldNumber))
					b.P("")

					b.P("// Encode length")
					b.P("uint64 len_pos = pos;")
					b.P("pos += 1;")
					b.P("")

					b.P("// Encode elements")
					b.P(fmt.Sprintf("for (uint64 i = 0; i < instance.%s.length; i++) {", fieldName))
					b.Indent()
					b.P(fmt.Sprintf("pos = ProtobufLib.encode_enum(pos, buf, int32(instance.%s[i]));", fieldName))
					b.Unindent()
					b.P("}")
					b.P("")

					b.P("// Encode length")
					b.P("uint64 len = pos - len_pos - 1;")
					b.P("buf[len_pos] = bytes1(uint8(len));")
					b.Unindent()
					b.P("}")
				default:
					// Packed repeated numeric

					_, err := typeToSol(fieldDescriptorType)
					if err != nil {
						return errors.New(err.Error() + ": " + structName + "." + fieldName)
					}
					fieldEncodeType, err := typeToEncodeSol(fieldDescriptorType)
					if err != nil {
						return errors.New(err.Error() + ": " + structName + "." + fieldName)
					}

					b.P(fmt.Sprintf("if (instance.%s.length > 0) {", fieldName))
					b.Indent()
					b.P("// Encode key")
					b.P(fmt.Sprintf("pos = ProtobufLib.encode_key(%d, ProtobufLib.WireType.LengthDelimited, pos, buf);", fieldNumber))
					b.P("")

					b.P("// Encode length")
					b.P("uint64 len_pos = pos;")
					b.P("pos += 1;")
					b.P("")

					b.P("// Encode elements")
					b.P(fmt.Sprintf("for (uint64 i = 0; i < instance.%s.length; i++) {", fieldName))
					b.Indent()
					b.P(fmt.Sprintf("pos = %s(pos, buf, instance.%s[i]);", fieldEncodeType, fieldName))
					b.Unindent()
					b.P("}")
					b.P("")

					b.P("// Encode length")
					b.P("uint64 len = pos - len_pos - 1;")
					b.P("buf[len_pos] = bytes1(uint8(len));")
					b.Unindent()
					b.P("}")
				}
			} else {
				// Non-packed repeated field (i.e. message, string, or bytes)

				// Special handling for repeated string and bytes fields
				if fieldDescriptorType == descriptorpb.FieldDescriptorProto_TYPE_STRING || fieldDescriptorType == descriptorpb.FieldDescriptorProto_TYPE_BYTES {
					wrapperName := fmt.Sprintf("%sList", strings.Title(fieldName))
					
					b.P(fmt.Sprintf("for (uint64 i = 0; i < instance.%s.length; i++) {", fieldName))
					b.Indent()
					b.P("// Encode key")
					b.P(fmt.Sprintf("pos = ProtobufLib.encode_key(%d, ProtobufLib.WireType.LengthDelimited, pos, buf);", fieldNumber))
					b.P("")

					b.P("// Encode length")
					b.P("uint64 len_pos = pos;")
					b.P("pos += 1;")
					b.P("")

					b.P("// Encode wrapper message")
					b.P(fmt.Sprintf("pos = %sCodec.encode(pos, buf, instance.%s[i]);", wrapperName, fieldName))
					b.P("")

					b.P("// Encode length")
					b.P("uint64 len = pos - len_pos - 1;")
					b.P("buf[len_pos] = bytes1(uint8(len));")
					b.Unindent()
					b.P("}")
				} else {
					// Regular message field
					fieldTypeName, err := g.getSolTypeName(field)
					if err != nil {
						return err
					}

					b.P(fmt.Sprintf("for (uint64 i = 0; i < instance.%s.length; i++) {", fieldName))
					b.Indent()
					b.P("// Encode key")
					b.P(fmt.Sprintf("pos = ProtobufLib.encode_key(%d, ProtobufLib.WireType.LengthDelimited, pos, buf);", fieldNumber))
					b.P("")

					b.P("// Encode length")
					b.P("uint64 len_pos = pos;")
					b.P("pos += 1;")
					b.P("")

					b.P("// Encode message")
					b.P(fmt.Sprintf("pos = %sCodec.encode(pos, buf, instance.%s[i]);", fieldTypeName, fieldName))
					b.P("")

					b.P("// Encode length")
					b.P("uint64 len = pos - len_pos - 1;")
					b.P("buf[len_pos] = bytes1(uint8(len));")
					b.Unindent()
					b.P("}")
				}
			}
		} else {
			// Optional field (i.e. not repeated)

			switch fieldDescriptorType {
			case descriptorpb.FieldDescriptorProto_TYPE_ENUM:
				fieldTypeName, err := g.getSolTypeName(field)
				if err != nil {
					return err
				}

				b.P(fmt.Sprintf("if (instance.%s != %s(0)) {", fieldName, fieldTypeName))
				b.Indent()
				b.P("// Encode key")
				b.P(fmt.Sprintf("pos = ProtobufLib.encode_key(%d, ProtobufLib.WireType.Varint, pos, buf);", fieldNumber))
				b.P("")

				b.P("// Encode value")
				b.P(fmt.Sprintf("pos = ProtobufLib.encode_enum(pos, buf, int32(instance.%s));", fieldName))
				b.Unindent()
				b.P("}")
			case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
				fieldTypeName, err := g.getSolTypeName(field)
				if err != nil {
					return err
				}

				b.P(fmt.Sprintf("if (instance.%s.value.length > 0) {", fieldName))
				b.Indent()
				b.P("// Encode key")
				b.P(fmt.Sprintf("pos = ProtobufLib.encode_key(%d, ProtobufLib.WireType.LengthDelimited, pos, buf);", fieldNumber))
				b.P("")

				b.P("// Encode length")
				b.P("uint64 len_pos = pos;")
				b.P("pos += 1;")
				b.P("")

				b.P("// Encode message")
				b.P(fmt.Sprintf("pos = %sCodec.encode(pos, buf, instance.%s);", fieldTypeName, fieldName))
				b.P("")

				b.P("// Encode length")
				b.P("uint64 len = pos - len_pos - 1;")
				b.P("buf[len_pos] = bytes1(uint8(len));")
				b.Unindent()
				b.P("}")
			default:
				_, err := typeToSol(fieldDescriptorType)
				if err != nil {
					return errors.New(err.Error() + ": " + structName + "." + fieldName)
				}
				fieldEncodeType, err := typeToEncodeSol(fieldDescriptorType)
				if err != nil {
					return errors.New(err.Error() + ": " + structName + "." + fieldName)
				}

				switch fieldDescriptorType {
				case descriptorpb.FieldDescriptorProto_TYPE_INT32,
					descriptorpb.FieldDescriptorProto_TYPE_INT64,
					descriptorpb.FieldDescriptorProto_TYPE_UINT32,
					descriptorpb.FieldDescriptorProto_TYPE_UINT64,
					descriptorpb.FieldDescriptorProto_TYPE_SINT32,
					descriptorpb.FieldDescriptorProto_TYPE_SINT64,
					descriptorpb.FieldDescriptorProto_TYPE_FIXED32,
					descriptorpb.FieldDescriptorProto_TYPE_FIXED64,
					descriptorpb.FieldDescriptorProto_TYPE_SFIXED32,
					descriptorpb.FieldDescriptorProto_TYPE_SFIXED64,
					descriptorpb.FieldDescriptorProto_TYPE_FLOAT,
					descriptorpb.FieldDescriptorProto_TYPE_DOUBLE:
					b.P(fmt.Sprintf("if (instance.%s != 0) {", fieldName))
					b.Indent()
					b.P("// Encode key")
					b.P(fmt.Sprintf("pos = ProtobufLib.encode_key(%d, ProtobufLib.WireType.Varint, pos, buf);", fieldNumber))
					b.P("")

					b.P("// Encode value")
					b.P(fmt.Sprintf("pos = %s(pos, buf, instance.%s);", fieldEncodeType, fieldName))
					b.Unindent()
					b.P("}")
				case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
					b.P(fmt.Sprintf("if (instance.%s != false) {", fieldName))
					b.Indent()
					b.P("// Encode key")
					b.P(fmt.Sprintf("pos = ProtobufLib.encode_key(%d, ProtobufLib.WireType.Varint, pos, buf);", fieldNumber))
					b.P("")

					b.P("// Encode value")
					b.P(fmt.Sprintf("pos = %s(pos, buf, instance.%s);", fieldEncodeType, fieldName))
					b.Unindent()
					b.P("}")
				case descriptorpb.FieldDescriptorProto_TYPE_STRING:
					b.P(fmt.Sprintf("if (bytes(instance.%s).length > 0) {", fieldName))
					b.Indent()
					b.P("// Encode key")
					b.P(fmt.Sprintf("pos = ProtobufLib.encode_key(%d, ProtobufLib.WireType.LengthDelimited, pos, buf);", fieldNumber))
					b.P("")

					b.P("// Encode value")
					b.P(fmt.Sprintf("pos = %s(pos, buf, instance.%s);", fieldEncodeType, fieldName))
					b.Unindent()
					b.P("}")
				case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
					b.P(fmt.Sprintf("if (instance.%s.length > 0) {", fieldName))
					b.Indent()
					b.P("// Encode key")
					b.P(fmt.Sprintf("pos = ProtobufLib.encode_key(%d, ProtobufLib.WireType.LengthDelimited, pos, buf);", fieldNumber))
					b.P("")

					b.P("// Encode value")
					b.P(fmt.Sprintf("pos = %s(pos, buf, instance.%s);", fieldEncodeType, fieldName))
					b.Unindent()
					b.P("}")
				default:
					return errors.New("unsupported field type: " + fieldDescriptorType.String())
				}
			}
		}

		b.P("return pos;")
		b.Unindent()
		b.P("}")
		b.P("")
	}

	return nil
} 