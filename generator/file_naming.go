package generator

import (
	"fmt"
	"path/filepath"
	"strings"

	"google.golang.org/protobuf/types/descriptorpb"
)

// FileNaming handles output file naming and path generation
type FileNaming struct{}

// NewFileNaming creates a new file naming handler
func NewFileNaming() *FileNaming {
	return &FileNaming{}
}

// GenerateOutputFileName generates the output file name for a proto file
func (fn *FileNaming) GenerateOutputFileName(protoFile *descriptorpb.FileDescriptorProto) string {
	fileName := protoFile.GetName()
	packageName := protoFile.GetPackage()

	var outFileName string
	if len(packageName) > 0 {
		// Convert package name to path format and use original file name
		packagePath := strings.ReplaceAll(packageName, ".", "/")
		baseName := strings.TrimSuffix(filepath.Base(fileName), ".proto")
		// Create directory structure that matches imports
		outFileName = fmt.Sprintf("%s/%s.sol", packagePath, baseName)
	} else {
		// For files without package, use the file name without .proto
		baseName := strings.TrimSuffix(filepath.Base(fileName), ".proto")
		outFileName = fmt.Sprintf("%s.sol", baseName)
	}

	return outFileName
}
