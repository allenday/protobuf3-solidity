package generator

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

// SolidityVersionString is the Solidity version specifier.
const SolidityVersionString = ">=0.6.0 <8.0.0"

// SolidityABIString indicates ABIEncoderV2 use.
const SolidityABIString = "pragma experimental ABIEncoderV2;"

type compileFlag string

const (
	compileFlagLink    compileFlag = "link"
	compileFlagCompile compileFlag = "compile"
)

func fromCompileFlag(f compileFlag) string {
	return string(f)
}

func toCompileFlag(s string) (compileFlag, error) {
	switch s {
	case fromCompileFlag(compileFlagLink):
		return compileFlagLink, nil
	case fromCompileFlag(compileFlagCompile):
		return compileFlagCompile, nil
	}

	return compileFlagCompile, fmt.Errorf("unknown compile flag %s, allowed values are <link, compile>", s)
}

type generateFlag string

const (
	generateFlagAll     generateFlag = "all"
	generateFlagDecoder generateFlag = "decoder"
	generateFlagEncoder generateFlag = "encoder"
)

func fromGenerateFlag(f generateFlag) string {
	return string(f)
}

func toGenerateFlag(s string) (generateFlag, error) {
	switch s {
	case fromGenerateFlag(generateFlagAll):
		return generateFlagAll, nil
	case fromGenerateFlag(generateFlagDecoder):
		return generateFlagDecoder, nil
	case fromGenerateFlag(generateFlagEncoder):
		return generateFlagEncoder, nil
	}

	return generateFlagAll, fmt.Errorf("unknown generate flag %s, allowed values are <all, decoder, encoder>", s)
}

// Generator generates Solidity code from .proto files.
type Generator struct {
	request   *pluginpb.CodeGeneratorRequest
	enumMaxes map[string]int

	versionString string
	licenseString string
	compileFlag   compileFlag
	generateFlag  generateFlag
	
	// Enhanced features for PostFiat support
	helperMessages map[string]map[string]*descriptorpb.DescriptorProto // package -> message name -> descriptor (only wrapper messages)
	// Track map field type mappings: original type name -> wrapper name
	mapFieldMappings map[string]string
	// Track nested enum name mappings: original nested name -> flattened name
	enumMappings map[string]string
	// Track nested message name mappings: original nested name -> flattened name
	messageMappings map[string]string
	
	// Global message registry for type resolution
	messageRegistry map[string]*descriptorpb.DescriptorProto
	
	// Configuration options
	strictFieldNumberValidation bool
	strictEnumValidation       bool
	allowEmptyPackedArrays     bool
	allowNonMonotonicFields    bool
}

// New initializes a new Generator.
func New(request *pluginpb.CodeGeneratorRequest, versionString string) *Generator {
	g := new(Generator)

	g.request = request
	g.enumMaxes = make(map[string]int)
	g.helperMessages = make(map[string]map[string]*descriptorpb.DescriptorProto)
	g.mapFieldMappings = make(map[string]string)
	g.enumMappings = make(map[string]string)
	g.messageMappings = make(map[string]string)
	g.messageRegistry = make(map[string]*descriptorpb.DescriptorProto)

	g.versionString = versionString
	g.licenseString = "CC0"

	g.compileFlag = compileFlagCompile
	g.generateFlag = generateFlagDecoder
	
	// Default configuration
	g.strictFieldNumberValidation = true
	g.strictEnumValidation = true
	g.allowEmptyPackedArrays = false
	g.allowNonMonotonicFields = false

	return g
}

// ParseParameters parses command-line parameters
func (g *Generator) ParseParameters() error {
	parameterString := g.request.GetParameter()
	if len(parameterString) == 0 {
		return nil
	}

	for _, parameter := range strings.Split(parameterString, ",") {
		keyvalue := strings.Split(parameter, "=")
		key, value := keyvalue[0], keyvalue[1]

		switch key {
		case "license":
			g.licenseString = value
		case "compile":
			flag, err := toCompileFlag(value)
			if err != nil {
				return err
			}
			g.compileFlag = flag

			// TODO implement these
			switch flag {
			case compileFlagLink:
				return fmt.Errorf("unimplemented feature %s", flag)
			}
		case "generate":
			flag, err := toGenerateFlag(value)
			if err != nil {
				return err
			}
			g.generateFlag = flag
		case "strict_field_numbers":
			if value == "false" {
				g.strictFieldNumberValidation = false
			} else if value == "true" {
				g.strictFieldNumberValidation = true
			} else {
				return errors.New("strict_field_numbers must be 'true' or 'false'")
			}
		case "strict_enum_validation":
			if value == "false" {
				g.strictEnumValidation = false
			} else if value == "true" {
				g.strictEnumValidation = true
			} else {
				return errors.New("strict_enum_validation must be 'true' or 'false'")
			}
		case "allow_empty_packed_arrays":
			if value == "true" {
				g.allowEmptyPackedArrays = true
			} else if value == "false" {
				g.allowEmptyPackedArrays = false
			} else {
				return errors.New("allow_empty_packed_arrays must be 'true' or 'false'")
			}
		case "allow_non_monotonic_fields":
			if value == "true" {
				g.allowNonMonotonicFields = true
			} else if value == "false" {
				g.allowNonMonotonicFields = false
			} else {
				return errors.New("allow_non_monotonic_fields must be 'true' or 'false'")
			}
		default:
			return errors.New("unrecognized option " + key)
		}
	}

	return nil
}

// Generate generates Solidity code from the requested .proto files.
func (g *Generator) Generate() (*pluginpb.CodeGeneratorResponse, error) {
	response := &pluginpb.CodeGeneratorResponse{}

	protoFiles := g.request.GetProtoFile()
	fileToGenerateSet := make(map[string]struct{})
	for _, fname := range g.request.GetFileToGenerate() {
		fileToGenerateSet[fname] = struct{}{}
	}

	// Build a global registry of all messages for type resolution
	g.buildGlobalMessageRegistry(protoFiles)

	log.Printf("DEBUG: Processing %d proto files", len(protoFiles))
	for i, protoFile := range protoFiles {
		if _, ok := fileToGenerateSet[protoFile.GetName()]; !ok {
			log.Printf("DEBUG: Skipping file %d: %s (not in FileToGenerate)", i, protoFile.GetName())
			continue
		}
		log.Printf("DEBUG: File %d: %s (package: %s, messages: %d)", i, protoFile.GetName(), protoFile.GetPackage(), len(protoFile.GetMessageType()))
		
		// Clear helper messages for this package before processing
		packageName := protoFile.GetPackage()
		if packageMessages, exists := g.helperMessages[packageName]; exists {
			for wrapperName := range packageMessages {
				delete(packageMessages, wrapperName)
			}
			delete(g.helperMessages, packageName)
		}
		
		// Process the file
		responseFile, err := g.generateFile(protoFile)
		if err != nil {
			log.Printf("ERROR: Failed to process file %d (%s): %v", i, protoFile.GetName(), err)
			return nil, err
		}

		if responseFile != nil {
			log.Printf("DEBUG: Successfully generated file for %s", protoFile.GetName())
			response.File = append(response.File, responseFile)
		} else {
			log.Printf("DEBUG: Skipped file %s (no output generated)", protoFile.GetName())
		}
		
		// Clear helper messages after processing the file
		if packageMessages, exists := g.helperMessages[packageName]; exists {
			for wrapperName := range packageMessages {
				delete(packageMessages, wrapperName)
			}
			delete(g.helperMessages, packageName)
		}
	}

	return response, nil
}

// buildGlobalMessageRegistry builds a registry of all messages for type resolution
func (g *Generator) buildGlobalMessageRegistry(protoFiles []*descriptorpb.FileDescriptorProto) {
	if g.messageRegistry == nil {
		g.messageRegistry = make(map[string]*descriptorpb.DescriptorProto)
	}
	for _, protoFile := range protoFiles {
		pkg := protoFile.GetPackage()
		for _, msg := range protoFile.GetMessageType() {
			// Use fully qualified name for global registry
			fullName := msg.GetName()
			if len(pkg) > 0 {
				fullName = pkg + "." + fullName
			}
			g.messageRegistry[fullName] = msg
		}
	}
}

// generateFile generates Solidity code from a single .proto file.
func (g *Generator) generateFile(protoFile *descriptorpb.FileDescriptorProto) (*pluginpb.CodeGeneratorResponse_File, error) {
	// Skip Google protobuf standard library files and Google API files
	// (they use proto2 or have complex nested structures)
	fileName := protoFile.GetName()
	if strings.HasPrefix(fileName, "google/protobuf/") || strings.HasPrefix(fileName, "google/api/") {
		// Skip these files as they are part of the Google standard library
		// and may use proto2 syntax or have complex nested structures
		return nil, nil
	}
	
	// Only support proto3
	syntax := protoFile.GetSyntax()
	if len(syntax) == 0 {
		return nil, fmt.Errorf("file %s has no syntax declaration", fileName)
	}
	
	err := checkSyntaxVersion(syntax)
	if err != nil {
		return nil, fmt.Errorf("file %s: %v", fileName, err)
	}

	// Buffer to hold the generate file's text
	b := &WriteableBuffer{}

	// Generate heading
	b.P(fmt.Sprintf("// File automatically generated by protoc-gen-sol %s", g.versionString))
	b.P(fmt.Sprintf("// SPDX-License-Identifier: %s", g.licenseString))
	b.P("pragma solidity " + SolidityVersionString + ";")
	b.P(SolidityABIString)
	b.P0()

	// Handle package namespace
	packageName := protoFile.GetPackage()
	if len(packageName) > 0 {
		// Generate package namespace
		b.P(fmt.Sprintf("// Package: %s", packageName))
		b.P(fmt.Sprintf("library %s {", g.packageToLibraryName(packageName)))
		b.P0()
		b.Indent()
	}

	// Generate imports
	b.P("import \"@lazyledger/protobuf3-solidity-lib/contracts/ProtobufLib.sol\";")
	
	// Generate imports for dependencies
	for _, dependency := range protoFile.GetDependency() {
		// Convert dependency path to import path
		importPath := g.dependencyToImportPath(dependency)
		b.P(fmt.Sprintf("import \"%s\";", importPath))
	}
	b.P0()

	// Generate enums
	for _, descriptor := range protoFile.GetEnumType() {
		err := g.generateEnum(descriptor, b)
		if err != nil {
			return nil, err
		}
	}

	// Generate regular messages
	for _, descriptor := range protoFile.GetMessageType() {
		err := g.generateMessage(descriptor, packageName, b)
		if err != nil {
			return nil, err
		}
	}
	
	// Generate wrapper messages (Entry/List) if any exist
	if packageMessages, exists := g.helperMessages[packageName]; exists && len(packageMessages) > 0 {
		// Collect only wrapper messages
		wrapperMessages := make([]*descriptorpb.DescriptorProto, 0)
		for wrapperName, wrapperDescriptor := range packageMessages {
			if strings.HasSuffix(wrapperName, "Entry") || strings.HasSuffix(wrapperName, "List") {
				wrapperMessages = append(wrapperMessages, wrapperDescriptor)
			}
		}
		
		// Only generate helper messages section if we have wrapper messages
		if len(wrapperMessages) > 0 {
			b.P("// Helper messages for PostFiat enhancements")
			b.P0()
			
			for _, wrapperDescriptor := range wrapperMessages {
				err := g.generateMessage(wrapperDescriptor, packageName, b)
				if err != nil {
					return nil, err
				}
			}
		}
		
		// Clear helper messages after generation to prevent duplicates
		for wrapperName := range packageMessages {
			delete(packageMessages, wrapperName)
		}
		delete(g.helperMessages, packageName)
	}

	// Generate float/double scaling helper functions
	g.generateFloatDoubleHelpers(b)

	// Close package namespace
	if len(packageName) > 0 {
		b.Unindent()
		b.P("}")
		b.P0()
	}

	// Create response file
	responseFile := &pluginpb.CodeGeneratorResponse_File{}
	responseFile.Name = proto.String(strings.TrimSuffix(fileName, ".proto") + ".sol")
	responseFile.Content = proto.String(b.String())

	return responseFile, nil
}

// generateService generates Solidity interface code from a protobuf service descriptor
func (g *Generator) generateService(service *descriptorpb.ServiceDescriptorProto, b *WriteableBuffer) error {
	serviceName := sanitizeKeyword(service.GetName())

	b.P(fmt.Sprintf("interface %s {", serviceName))
	b.Indent()

	for _, method := range service.GetMethod() {
		methodName := method.GetName()
		inputType := method.GetInputType()
		outputType := method.GetOutputType()

		// Handle package-qualified type names
		inputTypeName, err := g.resolveTypeName(inputType)
		if err != nil {
			return err
		}
		outputTypeName, err := g.resolveTypeName(outputType)
		if err != nil {
			return err
		}

		// Generate method signature
		b.P(fmt.Sprintf("function %s(%s memory request) external pure returns (%s memory);", 
			methodName, inputTypeName, outputTypeName))
	}

	b.Unindent()
	b.P("}")
	b.P()

	return nil
} 

// packageToLibraryName converts a protobuf package name to a valid Solidity library name
func (g *Generator) packageToLibraryName(packageName string) string {
	// Replace dots with underscores and capitalize
	parts := strings.Split(packageName, ".")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "_")
}

// resolveTypeName resolves a protobuf type name to a Solidity type name with package support
func (g *Generator) resolveTypeName(typeName string) (string, error) {
	log.Printf("DEBUG: resolveTypeName called with typeName: '%s'", typeName)
	
	if len(typeName) == 0 {
		log.Printf("WARNING: Empty type name passed to resolveTypeName, using default type")
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

// dependencyToImportPath converts a protobuf dependency to a Solidity import path
func (g *Generator) dependencyToImportPath(dependency string) string {
	// Remove .proto extension if present
	if strings.HasSuffix(dependency, ".proto") {
		dependency = strings.TrimSuffix(dependency, ".proto")
	}
	
	// For now, assume relative imports in the same directory
	// In a more sophisticated implementation, this could:
	// 1. Look up the package name of the dependency
	// 2. Generate appropriate import paths based on package structure
	// 3. Handle different import strategies (relative vs absolute)
	
	return fmt.Sprintf("./%s.sol", dependency)
} 

// generateFloatDoubleHelpers generates helper functions for float/double fixed-point scaling
func (g *Generator) generateFloatDoubleHelpers(b *WriteableBuffer) {
	b.P("// Helper functions for float/double fixed-point scaling")
	b.P0()
	
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
	b.P0()

	b.P("// Convert IEEE 754 float to fixed-point int32 with 1e6 scaling")
	b.P("// This preserves 6 decimal places of precision")
	b.P("int32 scaled_value;")
	b.P("assembly {")
	b.Indent()
	b.P("// Extract sign, exponent, and mantissa from IEEE 754")
	b.P("let sign := shr(31, raw_value)")
	b.P("let exponent := and(shr(23, raw_value), 0xFF)")
	b.P("let mantissa := and(raw_value, 0x7FFFFF)")
	b.P0()

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
	b.P0()

	b.P("// Normal case: convert to fixed-point")
	b.P("// Add implicit leading 1 to mantissa")
	b.P("mantissa := or(mantissa, 0x800000)")
	b.P0()

	b.P("// Calculate actual value: mantissa * 2^(exponent-127)")
	b.P("let shift := sub(exponent, 127)")
	b.P("let scaled_mantissa := mantissa")
	b.P0()

	b.P("// Apply scaling factor of 1e6 (1,000,000)")
	b.P("scaled_mantissa := mul(scaled_mantissa, 1000000)")
	b.P0()

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
	b.P0()

	b.P("// Apply sign")
	b.P("if sign {")
	b.Indent()
	b.P("scaled_mantissa := mul(scaled_mantissa, -1)")
	b.Unindent()
	b.P("}")
	b.P0()

	b.P("scaled_value := scaled_mantissa")
	b.Unindent()
	b.P("}")
	b.P0()

	b.P("return (true, new_pos, scaled_value);")
	b.Unindent()
	b.P("}")
	b.P0()

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
	b.P0()

	b.P("// Convert IEEE 754 double to fixed-point int64 with 1e15 scaling")
	b.P("// This preserves 15 decimal places of precision")
	b.P("int64 scaled_value;")
	b.P("assembly {")
	b.Indent()
	b.P("// Extract sign, exponent, and mantissa from IEEE 754")
	b.P("let sign := shr(63, raw_value)")
	b.P("let exponent := and(shr(52, raw_value), 0x7FF)")
	b.P("let mantissa := and(raw_value, 0xFFFFFFFFFFFFF)")
	b.P0()

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
	b.P0()

	b.P("// Normal case: convert to fixed-point")
	b.P("// Add implicit leading 1 to mantissa")
	b.P("mantissa := or(mantissa, 0x10000000000000)")
	b.P0()

	b.P("// Calculate actual value: mantissa * 2^(exponent-1023)")
	b.P("let shift := sub(exponent, 1023)")
	b.P("let scaled_mantissa := mantissa")
	b.P0()

	b.P("// Apply scaling factor of 1e15 (1,000,000,000,000,000)")
	b.P("scaled_mantissa := mul(scaled_mantissa, 1000000000000000)")
	b.P0()

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
	b.P0()

	b.P("// Apply sign")
	b.P("if sign {")
	b.Indent()
	b.P("scaled_mantissa := mul(scaled_mantissa, -1)")
	b.Unindent()
	b.P("}")
	b.P0()

	b.P("scaled_value := scaled_mantissa")
	b.Unindent()
	b.P("}")
	b.P0()

	b.P("return (true, new_pos, scaled_value);")
	b.Unindent()
	b.P("}")
	b.P0()
} 