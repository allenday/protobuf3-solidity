package generator

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/types/descriptorpb"
)

// ImportManager handles import generation and dependency tracking
type ImportManager struct {
	protobufLibImportPath string
}

// NewImportManager creates a new import manager
func NewImportManager(protobufLibImportPath string) *ImportManager {
	return &ImportManager{
		protobufLibImportPath: protobufLibImportPath,
	}
}

// GenerateImports generates all necessary imports for a proto file
func (im *ImportManager) GenerateImports(protoFile *descriptorpb.FileDescriptorProto, b *WriteableBuffer) {
	// Add ProtobufLib import
	b.P(fmt.Sprintf("import \"%s\";", im.dependencyToImportPath("ProtobufLib")))

	// Track imported files to avoid duplicates
	importedFiles := make(map[string]bool)
	importedFiles[im.dependencyToImportPath("ProtobufLib")] = true

	// Generate imports for dependencies
	for _, dependency := range protoFile.GetDependency() {
		if strings.HasPrefix(dependency, "google/protobuf/") || strings.HasPrefix(dependency, "google/api/") {
			continue
		}
		importPath := im.dependencyToImportPath(dependency)
		if !importedFiles[importPath] {
			b.P(fmt.Sprintf("import \"%s\";", importPath))
			importedFiles[importPath] = true
		}
	}

	if len(protoFile.GetDependency()) > 0 {
		b.P0()
	}
}

// dependencyToImportPath converts a protobuf dependency to a Solidity import path
// Uses configured import path for ProtobufLib, local paths for other dependencies
func (im *ImportManager) dependencyToImportPath(dependency string) string {
	// Remove .proto extension if present
	dependency = strings.TrimSuffix(dependency, ".proto")

	// Convert path separators to forward slashes
	dependency = strings.ReplaceAll(dependency, "\\", "/")

	// Handle ProtobufLib import - use configured path
	if dependency == "ProtobufLib" {
		return im.protobufLibImportPath
	}

	// For all other imports, use local paths
	return dependency + ".sol"
}
