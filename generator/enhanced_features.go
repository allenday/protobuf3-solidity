package generator

import (
	"errors"
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"log"
)

// getSolTypeName gets the Solidity type name for a field, handling map field mappings
func (g *Generator) getSolTypeName(field *descriptorpb.FieldDescriptorProto) (string, error) {
	log.Printf("DEBUG: getSolTypeName called for field '%s' with type: %s", field.GetName(), field.GetType())
	
	originalTypeName, err := toSolMessageOrEnumName(field)
	if err != nil {
		log.Printf("ERROR: toSolMessageOrEnumName failed for field '%s': %v", field.GetName(), err)
		return "", err
	}
	
	log.Printf("DEBUG: getSolTypeName resolved '%s' to '%s'", field.GetName(), originalTypeName)
	
	// Check if this is a map field that has been mapped to a wrapper
	if wrapperName, exists := g.mapFieldMappings[originalTypeName]; exists {
		log.Printf("DEBUG: Found map field mapping: '%s' -> '%s'", originalTypeName, wrapperName)
		return wrapperName, nil
	}
	
	return originalTypeName, nil
}

// createStringWrapperMessage creates a wrapper message for repeated string fields
func (g *Generator) createStringWrapperMessage(fieldName string) *descriptorpb.DescriptorProto {
	wrapperName := fmt.Sprintf("%sList", strings.Title(fieldName))
	
	// Create a field for the string value
	stringField := &descriptorpb.FieldDescriptorProto{
		Name:   proto.String("value"),
		Number: proto.Int32(1),
		Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
		Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
	}
	
	return &descriptorpb.DescriptorProto{
		Name:  proto.String(wrapperName),
		Field: []*descriptorpb.FieldDescriptorProto{stringField},
	}
}

// createMapWrapperMessage creates a wrapper message for map fields
func (g *Generator) createMapWrapperMessage(fieldName string, keyType, valueType descriptorpb.FieldDescriptorProto_Type) *descriptorpb.DescriptorProto {
	wrapperName := fmt.Sprintf("%sEntry", strings.Title(fieldName))
	
	// Create key field
	keyField := &descriptorpb.FieldDescriptorProto{
		Name:   proto.String("key"),
		Number: proto.Int32(1),
		Type:   &keyType,
		Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
	}
	
	// Create value field
	valueField := &descriptorpb.FieldDescriptorProto{
		Name:   proto.String("value"),
		Number: proto.Int32(2),
		Type:   &valueType,
		Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
	}
	
	return &descriptorpb.DescriptorProto{
		Name:  proto.String(wrapperName),
		Field: []*descriptorpb.FieldDescriptorProto{keyField, valueField},
	}
}

// isMapField checks if a field is a map field by looking for the map entry message type
func (g *Generator) isMapField(field *descriptorpb.FieldDescriptorProto, parentDescriptor *descriptorpb.DescriptorProto) bool {
	// Must be a repeated message field
	if !isFieldRepeated(field) || field.GetType() != descriptorpb.FieldDescriptorProto_TYPE_MESSAGE {
		return false
	}
	
	// Get the type name (remove leading dot)
	typeName := field.GetTypeName()
	if len(typeName) > 0 && typeName[0] == '.' {
		typeName = typeName[1:]
	}
	
	// Check if any nested type matches this name and is a map entry
	for _, nestedType := range parentDescriptor.GetNestedType() {
		// Extract just the type name part (after the last dot)
		parts := strings.Split(typeName, ".")
		simpleTypeName := parts[len(parts)-1]
		
		if nestedType.GetName() == simpleTypeName && nestedType.GetOptions().GetMapEntry() {
			return true
		}
	}
	
	return false
}

// getMapKeyValueTypes extracts the key and value types from a map field
func (g *Generator) getMapKeyValueTypes(field *descriptorpb.FieldDescriptorProto, parentDescriptor *descriptorpb.DescriptorProto) (descriptorpb.FieldDescriptorProto_Type, descriptorpb.FieldDescriptorProto_Type, error) {
	// Get the type name (remove leading dot)
	typeName := field.GetTypeName()
	if len(typeName) > 0 && typeName[0] == '.' {
		typeName = typeName[1:]
	}
	
	// Find the map entry message
	for _, nestedType := range parentDescriptor.GetNestedType() {
		// Extract just the type name part (after the last dot)
		parts := strings.Split(typeName, ".")
		simpleTypeName := parts[len(parts)-1]
		
		if nestedType.GetName() == simpleTypeName && nestedType.GetOptions().GetMapEntry() {
			// Map entry messages have exactly 2 fields: key and value
			if len(nestedType.GetField()) != 2 {
				return 0, 0, errors.New("invalid map entry message: " + typeName)
			}
			
			var keyType, valueType descriptorpb.FieldDescriptorProto_Type
			for _, mapField := range nestedType.GetField() {
				if mapField.GetName() == "key" {
					keyType = mapField.GetType()
				} else if mapField.GetName() == "value" {
					valueType = mapField.GetType()
				}
			}
			
			return keyType, valueType, nil
		}
	}
	
	return 0, 0, errors.New("map entry message not found: " + typeName)
} 