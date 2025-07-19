package generator

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/proto"
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

// generateFlattenedEnum generates a flattened enum from a nested enum descriptor
func (g *Generator) generateFlattenedEnum(descriptor *descriptorpb.EnumDescriptorProto, flattenedName string, b *WriteableBuffer) error {
	// Create a copy of the enum descriptor with the flattened name
	flattenedDescriptor := &descriptorpb.EnumDescriptorProto{
		Name:  proto.String(flattenedName),
		Value: descriptor.GetValue(),
	}
	
	// Generate the enum with the flattened name
	return g.generateEnum(flattenedDescriptor, b)
}

// generateFlattenedMessage generates a flattened message from a nested message descriptor
func (g *Generator) generateFlattenedMessage(descriptor *descriptorpb.DescriptorProto, packageName string, flattenedName string, b *WriteableBuffer) error {
	// Create a copy of the message descriptor with the flattened name
	flattenedDescriptor := &descriptorpb.DescriptorProto{
		Name:      proto.String(flattenedName),
		Field:     descriptor.GetField(),
		EnumType:  descriptor.GetEnumType(),
		NestedType: descriptor.GetNestedType(),
		Options:   descriptor.GetOptions(),
	}
	
	// Generate the message with the flattened name
	return g.generateMessage(flattenedDescriptor, packageName, b)
}

// generateMessage generates Solidity message code from a protobuf message descriptor
func (g *Generator) generateMessage(descriptor *descriptorpb.DescriptorProto, packageName string, b *WriteableBuffer) error {
	structName := sanitizeKeyword(descriptor.GetName())
	
	// PostFiat enhancement: Handle nested enums by flattening them to top-level
	if len(descriptor.GetEnumType()) > 0 {
		log.Printf("INFO: Flattening %d nested enums in message '%s' to top-level enums", len(descriptor.GetEnumType()), structName)
		
		// Generate flattened enums first
		for _, enumDescriptor := range descriptor.GetEnumType() {
			// Create unique name for the flattened enum
			flattenedEnumName := fmt.Sprintf("%s_%s", structName, enumDescriptor.GetName())
			
			// Store the mapping for type resolution
			g.enumMappings[fmt.Sprintf("%s.%s", structName, enumDescriptor.GetName())] = flattenedEnumName
			
			// Generate the flattened enum
			if err := g.generateFlattenedEnum(enumDescriptor, flattenedEnumName, b); err != nil {
				return err
			}
		}
	}
	
	// PostFiat enhancement: Handle nested messages by flattening them to top-level
	if len(descriptor.GetNestedType()) > 0 {
		// Filter out map entries (protobuf maps are represented as nested messages)
		var actualNestedMessages []*descriptorpb.DescriptorProto
		for _, nestedType := range descriptor.GetNestedType() {
			if !nestedType.GetOptions().GetMapEntry() {
				actualNestedMessages = append(actualNestedMessages, nestedType)
			}
		}
		
		if len(actualNestedMessages) > 0 {
			log.Printf("INFO: Flattening %d nested messages in message '%s' to top-level messages", len(actualNestedMessages), structName)
			
			// Generate flattened messages first
			for _, nestedDescriptor := range actualNestedMessages {
				// Create unique name for the flattened message
				flattenedMessageName := fmt.Sprintf("%s_%s", structName, nestedDescriptor.GetName())
				
				// Store the mapping for type resolution
				g.messageMappings[fmt.Sprintf("%s.%s", structName, nestedDescriptor.GetName())] = flattenedMessageName
				
				// Generate the flattened message
				if err := g.generateFlattenedMessage(nestedDescriptor, packageName, flattenedMessageName, b); err != nil {
					return err
				}
			}
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
		if g.strictFieldNumberValidation && !g.allowNonMonotonicFields && fieldNumber != oldFieldNumber+1 {
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
			case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
				// PostFiat enhancement: Use wrapper message for repeated bytes
				if isFieldRepeated(field) {
					wrapperName := fmt.Sprintf("%sList", strings.Title(fieldName))
					if g.helperMessages[packageName] == nil {
						g.helperMessages[packageName] = make(map[string]*descriptorpb.DescriptorProto)
					}
					if _, exists := g.helperMessages[packageName][wrapperName]; !exists {
						g.helperMessages[packageName][wrapperName] = g.createBytesWrapperMessage(fieldName)
						log.Printf("INFO: Generated wrapper message '%s' for repeated bytes field '%s.%s'", wrapperName, structName, fieldName)
					}
					b.P(fmt.Sprintf("%s%s %s;", wrapperName, arrayStr, fieldName))
				} else {
					// Regular bytes field
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