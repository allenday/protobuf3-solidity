package generator

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"google.golang.org/protobuf/types/descriptorpb"
)

// generateEnum generates Solidity enum code from a protobuf enum descriptor
func (g *Generator) generateEnum(descriptor *descriptorpb.EnumDescriptorProto, b *WriteableBuffer) error {
	enumName := sanitizeKeyword(descriptor.GetName())
	enumValues := descriptor.GetValue()

	// Note: we don't need this check since it's enforced by protoc, but keep it just in case
	if len(enumValues) == 0 {
		return errors.New("enums must have at least one value: " + enumName)
	}

	enumNamesString := ""
	oldValue := -1
	for _, enumValue := range enumValues {
		if oldValue != -1 {
			enumNamesString += ", "
		}

		name := enumValue.GetName()
		value := int(enumValue.GetNumber())

		enumNamesString += name

		// Configurable enum validation - can be disabled via parameters
		if g.strictEnumValidation && value != oldValue+1 {
			return errors.New("enums must start at 0 and increment by 1: " + enumName + "." + name)
		}
		oldValue = value
	}

	b.P(fmt.Sprintf("enum %s { %s }", enumName, enumNamesString))
	b.P()

	// Store the maximum enum value for later use
	g.enumMaxes[enumName] = oldValue

	return nil
}

// generateMessage generates Solidity message code from a protobuf message descriptor
func (g *Generator) generateMessage(descriptor *descriptorpb.DescriptorProto, packageName string, b *WriteableBuffer) error {
	structName := sanitizeKeyword(descriptor.GetName())
	
	// PostFiat enhancement: Handle maps and warn about other nested types
	if len(descriptor.GetEnumType()) > 0 {
		log.Printf("WARNING: Nested enums are not supported in protobuf3-solidity. " +
			"Message '%s' contains nested enums that will be ignored. " +
			"Consider flattening your protobuf structure. See BACKLOG.md for planned future support.", structName)
		return errors.New("nested enum definitions are forbidden: " + structName)
	}
	
	// Handle nested messages (which includes maps)
	if len(descriptor.GetNestedType()) > 0 {
		// Check if these are map entries (protobuf maps are represented as nested messages)
		hasNonMapNested := false
		for _, nestedType := range descriptor.GetNestedType() {
			if !nestedType.GetOptions().GetMapEntry() {
				hasNonMapNested = true
				break
			}
		}
		
		if hasNonMapNested {
			log.Printf("WARNING: Nested messages are not supported in protobuf3-solidity. " +
				"Message '%s' contains nested message that will be ignored. " +
				"Consider flattening your protobuf structure. See BACKLOG.md for planned future support.", structName)
			return errors.New("nested message definitions are forbidden: " + structName)
		}
	}

	fields := descriptor.GetField()

	// Debug: Print detailed field information for problematic messages
	if structName == "GetAgentCardRequest" {
		log.Printf("DEBUG: GetAgentCardRequest descriptor details:")
		log.Printf("  - Name: %s", descriptor.GetName())
		log.Printf("  - Field count: %d", len(fields))
		for i, field := range fields {
			log.Printf("  - Field %d: %s (type: %v, number: %d)", i, field.GetName(), field.GetType(), field.GetNumber())
		}
		log.Printf("  - Nested type count: %d", len(descriptor.GetNestedType()))
		log.Printf("  - Enum type count: %d", len(descriptor.GetEnumType()))
	}

	// Note: we don't need this check since it's enforced by protoc, but keep it just in case
	// However, some protoc versions may have bugs where valid messages show 0 fields in descriptors
	if len(fields) == 0 {
		// Check if this might be a protoc descriptor generation issue
		// If the message has a valid name and no nested types/enums, it's likely a valid message
		if len(descriptor.GetNestedType()) == 0 && len(descriptor.GetEnumType()) == 0 {
			log.Printf("WARNING: Message '%s' has 0 fields but appears to be a valid protobuf message. "+
				"This may be due to a protoc descriptor generation issue. Skipping field validation.", structName)
			// Continue processing without fields - this will generate an empty struct
		} else {
			return errors.New("messages must have at least one field: " + structName)
		}
	}

	// Check field numbers start at 1 and increment by 1
	oldFieldNumber := 0
	for _, field := range fields {
		fieldNumber := int(field.GetNumber())

		// Configurable field number validation - can be disabled via parameters
		if g.strictFieldNumberValidation && fieldNumber != oldFieldNumber+1 {
			return errors.New("field number does not increment by 1: " + structName + "." + field.GetName())
		}
		oldFieldNumber = fieldNumber
	}

	// Generate struct
	b.P(fmt.Sprintf("struct %s {", structName))
	b.Indent()

	// Generate fields (only if we have fields)
	if len(fields) > 0 {
		for _, field := range fields {
			fieldName := sanitizeKeyword(field.GetName())
			fieldDescriptorType := field.GetType()

			// Determine if field is repeated
			arrayStr := ""
			if isFieldRepeated(field) {
				arrayStr = "[]"
			}

			switch fieldDescriptorType {
			case descriptorpb.FieldDescriptorProto_TYPE_ENUM,
				descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
				// PostFiat enhancement: Check if this is a map field
				if g.isMapField(field, descriptor) {
					// Handle map field with wrapper message
					keyType, valueType, err := g.getMapKeyValueTypes(field, descriptor)
					if err != nil {
						return err
					}
					
					wrapperName := fmt.Sprintf("%sEntry", strings.Title(fieldName))
					if g.helperMessages[packageName] == nil {
						g.helperMessages[packageName] = make(map[string]*descriptorpb.DescriptorProto)
					}
					if _, exists := g.helperMessages[packageName][wrapperName]; !exists {
						g.helperMessages[packageName][wrapperName] = g.createMapWrapperMessage(fieldName, keyType, valueType)
						log.Printf("INFO: Generated wrapper message '%s' for map field '%s.%s'", wrapperName, structName, fieldName)
					}
					
					// Store the mapping from original type name to wrapper name
					originalTypeName := field.GetTypeName()
					if len(originalTypeName) > 0 && originalTypeName[0] == '.' {
						originalTypeName = originalTypeName[1:]
					}
					g.mapFieldMappings[originalTypeName] = wrapperName
					
					b.P(fmt.Sprintf("%s%s %s;", wrapperName, arrayStr, fieldName))
				} else {
					// Regular enum or message field
					fieldTypeName, err := g.getSolTypeName(field)
					if err != nil {
						return err
					}
					b.P(fmt.Sprintf("%s%s %s;", fieldTypeName, arrayStr, fieldName))
				}
			case descriptorpb.FieldDescriptorProto_TYPE_STRING:
				// PostFiat enhancement: Use wrapper message for repeated strings
				if isFieldRepeated(field) {
					wrapperName := fmt.Sprintf("%sList", strings.Title(fieldName))
					if g.helperMessages[packageName] == nil {
						g.helperMessages[packageName] = make(map[string]*descriptorpb.DescriptorProto)
					}
					if _, exists := g.helperMessages[packageName][wrapperName]; !exists {
						g.helperMessages[packageName][wrapperName] = g.createStringWrapperMessage(fieldName)
						log.Printf("INFO: Generated wrapper message '%s' for repeated string field '%s.%s'", wrapperName, structName, fieldName)
					}
					b.P(fmt.Sprintf("%s%s %s;", wrapperName, arrayStr, fieldName))
				} else {
					// Regular string field
					fieldType, err := typeToSol(fieldDescriptorType)
					if err != nil {
						return errors.New(err.Error() + ": " + structName + "." + fieldName)
					}
					b.P(fmt.Sprintf("%s %s;", fieldType, fieldName))
				}
			default:
				// Convert protobuf field type to Solidity native type
				fieldType, err := typeToSol(fieldDescriptorType)
				if err != nil {
					return errors.New(err.Error() + ": " + structName + "." + fieldName)
				}

				b.P(fmt.Sprintf("%s%s %s;", fieldType, arrayStr, fieldName))
			}
		}
	}

	b.Unindent()
	b.P("}")
	b.P()

	b.P(fmt.Sprintf("library %sCodec {", structName))
	b.Indent()

	// Only generate codec functions if we have fields
	if len(fields) > 0 {
		if g.generateFlag == generateFlagAll || g.generateFlag == generateFlagDecoder {
			var err error
			err = g.generateMessageDecoder(structName, fields, b)
			if err != nil {
				return err
			}
		}

		if g.generateFlag == generateFlagAll || g.generateFlag == generateFlagEncoder {
			var err error
			err = g.generateMessageEncoder(structName, fields, b)
			if err != nil {
				return err
			}
		}
	} else {
		// For messages with 0 fields, generate a minimal codec with empty functions
		b.P("// Empty message - no fields to encode/decode")
		b.P("function encode(bytes memory) internal pure returns (bytes memory) {")
		b.Indent()
		b.P("return new bytes(0);")
		b.Unindent()
		b.P("}")
		b.P("")
		b.P("function decode(bytes memory) internal pure returns (bytes memory) {")
		b.Indent()
		b.P("return new bytes(0);")
		b.Unindent()
		b.P("}")
	}

	b.Unindent()
	b.P("}")
	b.P()

	return nil
} 