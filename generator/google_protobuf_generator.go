package generator

import (
	"google.golang.org/protobuf/types/descriptorpb"
)

// GoogleProtobufGenerator handles generation of Google protobuf type definitions
type GoogleProtobufGenerator struct{}

// NewGoogleProtobufGenerator creates a new Google protobuf generator
func NewGoogleProtobufGenerator() *GoogleProtobufGenerator {
	return &GoogleProtobufGenerator{}
}

// GenerateGoogleProtobufTypes generates Solidity definitions for Google protobuf types
func (gpg *GoogleProtobufGenerator) GenerateGoogleProtobufTypes(protoFile *descriptorpb.FileDescriptorProto, b *WriteableBuffer, alreadyGenerated bool) error {
	// Check if this file uses Google protobuf types
	usesGoogleTypes := false
	for _, dependency := range protoFile.GetDependency() {
		if IsGoogleProtobufDependency(dependency) {
			usesGoogleTypes = true
			break
		}
	}

	if !usesGoogleTypes {
		return nil
	}

	// Skip inline generation if already generated to avoid duplicates
	// The shared library approach handles this at the generator level
	if alreadyGenerated {
		return nil
	}

	// This method is now only called for inline generation (legacy behavior)
	// The shared library approach generates the types separately
	return nil
}

// Legacy methods - no longer used since switching to shared library approach
// Kept for backward compatibility but not called in current implementation
