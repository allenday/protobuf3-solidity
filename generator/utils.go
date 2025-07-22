package generator

import (
	"fmt"
	"strings"
)

// PackageToLibraryName converts a package name to Solidity library name format
// Example: "a2a.v1" -> "A2a_V1", "" -> "DefaultPackage"
func PackageToLibraryName(packageName string) string {
	if len(packageName) == 0 {
		return "DefaultPackage"
	}
	
	parts := strings.Split(packageName, ".")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "_")
}

// IsGoogleProtobufDependency checks if a dependency is a Google protobuf type
func IsGoogleProtobufDependency(dependency string) bool {
	return strings.HasPrefix(dependency, "google/protobuf/")
}

// IsGoogleAPIDependency checks if a dependency is a Google API type
func IsGoogleAPIDependency(dependency string) bool {
	return strings.HasPrefix(dependency, "google/api/")
}

// IsGoogleDependency checks if a dependency is any Google-provided type
func IsGoogleDependency(dependency string) bool {
	return IsGoogleProtobufDependency(dependency) || IsGoogleAPIDependency(dependency)
}

// CreateListWrapperName creates a wrapper name for repeated fields
func CreateListWrapperName(fieldName string) string {
	return fmt.Sprintf("%sList", strings.Title(fieldName))
}

// CreateMapEntryWrapperName creates a wrapper name for map entry fields
func CreateMapEntryWrapperName(fieldName string) string {
	return fmt.Sprintf("%sEntry", strings.Title(fieldName))
}