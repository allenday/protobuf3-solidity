package generator

import (
	"fmt"
)

// checkSyntaxVersion checks that the syntax version is proto3
func checkSyntaxVersion(syntax string) error {
	if syntax != "proto3" {
		return fmt.Errorf("only proto3 is supported, got %s", syntax)
	}
	return nil
} 