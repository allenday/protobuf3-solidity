package generator

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"google.golang.org/protobuf/types/descriptorpb"
)

// typeToSol converts protobuf field type to Solidity native type
func typeToSol(fType descriptorpb.FieldDescriptorProto_Type) (string, error) {
	switch fType {
	case descriptorpb.FieldDescriptorProto_TYPE_DOUBLE:
		// Convert double to int64 with fixed-point scaling (1e15 precision)
		return "int64", nil
	case descriptorpb.FieldDescriptorProto_TYPE_FLOAT:
		// Convert float to int32 with fixed-point scaling (1e6 precision)
		return "int32", nil
	case descriptorpb.FieldDescriptorProto_TYPE_INT64:
		return "int64", nil
	case descriptorpb.FieldDescriptorProto_TYPE_UINT64:
		return "uint64", nil
	case descriptorpb.FieldDescriptorProto_TYPE_INT32:
		return "int32", nil
	case descriptorpb.FieldDescriptorProto_TYPE_FIXED64:
		return "uint64", nil
	case descriptorpb.FieldDescriptorProto_TYPE_FIXED32:
		return "uint32", nil
	case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
		return "bool", nil
	case descriptorpb.FieldDescriptorProto_TYPE_STRING:
		return "string", nil
	case descriptorpb.FieldDescriptorProto_TYPE_GROUP:
		return "", errors.New("unsupported field type TYPE_GROUP")
	case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
		return "", errors.New("unsupported field type TYPE_MESSAGE")
	case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
		return "bytes", nil
	case descriptorpb.FieldDescriptorProto_TYPE_UINT32:
		return "uint32", nil
	case descriptorpb.FieldDescriptorProto_TYPE_ENUM:
		return "", errors.New("unsupported field type TYPE_ENUM")
	case descriptorpb.FieldDescriptorProto_TYPE_SFIXED32:
		return "int32", nil
	case descriptorpb.FieldDescriptorProto_TYPE_SFIXED64:
		return "int64", nil
	case descriptorpb.FieldDescriptorProto_TYPE_SINT32:
		return "int32", nil
	case descriptorpb.FieldDescriptorProto_TYPE_SINT64:
		return "int64", nil
	default:
		return "", errors.New("unsupported field type: " + fType.String())
	}
}

// typeToDecodeSol converts protobuf field type to Solidity decode function name
func typeToDecodeSol(fType descriptorpb.FieldDescriptorProto_Type) (string, error) {
	switch fType {
	case descriptorpb.FieldDescriptorProto_TYPE_DOUBLE:
		// Use custom decode function for double with scaling
		return "double_scaled", nil
	case descriptorpb.FieldDescriptorProto_TYPE_FLOAT:
		// Use custom decode function for float with scaling
		return "float_scaled", nil
	case descriptorpb.FieldDescriptorProto_TYPE_INT64:
		return "int64", nil
	case descriptorpb.FieldDescriptorProto_TYPE_UINT64:
		return "uint64", nil
	case descriptorpb.FieldDescriptorProto_TYPE_INT32:
		return "int32", nil
	case descriptorpb.FieldDescriptorProto_TYPE_FIXED64:
		return "uint64", nil
	case descriptorpb.FieldDescriptorProto_TYPE_FIXED32:
		return "uint32", nil
	case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
		return "bool", nil
	case descriptorpb.FieldDescriptorProto_TYPE_STRING:
		return "string", nil
	case descriptorpb.FieldDescriptorProto_TYPE_GROUP:
		return "", errors.New("unsupported field type TYPE_GROUP")
	case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
		return "", errors.New("unsupported field type TYPE_MESSAGE")
	case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
		return "bytes", nil
	case descriptorpb.FieldDescriptorProto_TYPE_UINT32:
		return "uint32", nil
	case descriptorpb.FieldDescriptorProto_TYPE_ENUM:
		return "", errors.New("unsupported field type TYPE_ENUM")
	case descriptorpb.FieldDescriptorProto_TYPE_SFIXED32:
		return "int32", nil
	case descriptorpb.FieldDescriptorProto_TYPE_SFIXED64:
		return "int64", nil
	case descriptorpb.FieldDescriptorProto_TYPE_SINT32:
		return "int32", nil
	case descriptorpb.FieldDescriptorProto_TYPE_SINT64:
		return "int64", nil
	default:
		return "", errors.New("unsupported field type: " + fType.String())
	}
}

// typeToEncodeSol converts protobuf field type to Solidity encode function name
func typeToEncodeSol(fType descriptorpb.FieldDescriptorProto_Type) (string, error) {
	switch fType {
	case descriptorpb.FieldDescriptorProto_TYPE_DOUBLE:
		// Use custom encode function for double with scaling
		return "encode_double_scaled", nil
	case descriptorpb.FieldDescriptorProto_TYPE_FLOAT:
		// Use custom encode function for float with scaling
		return "encode_float_scaled", nil
	case descriptorpb.FieldDescriptorProto_TYPE_INT64:
		return "ProtobufLib.encode_int64", nil
	case descriptorpb.FieldDescriptorProto_TYPE_UINT64:
		return "ProtobufLib.encode_uint64", nil
	case descriptorpb.FieldDescriptorProto_TYPE_INT32:
		return "ProtobufLib.encode_int32", nil
	case descriptorpb.FieldDescriptorProto_TYPE_FIXED64:
		return "ProtobufLib.encode_fixed64", nil
	case descriptorpb.FieldDescriptorProto_TYPE_FIXED32:
		return "ProtobufLib.encode_fixed32", nil
	case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
		return "ProtobufLib.encode_bool", nil
	case descriptorpb.FieldDescriptorProto_TYPE_STRING:
		return "ProtobufLib.encode_string", nil
	case descriptorpb.FieldDescriptorProto_TYPE_GROUP:
		return "", errors.New("unsupported field type TYPE_GROUP")
	case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
		return "", errors.New("unsupported field type TYPE_MESSAGE")
	case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
		return "ProtobufLib.encode_bytes", nil
	case descriptorpb.FieldDescriptorProto_TYPE_UINT32:
		return "ProtobufLib.encode_uint32", nil
	case descriptorpb.FieldDescriptorProto_TYPE_ENUM:
		return "ProtobufLib.encode_enum", nil
	case descriptorpb.FieldDescriptorProto_TYPE_SFIXED32:
		return "ProtobufLib.encode_sfixed32", nil
	case descriptorpb.FieldDescriptorProto_TYPE_SFIXED64:
		return "ProtobufLib.encode_sfixed64", nil
	case descriptorpb.FieldDescriptorProto_TYPE_SINT32:
		return "ProtobufLib.encode_sint32", nil
	case descriptorpb.FieldDescriptorProto_TYPE_SINT64:
		return "ProtobufLib.encode_sint64", nil
	default:
		return "", errors.New("unsupported field type: " + fType.String())
	}
}

// isPrimitiveNumericType checks if a field type is a primitive numeric type
func isPrimitiveNumericType(fType descriptorpb.FieldDescriptorProto_Type) bool {
	switch fType {
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
		return true
	default:
		return false
	}
}

// isFieldRepeated checks if a field is repeated
func isFieldRepeated(field *descriptorpb.FieldDescriptorProto) bool {
	return field.GetLabel() == descriptorpb.FieldDescriptorProto_LABEL_REPEATED
}

// isFieldPacked checks if a field is packed
func isFieldPacked(field *descriptorpb.FieldDescriptorProto) bool {
	return field.GetOptions().GetPacked()
}

// toSolWireType converts protobuf field type to Solidity wire type
func toSolWireType(field *descriptorpb.FieldDescriptorProto) (string, error) {
	fieldType := field.GetType()

	switch fieldType {
	case descriptorpb.FieldDescriptorProto_TYPE_ENUM,
		descriptorpb.FieldDescriptorProto_TYPE_INT32,
		descriptorpb.FieldDescriptorProto_TYPE_INT64,
		descriptorpb.FieldDescriptorProto_TYPE_UINT32,
		descriptorpb.FieldDescriptorProto_TYPE_UINT64,
		descriptorpb.FieldDescriptorProto_TYPE_SINT32,
		descriptorpb.FieldDescriptorProto_TYPE_SINT64,
		descriptorpb.FieldDescriptorProto_TYPE_BOOL:
		return "ProtobufLib.WireType.Varint", nil
	case descriptorpb.FieldDescriptorProto_TYPE_FIXED64,
		descriptorpb.FieldDescriptorProto_TYPE_SFIXED64,
		descriptorpb.FieldDescriptorProto_TYPE_DOUBLE:
		return "ProtobufLib.WireType.Fixed64", nil
	case descriptorpb.FieldDescriptorProto_TYPE_STRING,
		descriptorpb.FieldDescriptorProto_TYPE_BYTES,
		descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
		return "ProtobufLib.WireType.LengthDelimited", nil
	case descriptorpb.FieldDescriptorProto_TYPE_FIXED32,
		descriptorpb.FieldDescriptorProto_TYPE_SFIXED32,
		descriptorpb.FieldDescriptorProto_TYPE_FLOAT:
		return "ProtobufLib.WireType.Fixed32", nil
	default:
		return "", errors.New("unsupported field type: " + fieldType.String())
	}
}

// toSolMessageOrEnumName extracts the message or enum name from a field
func toSolMessageOrEnumName(field *descriptorpb.FieldDescriptorProto) (string, error) {
	// Names take the form ".name", so remove the leading period
	typeName := field.GetTypeName()
	log.Printf("DEBUG: toSolMessageOrEnumName called for field '%s' with typeName: '%s'", field.GetName(), typeName)
	
	if len(typeName) == 0 {
		log.Printf("WARNING: Empty type name for field '%s', using default type", field.GetName())
		// Workaround for corrupted descriptors: use a default type name
		return "UnknownType", nil
	}
	
	// Remove leading dot
	if typeName[0] == '.' {
		typeName = typeName[1:]
		log.Printf("DEBUG: Removed leading dot, typeName now: '%s'", typeName)
	}
	
	// Handle package-qualified type names
	// Format: "package.name.TypeName" -> "Package_Name.TypeName"
	if strings.Contains(typeName, ".") {
		parts := strings.Split(typeName, ".")
		if len(parts) >= 2 {
			// Convert package name to library name format
			packageParts := parts[:len(parts)-1]
			typeNamePart := parts[len(parts)-1]
			
			// Convert package parts to library name format
			for i, part := range packageParts {
				if len(part) > 0 {
					packageParts[i] = strings.ToUpper(part[:1]) + part[1:]
				}
			}
			libraryName := strings.Join(packageParts, "_")
			
			// Return library-qualified type name
			result := fmt.Sprintf("%s.%s", libraryName, typeNamePart)
			log.Printf("DEBUG: Package-qualified type resolved to: '%s'", result)
			return result, nil
		}
	}
	
	log.Printf("DEBUG: Simple type name resolved to: '%s'", typeName)
	return typeName, nil
} 