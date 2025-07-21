package generator

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"google.golang.org/protobuf/proto"
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
	b.P0()

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
	err := g.generateEnum(flattenedDescriptor, b)
	if err != nil {
		return err
	}

	b.P0()
	return nil
}

// generateFlattenedMessage generates a flattened message from a nested message descriptor
func (g *Generator) generateFlattenedMessage(descriptor *descriptorpb.DescriptorProto, packageName string, flattenedName string, b *WriteableBuffer) error {
	// Create a copy of the message descriptor with the flattened name
	flattenedDescriptor := &descriptorpb.DescriptorProto{
		Name:       proto.String(flattenedName),
		Field:      descriptor.GetField(),
		EnumType:   descriptor.GetEnumType(),
		NestedType: descriptor.GetNestedType(),
		Options:    descriptor.GetOptions(),
	}

	// Generate the message with the flattened name
	return g.generateMessage(flattenedDescriptor, packageName, b)
}

// generateMessage generates Solidity code for a protobuf message.
func (g *Generator) generateMessage(descriptor *descriptorpb.DescriptorProto, packageName string, b *WriteableBuffer) error {
	structName := sanitizeKeyword(descriptor.GetName())
	fields := descriptor.GetField()

	// Create a map to track field names and ensure uniqueness
	fieldNameMap := make(map[int32]string) // field number -> sanitized name

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

	// Generate struct
	b.P(fmt.Sprintf("// %s represents a protobuf message", structName))
	b.P(fmt.Sprintf("struct %s {", structName))
	b.Indent()

	// Generate fields (only if we have fields)
	if len(fields) > 0 {
		// Create a map to track used field names and ensure uniqueness
		usedFieldNames := make(map[string]bool)

		// First pass: collect all field names and their sanitized versions
		type fieldInfo struct {
			originalName  string
			sanitizedName string
			fieldNumber   int32
		}
		var allFields []fieldInfo

		for _, field := range fields {
			originalName := field.GetName()
			sanitizedName := sanitizeKeyword(originalName)
			fieldNumber := field.GetNumber()

			// For already sanitized names (starting with _), keep the original prefix
			if strings.HasPrefix(originalName, "_") {
				sanitizedName = originalName
			}

			allFields = append(allFields, fieldInfo{
				originalName:  originalName,
				sanitizedName: sanitizedName,
				fieldNumber:   fieldNumber,
			})
		}

		// Second pass: ensure unique field names
		for i := range allFields {
			field := &allFields[i]
			baseName := field.sanitizedName
			counter := 1

			// Keep trying names until we find a unique one
			for {
				used := false

				// Check if this name is already used
				for j := 0; j < i; j++ {
					if allFields[j].sanitizedName == field.sanitizedName {
						used = true
						break
					}
				}

				if !used {
					break
				}

				// Try next name variant
				if strings.HasPrefix(baseName, "_") {
					// For already sanitized names, append counter after the original name
					field.sanitizedName = fmt.Sprintf("%s_%d", baseName, counter)
				} else {
					// For regular names, prepend underscore and append counter
					field.sanitizedName = fmt.Sprintf("_%s_%d", baseName, counter)
				}
				counter++
			}

			usedFieldNames[field.sanitizedName] = true
			fieldNameMap[field.fieldNumber] = field.sanitizedName
		}

		// Generate fields with unique names
		for _, field := range fields {
			fieldName := fieldNameMap[field.GetNumber()]
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
					g.messageMappings[originalTypeName] = wrapperName

					// Use the wrapper message type for the map field
					b.P(fmt.Sprintf("%s%s %s;", wrapperName, arrayStr, fieldName))
				} else {
					// Handle regular enum or message field
					typeName, err := g.getSolTypeName(field)
					if err != nil {
						return err
					}
					b.P(fmt.Sprintf("%s%s %s;", typeName, arrayStr, fieldName))
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
					b.P(fmt.Sprintf("%s%s %s;", fieldType, arrayStr, fieldName))
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
					b.P(fmt.Sprintf("%s%s %s;", fieldType, arrayStr, fieldName))
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
	b.P0()

	// Generate codec library at the top level
	b.P(fmt.Sprintf("library %sCodec {", structName))
	b.Indent()

	// Only generate codec functions if we have fields
	if len(fields) > 0 {
		if g.generateFlag == generateFlagAll || g.generateFlag == generateFlagDecoder {
			err := g.generateMessageDecoder(structName, fields, fieldNameMap, b)
			if err != nil {
				return err
			}
		}

		if g.generateFlag == generateFlagAll || g.generateFlag == generateFlagEncoder {
			err := g.generateMessageEncoder(structName, fields, fieldNameMap, b)
			if err != nil {
				return err
			}
		}
	}

	b.Unindent()
	b.P("}")
	b.P0()

	return nil
}

// generateMessageStruct generates only the struct definition for a protobuf message (no codec library)
func (g *Generator) generateMessageStruct(descriptor *descriptorpb.DescriptorProto, packageName string, b *WriteableBuffer) error {
	structName := sanitizeKeyword(descriptor.GetName())
	fields := descriptor.GetField()

	// Use the field processor to handle field name processing
	fieldProcessor := NewFieldProcessor()
	fieldNameMap, err := fieldProcessor.ProcessFieldNames(fields)
	if err != nil {
		return err
	}

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

	// Generate struct
	b.P(fmt.Sprintf("// %s represents a protobuf message", structName))
	b.P(fmt.Sprintf("struct %s {", structName))
	b.Indent()

	// Generate fields (only if we have fields)
	if len(fields) > 0 {
		// Generate field definitions
		for _, field := range fields {
			fieldName := fieldNameMap[field.GetNumber()]
			fieldDescriptorType := field.GetType()

			// Get array suffix for repeated fields
			arrayStr := fieldProcessor.GetArrayString(field)

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
					g.messageMappings[originalTypeName] = wrapperName

					// Use the wrapper message type for the map field
					b.P(fmt.Sprintf("%s%s %s;", wrapperName, arrayStr, fieldName))
				} else {
					// Handle regular enum or message field
					typeName, err := g.getSolTypeName(field)
					if err != nil {
						return err
					}
					b.P(fmt.Sprintf("%s%s %s;", typeName, arrayStr, fieldName))
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
	b.P0()

	return nil
}

// generateMessageCodec generates only the codec library for a protobuf message (outside main library)
func (g *Generator) generateMessageCodec(descriptor *descriptorpb.DescriptorProto, packageName string, b *WriteableBuffer) error {
	structName := sanitizeKeyword(descriptor.GetName())
	fields := descriptor.GetField()

	// Use the field processor to handle field name processing
	fieldProcessor := NewFieldProcessor()
	fieldNameMap, err := fieldProcessor.ProcessFieldNames(fields)
	if err != nil {
		return err
	}

	// Generate codec library at the top level (outside main library)
	b.P(fmt.Sprintf("library %sCodec {", structName))
	b.Indent()

	// Only generate codec functions if we have fields
	if len(fields) > 0 {
		// Generate helper functions first
		codecHelperGen := NewCodecHelperGenerator()
		// Create qualified struct name for codec functions
		qualifiedStructName := PackageToLibraryName(packageName) + "." + structName
		err := codecHelperGen.GenerateCodecHelpers(qualifiedStructName, fields, fieldNameMap, b)
		if err != nil {
			return err
		}

		if g.generateFlag == generateFlagAll || g.generateFlag == generateFlagDecoder {
			err := g.generateMessageDecoder(qualifiedStructName, fields, fieldNameMap, b)
			if err != nil {
				return err
			}
		}

		if g.generateFlag == generateFlagAll || g.generateFlag == generateFlagEncoder {
			err := g.generateMessageEncoder(qualifiedStructName, fields, fieldNameMap, b)
			if err != nil {
				return err
			}
		}
	}

	b.Unindent()
	b.P("}")
	b.P0()

	return nil
}
