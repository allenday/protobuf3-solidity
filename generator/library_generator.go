package generator

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/types/descriptorpb"
)

// LibraryGenerator handles generation of library structures
type LibraryGenerator struct {
	generateFlag generateFlag
}

// NewLibraryGenerator creates a new library generator
func NewLibraryGenerator(generateFlag generateFlag) *LibraryGenerator {
	return &LibraryGenerator{
		generateFlag: generateFlag,
	}
}

// GenerateMainLibrary generates the main library structure for a package
func (lg *LibraryGenerator) GenerateMainLibrary(packageName string, b *WriteableBuffer) {
	// Generate library name from package name
	libraryName := lg.packageToLibraryName(packageName)
	b.P(fmt.Sprintf("library %s {", libraryName))
	b.Indent()
	b.P0()
}

// CloseMainLibrary closes the main library block
func (lg *LibraryGenerator) CloseMainLibrary(b *WriteableBuffer) {
	b.Unindent()
	b.P("}")
	b.P0()
}

// GenerateEnums generates all enums for a proto file
func (lg *LibraryGenerator) GenerateEnums(protoFile *descriptorpb.FileDescriptorProto, g *Generator, b *WriteableBuffer) error {
	for _, enum := range protoFile.GetEnumType() {
		err := g.generateEnum(enum, b)
		if err != nil {
			return err
		}
	}
	return nil
}

// GenerateMessageStructs generates all message structs for a proto file
func (lg *LibraryGenerator) GenerateMessageStructs(protoFile *descriptorpb.FileDescriptorProto, g *Generator, b *WriteableBuffer) error {
	packageName := protoFile.GetPackage()

	// Track successfully generated structs to ensure codec generation matches
	if g.successfullyGeneratedStructs == nil {
		g.successfullyGeneratedStructs = make(map[string]bool)
	}

	// Generate messages (structs only, codec libraries will be generated separately)
	for _, message := range protoFile.GetMessageType() {
		err := g.generateMessageStruct(message, packageName, b)
		if err != nil {
			return err
		}
		// Mark this message as successfully processed
		g.successfullyGeneratedStructs[message.GetName()] = true
	}

	// Generate helper messages (structs only, codec libraries will be generated separately)
	if g.helperMessages[packageName] != nil {
		b.P("// Helper messages for PostFiat enhancements")
		b.P0()
		for _, helperMessage := range g.helperMessages[packageName] {
			err := g.generateMessageStruct(helperMessage, packageName, b)
			if err != nil {
				return err
			}
			// Mark this helper message as successfully processed
			g.successfullyGeneratedStructs[helperMessage.GetName()] = true
		}
	}

	return nil
}

// GenerateCodecLibraries generates all codec libraries outside the main library
func (lg *LibraryGenerator) GenerateCodecLibraries(protoFile *descriptorpb.FileDescriptorProto, g *Generator, b *WriteableBuffer) error {
	packageName := protoFile.GetPackage()

	// Generate codec libraries OUTSIDE the main library block
	// Only generate codecs for messages that have successfully generated structs
	for _, message := range protoFile.GetMessageType() {
		if g.successfullyGeneratedStructs[message.GetName()] {
			err := g.generateMessageCodec(message, packageName, b)
			if err != nil {
				return err
			}
		}
	}

	// Generate helper message codec libraries OUTSIDE the main library block
	// Only generate codecs for helper messages that have successfully generated structs
	if g.helperMessages[packageName] != nil {
		for _, helperMessage := range g.helperMessages[packageName] {
			if g.successfullyGeneratedStructs[helperMessage.GetName()] {
				err := g.generateMessageCodec(helperMessage, packageName, b)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// packageToLibraryName converts a protobuf package name to a valid Solidity library name
func (lg *LibraryGenerator) packageToLibraryName(packageName string) string {
	// Handle empty package name
	if len(packageName) == 0 {
		return "DefaultPackage"
	}

	// Replace dots with underscores and capitalize
	parts := strings.Split(packageName, ".")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "_")
}
