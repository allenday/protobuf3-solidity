package generator

import (
	"errors"
	"fmt"
	"strings"

	"google.golang.org/protobuf/types/descriptorpb"
)

// generateMessageDecoder generates the decoder functions for a message
func (g *Generator) generateMessageDecoder(structName string, fields []*descriptorpb.FieldDescriptorProto, b *WriteableBuffer) error {
	// Top-level decoder function
	b.P(fmt.Sprintf("function decode(uint64 initial_pos, bytes memory buf, uint64 len) internal pure returns (bool, uint64, %s memory) {", structName))
	b.Indent()

	b.P("// Message instance")
	b.P(fmt.Sprintf("%s memory instance;", structName))
	b.P("// Previous field number")
	b.P("uint64 previous_field_number = 0;")
	b.P("// Current position in the buffer")
	b.P("uint64 pos = initial_pos;")
	b.P()

	b.P("// Sanity checks")
	b.P("if (pos + len < pos) {")
	b.Indent()
	b.P("return (false, pos, instance);")
	b.Unindent()
	b.P("}")
	b.P()

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
	b.P()

	b.P("// Check that the field number is within bounds")
	b.P(fmt.Sprintf("if (field_number > %d) {", len(fields)))
	b.Indent()
	b.P("return (false, pos, instance);")
	b.Unindent()
	b.P("}")
	b.P()

	b.P("// Check that the field number of monotonically increasing")
	if !g.allowNonMonotonicFields {
		b.P("if (field_number <= previous_field_number) {")
		b.Indent()
		b.P("return (false, pos, instance);")
		b.Unindent()
		b.P("}")
	}
	b.P()

	b.P("// Check that the wire type is correct")
	b.P("success = check_key(field_number, wire_type);")
	b.P("if (!success) {")
	b.Indent()
	b.P("return (false, pos, instance);")
	b.Unindent()
	b.P("}")
	b.P()

	b.P("// Actually decode the field")
	b.P("(success, pos) = decode_field(pos, buf, len, field_number, instance);")
	b.P("if (!success) {")
	b.Indent()
	b.P("return (false, pos, instance);")
	b.Unindent()
	b.P("}")
	b.P()

	b.P("previous_field_number = field_number;")
	b.Unindent()
	b.P("}")
	b.P()

	b.P("// Decoding must have consumed len bytes")
	b.P("if (pos != initial_pos + len) {")
	b.Indent()
	b.P("return (false, pos, instance);")
	b.Unindent()
	b.P("}")
	b.P()

	b.P("return (true, pos, instance);")
	b.Unindent()
	b.P("}")
	b.P()

	// Check key function
	b.P("function check_key(uint64 field_number, ProtobufLib.WireType wire_type) internal pure returns (bool) {")
	b.Indent()
	for _, field := range fields {
		fieldNumber := field.GetNumber()

		b.P(fmt.Sprintf("if (field_number == %d) {", fieldNumber))
		b.Indent()
		wireStr, err := toSolWireType(field)
		if err != nil {
			return err
		}
		b.P(fmt.Sprintf("return wire_type == %s;", wireStr))
		b.Unindent()
		b.P("}")
		b.P()
	}

	b.P("return false;")
	b.Unindent()
	b.P("}")
	b.P()

	// Decode field dispatcher function
	b.P(fmt.Sprintf("function decode_field(uint64 initial_pos, bytes memory buf, uint64 len, uint64 field_number, %s memory instance) internal pure returns (bool, uint64) {", structName))
	b.Indent()
	b.P("uint64 pos = initial_pos;")
	b.P()

	for _, field := range fields {
		fieldNumber := field.GetNumber()

		b.P(fmt.Sprintf("if (field_number == %d) {", fieldNumber))
		b.Indent()
		b.P("bool success;")
		b.P(fmt.Sprintf("(success, pos) = decode_%d(pos, buf, instance);", fieldNumber))
		b.P("if (!success) {")
		b.Indent()
		b.P("return (false, pos);")
		b.Unindent()
		b.P("}")
		b.P()

		b.P("return (true, pos);")
		b.Unindent()
		b.P("}")
		b.P()
	}

	b.P("return (false, pos);")
	b.Unindent()
	b.P("}")
	b.P()

	// Individual field decoders
	for _, field := range fields {
		fieldName := field.GetName()
		fieldDescriptorType := field.GetType()
		fieldNumber := field.GetNumber()

		b.P(fmt.Sprintf("// %s.%s", structName, fieldName))
		b.P(fmt.Sprintf("function decode_%d(uint64 pos, bytes memory buf, %s memory instance) internal pure returns (bool, uint64) {", fieldNumber, structName))
		b.Indent()

		b.P("bool success;")
		b.P()

		if isFieldRepeated(field) {
			// Repeated field

			if isFieldPacked(field) {
				// Packed repeated field

				switch fieldDescriptorType {
				case descriptorpb.FieldDescriptorProto_TYPE_ENUM:
					// Packed repeated enum

					fieldTypeName, err := g.getSolTypeName(field)
					if err != nil {
						return err
					}

					b.P("uint64 len;")
					b.P(fmt.Sprintf("(success, pos, len) = ProtobufLib.decode_length_delimited(pos, buf);"))
					b.P("if (!success) {")
					b.Indent()
					b.P("return (false, pos);")
					b.Unindent()
					b.P("}")
					b.P()

					b.P("// Empty packed array must be omitted")
					if !g.allowEmptyPackedArrays {
						b.P("if (len == 0) {")
						b.Indent()
						b.P("return (false, pos);")
						b.Unindent()
						b.P("}")
					}
					b.P()

					b.P("uint64 initial_pos = pos;")
					b.P()

					b.P("// Sanity checks")
					b.P("if (initial_pos + len < initial_pos) {")
					b.Indent()
					b.P("return (false, pos);")
					b.Unindent()
					b.P("}")
					b.P()

					b.P("// Do one pass to count the number of elements")
					b.P("uint64 cnt = 0;")
					b.P("while (pos - initial_pos < len) {")
					b.Indent()
					b.P("int32 v;")
					b.P("(success, pos, v) = ProtobufLib.decode_enum(pos, buf);")
					b.P("if (!success) {")
					b.Indent()
					b.P("return (false, pos);")
					b.Unindent()
					b.P("}")
					b.P("cnt += 1;")
					b.Unindent()
					b.P("}")
					b.P()

					b.P("// Allocated memory")
					b.P(fmt.Sprintf("instance.%s = new %s[](cnt);", fieldName, fieldTypeName))
					b.P()

					b.P("// Now actually parse the elements")
					b.P("pos = initial_pos;")
					b.P("for (uint64 i = 0; i < cnt; i++) {")
					b.Indent()
					b.P("int32 v;")
					b.P("(success, pos, v) = ProtobufLib.decode_enum(pos, buf);")
					b.P("if (!success) {")
					b.Indent()
					b.P("return (false, pos);")
					b.Unindent()
					b.P("}")
					b.P()

					b.P("// Check that value is within enum range")
					b.P(fmt.Sprintf("if (v < 0 || v > %d) {", g.enumMaxes[fieldTypeName]))
					b.Indent()
					b.P("return (false, pos);")
					b.Unindent()
					b.P("}")
					b.P()

					b.P(fmt.Sprintf("instance.%s[i] = %s(v);", fieldName, fieldTypeName))
					b.Unindent()
					b.P("}")
					b.P()

					b.P("// Decoding must have consumed len bytes")
					b.P("if (pos != initial_pos + len) {")
					b.Indent()
					b.P("return (false, pos);")
					b.Unindent()
					b.P("}")
					b.P()
				default:
					// Packed repeated numeric

					fieldType, err := typeToSol(fieldDescriptorType)
					if err != nil {
						return errors.New(err.Error() + ": " + structName + "." + fieldName)
					}
					fieldDecodeType, err := typeToDecodeSol(fieldDescriptorType)
					if err != nil {
						return errors.New(err.Error() + ": " + structName + "." + fieldName)
					}

					b.P("uint64 len;")
					b.P(fmt.Sprintf("(success, pos, len) = ProtobufLib.decode_length_delimited(pos, buf);"))
					b.P("if (!success) {")
					b.Indent()
					b.P("return (false, pos);")
					b.Unindent()
					b.P("}")
					b.P()

					b.P("// Empty packed array must be omitted")
					if !g.allowEmptyPackedArrays {
						b.P("if (len == 0) {")
						b.Indent()
						b.P("return (false, pos);")
						b.Unindent()
						b.P("}")
					}
					b.P()

					b.P("uint64 initial_pos = pos;")
					b.P()

					b.P("// Sanity checks")
					b.P("if (pos + len < pos) {")
					b.Indent()
					b.P("return (false, pos);")
					b.Unindent()
					b.P("}")
					b.P()

					b.P("// Do one pass to count the number of elements")
					b.P("uint64 cnt = 0;")
					b.P("while (pos - initial_pos < len) {")
					b.Indent()
					b.P(fmt.Sprintf("%s v;", fieldType))
					b.P(fmt.Sprintf("(success, pos, v) = ProtobufLib.decode_%s(pos, buf);", fieldDecodeType))
					b.P("if (!success) {")
					b.Indent()
					b.P("return (false, pos);")
					b.Unindent()
					b.P("}")
					b.P("cnt += 1;")
					b.Unindent()
					b.P("}")
					b.P()

					b.P("// Allocated memory")
					b.P(fmt.Sprintf("instance.%s = new %s[](cnt);", fieldName, fieldType))
					b.P()

					b.P("// Now actually parse the elements")
					b.P("pos = initial_pos;")
					b.P("for (uint64 i = 0; i < cnt; i++) {")
					b.Indent()
					b.P(fmt.Sprintf("%s v;", fieldType))
					b.P(fmt.Sprintf("(success, pos, v) = ProtobufLib.decode_%s(pos, buf);", fieldDecodeType))
					b.P("if (!success) {")
					b.Indent()
					b.P("return (false, pos);")
					b.Unindent()
					b.P("}")
					b.P()

					b.P(fmt.Sprintf("instance.%s[i] = v;", fieldName))
					b.Unindent()
					b.P("}")
					b.P()

					b.P("// Decoding must have consumed len bytes")
					b.P("if (pos != initial_pos + len) {")
					b.Indent()
					b.P("return (false, pos);")
					b.Unindent()
					b.P("}")
					b.P()
				}
			} else {
				// Non-packed repeated field (i.e. message, string, or bytes)
				
				// Special handling for repeated string and bytes fields
				if fieldDescriptorType == descriptorpb.FieldDescriptorProto_TYPE_STRING || fieldDescriptorType == descriptorpb.FieldDescriptorProto_TYPE_BYTES {
					wrapperName := fmt.Sprintf("%sList", strings.Title(fieldName))
					
					b.P("uint64 initial_pos = pos;")
					b.P()

					b.P("// Do one pass to count the number of elements")
					b.P("uint64 cnt = 0;")
					b.P("while (pos < buf.length) {")
					b.Indent()
					b.P("uint64 len;")
					b.P("(success, pos, len) = ProtobufLib.decode_length_delimited(pos, buf);")
					b.P("if (!success) {")
					b.Indent()
					b.P("return (false, pos);")
					b.Unindent()
					b.P("}")
					b.P()

					b.P("// Sanity checks")
					b.P("if (pos + len < pos) {")
					b.Indent()
					b.P("return (false, pos);")
					b.Unindent()
					b.P("}")
					b.P()

					b.P("pos += len;")
					b.P("cnt += 1;")
					b.P()

					b.P("if (pos >= buf.length) {")
					b.Indent()
					b.P("break;")
					b.Unindent()
					b.P("}")
					b.P()

					b.P("// Decode next key")
					b.P("uint64 field_number;")
					b.P("ProtobufLib.WireType wire_type;")
					b.P("(success, pos, field_number, wire_type) = ProtobufLib.decode_key(pos, buf);")
					b.P("if (!success) {")
					b.Indent()
					b.P("return (false, pos);")
					b.Unindent()
					b.P("}")
					b.P()

					b.P("// Check if the field number is different")
					b.P(fmt.Sprintf("if (field_number != %d) {", fieldNumber))
					b.Indent()
					b.P("break;")
					b.Unindent()
					b.P("}")
					b.Unindent()
					b.P("}")
					b.P()

					b.P("// Allocated memory")
					b.P(fmt.Sprintf("instance.%s = new %s[](cnt);", fieldName, wrapperName))
					b.P()

					b.P("// Now actually parse the elements")
					b.P("pos = initial_pos;")
					b.P("for (uint64 i = 0; i < cnt; i++) {")
					b.Indent()
					b.P("uint64 len;")
					b.P("(success, pos, len) = ProtobufLib.decode_length_delimited(pos, buf);")
					b.P("if (!success) {")
					b.Indent()
					b.P("return (false, pos);")
					b.Unindent()
					b.P("}")
					b.P()

					b.P("initial_pos = pos;")
					b.P()

					b.P(fmt.Sprintf("%s memory nestedInstance;", wrapperName))
					b.P(fmt.Sprintf("(success, pos, nestedInstance) = %sCodec.decode(pos, buf, len);", wrapperName))
					b.P("if (!success) {")
					b.Indent()
					b.P("return (false, pos);")
					b.Unindent()
					b.P("}")
					b.P()

					b.P(fmt.Sprintf("instance.%s[i] = nestedInstance;", fieldName))
					b.P()

					b.P("// Skip over next key, reuse len")
					b.P("if (i < cnt - 1) {")
					b.Indent()
					b.P("(success, pos, len) = ProtobufLib.decode_uint64(pos, buf);")
					b.P("if (!success) {")
					b.Indent()
					b.P("return (false, pos);")
					b.Unindent()
					b.P("}")
					b.Unindent()
					b.P("}")
					b.Unindent()
					b.P("}")
					b.P()
				} else {
					// Regular message field
					fieldTypeName, err := g.getSolTypeName(field)
					if err != nil {
						return err
					}

					b.P("uint64 initial_pos = pos;")
					b.P()

					b.P("// Do one pass to count the number of elements")
					b.P("uint64 cnt = 0;")
					b.P("while (pos < buf.length) {")
					b.Indent()
					b.P("uint64 len;")
					b.P("(success, pos, len) = ProtobufLib.decode_embedded_message(pos, buf);")
					b.P("if (!success) {")
					b.Indent()
					b.P("return (false, pos);")
					b.Unindent()
					b.P("}")
					b.P()

					b.P("// Sanity checks")
					b.P("if (pos + len < pos) {")
					b.Indent()
					b.P("return (false, pos);")
					b.Unindent()
					b.P("}")
					b.P()

					b.P("pos += len;")
					b.P("cnt += 1;")
					b.P()

					b.P("if (pos >= buf.length) {")
					b.Indent()
					b.P("break;")
					b.Unindent()
					b.P("}")
					b.P()

					b.P("// Decode next key")
					b.P("uint64 field_number;")
					b.P("ProtobufLib.WireType wire_type;")
					b.P("(success, pos, field_number, wire_type) = ProtobufLib.decode_key(pos, buf);")
					b.P("if (!success) {")
					b.Indent()
					b.P("return (false, pos);")
					b.Unindent()
					b.P("}")
					b.P()

					b.P("// Check if the field number is different")
					b.P(fmt.Sprintf("if (field_number != %d) {", fieldNumber))
					b.Indent()
					b.P("break;")
					b.Unindent()
					b.P("}")
					b.Unindent()
					b.P("}")
					b.P()

					b.P("// Allocated memory")
					b.P(fmt.Sprintf("instance.%s = new %s[](cnt);", fieldName, fieldTypeName))
					b.P()

					b.P("// Now actually parse the elements")
					b.P("pos = initial_pos;")
					b.P("for (uint64 i = 0; i < cnt; i++) {")
					b.Indent()
					b.P("uint64 len;")
					b.P("(success, pos, len) = ProtobufLib.decode_embedded_message(pos, buf);")
					b.P("if (!success) {")
					b.Indent()
					b.P("return (false, pos);")
					b.Unindent()
					b.P("}")
					b.P()

					b.P("initial_pos = pos;")
					b.P()

					b.P(fmt.Sprintf("%s memory nestedInstance;", fieldTypeName))
					b.P(fmt.Sprintf("(success, pos, nestedInstance) = %sCodec.decode(pos, buf, len);", fieldTypeName))
					b.P("if (!success) {")
					b.Indent()
					b.P("return (false, pos);")
					b.Unindent()
					b.P("}")
					b.P()

					b.P(fmt.Sprintf("instance.%s[i] = nestedInstance;", fieldName))
					b.P()

					b.P("// Skip over next key, reuse len")
					b.P("if (i < cnt - 1) {")
					b.Indent()
					b.P("(success, pos, len) = ProtobufLib.decode_uint64(pos, buf);")
					b.P("if (!success) {")
					b.Indent()
					b.P("return (false, pos);")
					b.Unindent()
					b.P("}")
					b.Unindent()
					b.P("}")
					b.Unindent()
					b.P("}")
					b.P()
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

				b.P("int32 v;")
				b.P("(success, pos, v) = ProtobufLib.decode_enum(pos, buf);")
				b.P("if (!success) {")
				b.Indent()
				b.P("return (false, pos);")
				b.Unindent()
				b.P("}")
				b.P()

				b.P("// Default value must be omitted")
				b.P("if (v == 0) {")
				b.Indent()
				b.P("return (false, pos);")
				b.Unindent()
				b.P("}")
				b.P()

				b.P("// Check that value is within enum range")
				b.P(fmt.Sprintf("if (v < 0 || v > %d) {", g.enumMaxes[fieldTypeName]))
				b.Indent()
				b.P("return (false, pos);")
				b.Unindent()
				b.P("}")
				b.P()

				b.P(fmt.Sprintf("instance.%s = %s(v);", fieldName, fieldTypeName))
				b.P()
			case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
				// TODO check for default value of empty message
				fieldTypeName, err := g.getSolTypeName(field)
				if err != nil {
					return err
				}

				b.P("uint64 len;")
				b.P("(success, pos, len) = ProtobufLib.decode_embedded_message(pos, buf);")
				b.P("if (!success) {")
				b.Indent()
				b.P("return (false, pos);")
				b.Unindent()
				b.P("}")
				b.P()

				b.P("// Default value must be omitted")
				b.P("if (len == 0) {")
				b.Indent()
				b.P("return (false, pos);")
				b.Unindent()
				b.P("}")
				b.P()

				b.P(fmt.Sprintf("%s memory nestedInstance;", fieldTypeName))
				b.P(fmt.Sprintf("(success, pos, nestedInstance) = %sCodec.decode(pos, buf, len);", fieldTypeName))
				b.P("if (!success) {")
				b.Indent()
				b.P("return (false, pos);")
				b.Unindent()
				b.P("}")
				b.P()

				b.P(fmt.Sprintf("instance.%s = nestedInstance;", fieldName))
				b.P()
			default:
				fieldType, err := typeToSol(fieldDescriptorType)
				if err != nil {
					return errors.New(err.Error() + ": " + structName + "." + fieldName)
				}
				fieldDecodeType, err := typeToDecodeSol(fieldDescriptorType)
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
					b.P(fmt.Sprintf("%s v;", fieldType))
					b.P(fmt.Sprintf("(success, pos, v) = ProtobufLib.decode_%s(pos, buf);", fieldDecodeType))
					b.P("if (!success) {")
					b.Indent()
					b.P("return (false, pos);")
					b.Unindent()
					b.P("}")
					b.P()

					b.P("// Default value must be omitted")
					b.P("if (v == 0) {")
					b.Indent()
					b.P("return (false, pos);")
					b.Unindent()
					b.P("}")
					b.P()

					b.P(fmt.Sprintf("instance.%s = v;", fieldName))
					b.P()
				case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
					b.P(fmt.Sprintf("%s v;", fieldType))
					b.P(fmt.Sprintf("(success, pos, v) = ProtobufLib.decode_%s(pos, buf);", fieldDecodeType))
					b.P("if (!success) {")
					b.Indent()
					b.P("return (false, pos);")
					b.Unindent()
					b.P("}")
					b.P()

					b.P("// Default value must be omitted")
					b.P("if (v == false) {")
					b.Indent()
					b.P("return (false, pos);")
					b.Unindent()
					b.P("}")
					b.P()

					b.P(fmt.Sprintf("instance.%s = v;", fieldName))
					b.P()
				case descriptorpb.FieldDescriptorProto_TYPE_STRING:
					b.P(fmt.Sprintf("%s memory v;", fieldType))
					b.P(fmt.Sprintf("(success, pos, v) = ProtobufLib.decode_%s(pos, buf);", fieldDecodeType))
					b.P("if (!success) {")
					b.Indent()
					b.P("return (false, pos);")
					b.Unindent()
					b.P("}")
					b.P()

					b.P("// Default value must be omitted")
					b.P("if (bytes(v).length == 0) {")
					b.Indent()
					b.P("return (false, pos);")
					b.Unindent()
					b.P("}")
					b.P()

					b.P(fmt.Sprintf("instance.%s = v;", fieldName))
					b.P()
				case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
					b.P("uint64 len;")
					b.P(fmt.Sprintf("(success, pos, len) = ProtobufLib.decode_%s(pos, buf);", fieldDecodeType))
					b.P("if (!success) {")
					b.Indent()
					b.P("return (false, pos);")
					b.Unindent()
					b.P("}")
					b.P()

					b.P("// Default value must be omitted")
					b.P("if (len == 0) {")
					b.Indent()
					b.P("return (false, pos);")
					b.Unindent()
					b.P("}")
					b.P()

					b.P(fmt.Sprintf("instance.%s = new bytes(len);", fieldName))
					b.P("for (uint64 i = 0; i < len; i++) {")
					b.Indent()
					b.P(fmt.Sprintf("instance.%s[i] = buf[pos + i];", fieldName))
					b.Unindent()
					b.P("}")
					b.P()

					b.P("pos = pos + len;")
					b.P()
				default:
					return errors.New("unsupported field type: " + fieldDescriptorType.String())
				}
			}
		}

		b.P("return (true, pos);")
		b.Unindent()
		b.P("}")
		b.P()
	}

	return nil
}

// generateMessageEncoder generates the encoder functions for a message
func (g *Generator) generateMessageEncoder(structName string, fields []*descriptorpb.FieldDescriptorProto, b *WriteableBuffer) error {
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
	b.P()

	// Individual field encoders
	for _, field := range fields {
		fieldName := field.GetName()
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
					b.P()

					b.P("// Encode length")
					b.P("uint64 len_pos = pos;")
					b.P("pos += 1;")
					b.P()

					b.P("// Encode elements")
					b.P(fmt.Sprintf("for (uint64 i = 0; i < instance.%s.length; i++) {", fieldName))
					b.Indent()
					b.P(fmt.Sprintf("pos = ProtobufLib.encode_enum(pos, buf, int32(instance.%s[i]));", fieldName))
					b.Unindent()
					b.P("}")
					b.P()

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
					b.P()

					b.P("// Encode length")
					b.P("uint64 len_pos = pos;")
					b.P("pos += 1;")
					b.P()

					b.P("// Encode elements")
					b.P(fmt.Sprintf("for (uint64 i = 0; i < instance.%s.length; i++) {", fieldName))
					b.Indent()
					b.P(fmt.Sprintf("pos = %s(pos, buf, instance.%s[i]);", fieldEncodeType, fieldName))
					b.Unindent()
					b.P("}")
					b.P()

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
					b.P()

					b.P("// Encode length")
					b.P("uint64 len_pos = pos;")
					b.P("pos += 1;")
					b.P()

					b.P("// Encode wrapper message")
					b.P(fmt.Sprintf("pos = %sCodec.encode(pos, buf, instance.%s[i]);", wrapperName, fieldName))
					b.P()

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
					b.P()

					b.P("// Encode length")
					b.P("uint64 len_pos = pos;")
					b.P("pos += 1;")
					b.P()

					b.P("// Encode message")
					b.P(fmt.Sprintf("pos = %sCodec.encode(pos, buf, instance.%s[i]);", fieldTypeName, fieldName))
					b.P()

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
				b.P()

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
				b.P()

				b.P("// Encode length")
				b.P("uint64 len_pos = pos;")
				b.P("pos += 1;")
				b.P()

				b.P("// Encode message")
				b.P(fmt.Sprintf("pos = %sCodec.encode(pos, buf, instance.%s);", fieldTypeName, fieldName))
				b.P()

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
					b.P()

					b.P("// Encode value")
					b.P(fmt.Sprintf("pos = %s(pos, buf, instance.%s);", fieldEncodeType, fieldName))
					b.Unindent()
					b.P("}")
				case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
					b.P(fmt.Sprintf("if (instance.%s != false) {", fieldName))
					b.Indent()
					b.P("// Encode key")
					b.P(fmt.Sprintf("pos = ProtobufLib.encode_key(%d, ProtobufLib.WireType.Varint, pos, buf);", fieldNumber))
					b.P()

					b.P("// Encode value")
					b.P(fmt.Sprintf("pos = %s(pos, buf, instance.%s);", fieldEncodeType, fieldName))
					b.Unindent()
					b.P("}")
				case descriptorpb.FieldDescriptorProto_TYPE_STRING:
					b.P(fmt.Sprintf("if (bytes(instance.%s).length > 0) {", fieldName))
					b.Indent()
					b.P("// Encode key")
					b.P(fmt.Sprintf("pos = ProtobufLib.encode_key(%d, ProtobufLib.WireType.LengthDelimited, pos, buf);", fieldNumber))
					b.P()

					b.P("// Encode value")
					b.P(fmt.Sprintf("pos = %s(pos, buf, instance.%s);", fieldEncodeType, fieldName))
					b.Unindent()
					b.P("}")
				case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
					b.P(fmt.Sprintf("if (instance.%s.length > 0) {", fieldName))
					b.Indent()
					b.P("// Encode key")
					b.P(fmt.Sprintf("pos = ProtobufLib.encode_key(%d, ProtobufLib.WireType.LengthDelimited, pos, buf);", fieldNumber))
					b.P()

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
		b.P()
	}

	return nil
} 

// generateFloatDoubleHelpers generates helper functions for float/double scaling
func (g *Generator) generateFloatDoubleHelpers(b *WriteableBuffer) {
	b.P("// Helper functions for float/double fixed-point scaling")
	b.P()
	
	// Float scaling helper (1e6 precision)
	b.P("function decode_float_scaled(uint64 pos, bytes memory buf) internal pure returns (bool, uint64, int32) {")
	b.Indent()
	b.P("bool success;")
	b.P("uint64 new_pos;")
	b.P("uint32 raw_value;")
	b.P("(success, new_pos, raw_value) = ProtobufLib.decode_fixed32(pos, buf);")
	b.P("if (!success) {")
	b.Indent()
	b.P("return (false, pos, 0);")
	b.Unindent()
	b.P("}")
	b.P()
	b.P("// Convert IEEE 754 float to fixed-point int32 with 1e6 scaling")
	b.P("// This preserves 6 decimal places of precision")
	b.P("int32 scaled_value;")
	b.P("assembly {")
	b.Indent()
	b.P("// Extract sign, exponent, and mantissa from IEEE 754")
	b.P("let sign := shr(31, raw_value)")
	b.P("let exponent := and(shr(23, raw_value), 0xFF)")
	b.P("let mantissa := and(raw_value, 0x7FFFFF)")
	b.P()
	b.P("// Handle special cases")
	b.P("if eq(exponent, 0) {")
	b.Indent()
	b.P("// Zero or denormalized")
	b.P("scaled_value := 0")
	b.Unindent()
	b.P("}")
	b.P("if eq(exponent, 0xFF) {")
	b.Indent()
	b.P("// Infinity or NaN - return max value")
	b.P("scaled_value := 0x7FFFFFFF")
	b.Unindent()
	b.P("}")
	b.P()
	b.P("// Normal case: convert to fixed-point")
	b.P("// Add implicit leading 1 to mantissa")
	b.P("mantissa := or(mantissa, 0x800000)")
	b.P()
	b.P("// Calculate actual value: mantissa * 2^(exponent-127)")
	b.P("let shift := sub(exponent, 127)")
	b.P("let scaled_mantissa := mantissa")
	b.P()
	b.P("// Apply scaling factor of 1e6 (1,000,000)")
	b.P("scaled_mantissa := mul(scaled_mantissa, 1000000)")
	b.P()
	b.P("// Apply exponent shift")
	b.P("if gt(shift, 0) {")
	b.Indent()
	b.P("scaled_mantissa := shl(shift, scaled_mantissa)")
	b.Unindent()
	b.P("}")
	b.P("if lt(shift, 0) {")
	b.Indent()
	b.P("scaled_mantissa := shr(sub(0, shift), scaled_mantissa)")
	b.Unindent()
	b.P("}")
	b.P()
	b.P("// Apply sign")
	b.P("if sign {")
	b.Indent()
	b.P("scaled_value := sub(0, scaled_mantissa)")
	b.Unindent()
	b.P("}")
	b.P("if iszero(sign) {")
	b.Indent()
	b.P("scaled_value := scaled_mantissa")
	b.Unindent()
	b.P("}")
	b.Unindent()
	b.P("}")
	b.P()
	b.P("return (true, new_pos, scaled_value);")
	b.Unindent()
	b.P("}")
	b.P()
	
	// Double scaling helper (1e15 precision)
	b.P("function decode_double_scaled(uint64 pos, bytes memory buf) internal pure returns (bool, uint64, int64) {")
	b.Indent()
	b.P("bool success;")
	b.P("uint64 new_pos;")
	b.P("uint64 raw_value;")
	b.P("(success, new_pos, raw_value) = ProtobufLib.decode_fixed64(pos, buf);")
	b.P("if (!success) {")
	b.Indent()
	b.P("return (false, pos, 0);")
	b.Unindent()
	b.P("}")
	b.P()
	b.P("// Convert IEEE 754 double to fixed-point int64 with 1e15 scaling")
	b.P("// This preserves 15 decimal places of precision")
	b.P("int64 scaled_value;")
	b.P("assembly {")
	b.Indent()
	b.P("// Extract sign, exponent, and mantissa from IEEE 754")
	b.P("let sign := shr(63, raw_value)")
	b.P("let exponent := and(shr(52, raw_value), 0x7FF)")
	b.P("let mantissa := and(raw_value, 0xFFFFFFFFFFFFF)")
	b.P()
	b.P("// Handle special cases")
	b.P("if eq(exponent, 0) {")
	b.Indent()
	b.P("// Zero or denormalized")
	b.P("scaled_value := 0")
	b.Unindent()
	b.P("}")
	b.P("if eq(exponent, 0x7FF) {")
	b.Indent()
	b.P("// Infinity or NaN - return max value")
	b.P("scaled_value := 0x7FFFFFFFFFFFFFFF")
	b.Unindent()
	b.P("}")
	b.P()
	b.P("// Normal case: convert to fixed-point")
	b.P("// Add implicit leading 1 to mantissa")
	b.P("mantissa := or(mantissa, 0x10000000000000)")
	b.P()
	b.P("// Calculate actual value: mantissa * 2^(exponent-1023)")
	b.P("let shift := sub(exponent, 1023)")
	b.P("let scaled_mantissa := mantissa")
	b.P()
	b.P("// Apply scaling factor of 1e15 (1,000,000,000,000,000)")
	b.P("scaled_mantissa := mul(scaled_mantissa, 1000000000000000)")
	b.P()
	b.P("// Apply exponent shift")
	b.P("if gt(shift, 0) {")
	b.Indent()
	b.P("scaled_mantissa := shl(shift, scaled_mantissa)")
	b.Unindent()
	b.P("}")
	b.P("if lt(shift, 0) {")
	b.Indent()
	b.P("scaled_mantissa := shr(sub(0, shift), scaled_mantissa)")
	b.Unindent()
	b.P("}")
	b.P()
	b.P("// Apply sign")
	b.P("if sign {")
	b.Indent()
	b.P("scaled_value := sub(0, scaled_mantissa)")
	b.Unindent()
	b.P("}")
	b.P("if iszero(sign) {")
	b.Indent()
	b.P("scaled_value := scaled_mantissa")
	b.Unindent()
	b.P("}")
	b.Unindent()
	b.P("}")
	b.P()
	b.P("return (true, new_pos, scaled_value);")
	b.Unindent()
	b.P("}")
	b.P()
	
	// Encode helpers for float/double
	b.P("function encode_float_scaled(uint64 pos, bytes memory buf, int32 value) internal pure returns (uint64) {")
	b.Indent()
	b.P("// Convert fixed-point int32 back to IEEE 754 float")
	b.P("uint32 raw_value;")
	b.P("assembly {")
	b.Indent()
	b.P("// Extract sign")
	b.P("let sign := slt(value, 0)")
	b.P("let abs_value := value")
	b.P("if sign {")
	b.Indent()
	b.P("abs_value := sub(0, value)")
	b.Unindent()
	b.P("}")
	b.P()
	b.P("// Convert from fixed-point (1e6 scaling) to float")
	b.P("// This is a simplified conversion - in practice, you'd want more precision")
	b.P("let float_value := abs_value")
	b.P()
	b.P("// Normalize to IEEE 754 format")
	b.P("let exponent := 127")
	b.P("let mantissa := float_value")
	b.P()
	b.P("// Find the highest bit set")
	b.P("let highest_bit := 0")
	b.P("for { } lt(highest_bit, 32) { highest_bit := add(highest_bit, 1) } {")
	b.Indent()
	b.P("if gt(and(mantissa, shl(highest_bit, 1)), 0) {")
	b.Indent()
	b.P("break")
	b.Unindent()
	b.P("}")
	b.Unindent()
	b.P("}")
	b.P()
	b.P("// Adjust exponent and mantissa")
	b.P("if gt(highest_bit, 0) {")
	b.Indent()
	b.P("exponent := add(exponent, sub(23, highest_bit))")
	b.P("mantissa := shr(sub(highest_bit, 23), mantissa)")
	b.Unindent()
	b.P("}")
	b.P()
	b.P("// Remove implicit leading 1")
	b.P("mantissa := and(mantissa, 0x7FFFFF)")
	b.P()
	b.P("// Combine into IEEE 754 format")
	b.P("raw_value := or(shl(31, sign), or(shl(23, exponent), mantissa))")
	b.Unindent()
	b.P("}")
	b.P()
	b.P("return ProtobufLib.encode_fixed32(pos, buf, raw_value);")
	b.Unindent()
	b.P("}")
	b.P()
	
	b.P("function encode_double_scaled(uint64 pos, bytes memory buf, int64 value) internal pure returns (uint64) {")
	b.Indent()
	b.P("// Convert fixed-point int64 back to IEEE 754 double")
	b.P("uint64 raw_value;")
	b.P("assembly {")
	b.Indent()
	b.P("// Extract sign")
	b.P("let sign := slt(value, 0)")
	b.P("let abs_value := value")
	b.P("if sign {")
	b.Indent()
	b.P("abs_value := sub(0, value)")
	b.Unindent()
	b.P("}")
	b.P()
	b.P("// Convert from fixed-point (1e15 scaling) to double")
	b.P("// This is a simplified conversion - in practice, you'd want more precision")
	b.P("let double_value := abs_value")
	b.P()
	b.P("// Normalize to IEEE 754 format")
	b.P("let exponent := 1023")
	b.P("let mantissa := double_value")
	b.P()
	b.P("// Find the highest bit set")
	b.P("let highest_bit := 0")
	b.P("for { } lt(highest_bit, 64) { highest_bit := add(highest_bit, 1) } {")
	b.Indent()
	b.P("if gt(and(mantissa, shl(highest_bit, 1)), 0) {")
	b.Indent()
	b.P("break")
	b.Unindent()
	b.P("}")
	b.Unindent()
	b.P("}")
	b.P()
	b.P("// Adjust exponent and mantissa")
	b.P("if gt(highest_bit, 0) {")
	b.Indent()
	b.P("exponent := add(exponent, sub(52, highest_bit))")
	b.P("mantissa := shr(sub(highest_bit, 52), mantissa)")
	b.Unindent()
	b.P("}")
	b.P()
	b.P("// Remove implicit leading 1")
	b.P("mantissa := and(mantissa, 0xFFFFFFFFFFFFF)")
	b.P()
	b.P("// Combine into IEEE 754 format")
	b.P("raw_value := or(shl(63, sign), or(shl(52, exponent), mantissa))")
	b.Unindent()
	b.P("}")
	b.P()
	b.P("return ProtobufLib.encode_fixed64(pos, buf, raw_value);")
	b.Unindent()
	b.P("}")
	b.P()
} 