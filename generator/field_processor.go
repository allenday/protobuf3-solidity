package generator

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/types/descriptorpb"
)

// FieldProcessor handles field name processing and validation
type FieldProcessor struct{}

// NewFieldProcessor creates a new field processor
func NewFieldProcessor() *FieldProcessor {
	return &FieldProcessor{}
}

// FieldInfo represents information about a field
type FieldInfo struct {
	originalName  string
	sanitizedName string
	fieldNumber   int32
}

// ProcessFieldNames processes and validates field names, ensuring uniqueness
func (fp *FieldProcessor) ProcessFieldNames(fields []*descriptorpb.FieldDescriptorProto) (map[int32]string, error) {
	// Create a map to track field names and ensure uniqueness
	fieldNameMap := make(map[int32]string) // field number -> sanitized name

	// Create a map to track used field names and ensure uniqueness
	usedFieldNames := make(map[string]bool)

	// First pass: collect all field names and their sanitized versions
	var allFields []FieldInfo

	for _, field := range fields {
		originalName := field.GetName()
		sanitizedName := sanitizeKeyword(originalName)
		fieldNumber := field.GetNumber()

		// For already sanitized names (starting with _), keep the original prefix
		if strings.HasPrefix(originalName, "_") {
			sanitizedName = originalName
		}

		allFields = append(allFields, FieldInfo{
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

	return fieldNameMap, nil
}

// GetArrayString returns the array suffix for a field if it's repeated
func (fp *FieldProcessor) GetArrayString(field *descriptorpb.FieldDescriptorProto) string {
	if isFieldRepeated(field) {
		return "[]"
	}
	return ""
}
