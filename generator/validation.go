package generator

import (
	"fmt"
	"sort"

	"google.golang.org/protobuf/types/descriptorpb"
)

// checkSyntaxVersion checks that the syntax version is proto3
func checkSyntaxVersion(syntax string) error {
	if syntax != "proto3" {
		return fmt.Errorf("only proto3 is supported, got %s", syntax)
	}
	return nil
}

// checkFieldNumbers validates that field numbers follow the required rules
func checkFieldNumbers(fields []*descriptorpb.FieldDescriptorProto, strictFieldNumbers bool) error {
	if !strictFieldNumbers {
		return nil
	}

	if len(fields) == 0 {
		return nil
	}

	// Sort fields by number to check for gaps
	fieldNumbers := make([]int32, len(fields))
	for i, field := range fields {
		fieldNumbers[i] = field.GetNumber()
	}
	sort.Slice(fieldNumbers, func(i, j int) bool {
		return fieldNumbers[i] < fieldNumbers[j]
	})

	// Check that field numbers start at 1 and increment by 1
	if fieldNumbers[0] != 1 {
		return fmt.Errorf("field numbers must start at 1, got %d", fieldNumbers[0])
	}

	for i := 1; i < len(fieldNumbers); i++ {
		if fieldNumbers[i] != fieldNumbers[i-1]+1 {
			return fmt.Errorf("field numbers must increment by 1, got %d after %d", fieldNumbers[i], fieldNumbers[i-1])
		}
	}

	return nil
}

// checkRepeatedNumericFields validates that repeated numeric fields are packed
func checkRepeatedNumericFields(fields []*descriptorpb.FieldDescriptorProto) error {
	for _, field := range fields {
		if field.Label == nil || *field.Label != descriptorpb.FieldDescriptorProto_LABEL_REPEATED {
			continue
		}

		// Check if this is a numeric field
		switch field.GetType() {
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
			descriptorpb.FieldDescriptorProto_TYPE_DOUBLE,
			descriptorpb.FieldDescriptorProto_TYPE_BOOL:
			if !field.GetOptions().GetPacked() {
				return fmt.Errorf("repeated numeric field '%s' must be packed", field.GetName())
			}
		}
	}

	return nil
}
