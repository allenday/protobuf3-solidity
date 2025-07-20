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
	b.P(fmt.Sprintf("import \"%s\";", im.dependencyToImportPath("ProtobufLib", protoFile.GetName())))

	// Track imported files to avoid duplicates
	importedFiles := make(map[string]bool)
	importedFiles[im.dependencyToImportPath("ProtobufLib", protoFile.GetName())] = true

	// Generate imports for dependencies
	for _, dependency := range protoFile.GetDependency() {
		if strings.HasPrefix(dependency, "google/protobuf/") || strings.HasPrefix(dependency, "google/api/") {
			continue
		}
		importPath := im.dependencyToImportPath(dependency, protoFile.GetName())
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
// Uses configured import path for ProtobufLib, calculated relative paths for other dependencies
func (im *ImportManager) dependencyToImportPath(dependency string, currentFileName string) string {
	// Remove .proto extension if present
	dependency = strings.TrimSuffix(dependency, ".proto")

	// Convert path separators to forward slashes
	dependency = strings.ReplaceAll(dependency, "\\", "/")

	// Handle ProtobufLib import - use configured path
	if dependency == "ProtobufLib" {
		return im.protobufLibImportPath
	}

	// Calculate relative path from current file to dependency
	return im.calculateRelativePath(currentFileName, dependency) + ".sol"
}

// calculateRelativePath calculates the relative path from current file to dependency
func (im *ImportManager) calculateRelativePath(currentFileName string, dependency string) string {
	// Remove .proto extension from current file
	currentFileName = strings.TrimSuffix(currentFileName, ".proto")

	// Split paths into directories
	currentDirs := strings.Split(currentFileName, "/")
	dependencyDirs := strings.Split(dependency, "/")

	// Find common prefix
	commonPrefixLen := 0
	for i := 0; i < len(currentDirs) && i < len(dependencyDirs); i++ {
		if currentDirs[i] == dependencyDirs[i] {
			commonPrefixLen++
		} else {
			break
		}
	}

	// Calculate relative path
	// Go up (currentDirs - commonPrefixLen) directories, then down to dependency
	upCount := len(currentDirs) - commonPrefixLen

	var relativePath strings.Builder

	// Add "../" for each level up
	for i := 0; i < upCount; i++ {
		relativePath.WriteString("../")
	}

	// Add the remaining dependency path
	for i := commonPrefixLen; i < len(dependencyDirs); i++ {
		if i > commonPrefixLen {
			relativePath.WriteString("/")
		}
		relativePath.WriteString(dependencyDirs[i])
	}

	return relativePath.String()
}
