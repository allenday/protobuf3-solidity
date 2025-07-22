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
func (im *ImportManager) GenerateImports(protoFile *descriptorpb.FileDescriptorProto, generatedFileName string, b *WriteableBuffer) {
	// Add ProtobufLib import
	b.P(fmt.Sprintf("import \"%s\";", im.dependencyToImportPath("ProtobufLib", generatedFileName)))

	// Track imported files to avoid duplicates
	importedFiles := make(map[string]bool)
	importedFiles[im.dependencyToImportPath("ProtobufLib", generatedFileName)] = true

	// Check if this file uses Google protobuf types and add shared library import
	usesGoogleTypes := false
	for _, dependency := range protoFile.GetDependency() {
		if IsGoogleProtobufDependency(dependency) {
			usesGoogleTypes = true
			break
		}
	}

	if usesGoogleTypes {
		googleProtobufImportPath := im.calculateRelativePath(generatedFileName, "google/protobuf/google_protobuf") + ".sol"
		b.P(fmt.Sprintf("import \"%s\";", googleProtobufImportPath))
		importedFiles[googleProtobufImportPath] = true
	}

	// Generate imports for dependencies
	for _, dependency := range protoFile.GetDependency() {
		if IsGoogleDependency(dependency) {
			continue
		}
		importPath := im.dependencyToImportPath(dependency, generatedFileName)
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

	// Calculate relative path from current generated file to dependency
	// We need to use the generated file path, not the proto file path
	currentGeneratedPath := currentFileName
	dependencyGeneratedPath := im.getGeneratedFilePath(dependency)

	return im.calculateRelativePath(currentGeneratedPath, dependencyGeneratedPath) + ".sol"
}

// calculateRelativePath calculates the relative path from current file to dependency
func (im *ImportManager) calculateRelativePath(currentFileName string, dependency string) string {
	// Remove .sol extension from current file (it's already a generated file path)
	currentFileName = strings.TrimSuffix(currentFileName, ".sol")

	// Get the directory path of the current file (remove the filename)
	currentDir := currentFileName
	if lastSlash := strings.LastIndex(currentFileName, "/"); lastSlash != -1 {
		currentDir = currentFileName[:lastSlash]
	}

	// Split paths into directories
	currentDirs := strings.Split(currentDir, "/")
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

// getGeneratedFilePath converts a proto file path to its generated Solidity file path
func (im *ImportManager) getGeneratedFilePath(protoFilePath string) string {
	// Remove .proto extension
	protoFilePath = strings.TrimSuffix(protoFilePath, ".proto")

	// For dependencies, we need to convert the proto file path to the generated file path
	// The generated file path is based on the package name, which is typically
	// the same as the directory structure of the proto file
	// For example: a2a/v1/a2a.proto -> a2a/v1/a2a.sol
	return protoFilePath
}
