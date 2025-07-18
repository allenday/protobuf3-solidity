package generator

import (
	"errors"
	"fmt"
	"strings"
)

// checkSyntaxVersion validates the protobuf syntax version
func checkSyntaxVersion(v string) error {
	if v != "proto3" {
		return errors.New("must use syntax = \"proto3\";")
	}
	return nil
}

// checkKeyword checks if a word is a Solidity keyword
func checkKeyword(w string) error {
	keywords := []string{
		"abstract", "after", "alias", "apply", "auto", "case", "catch", "copyof", "default", "define", "final", "immutable", "implements", "in", "inline", "let", "macro", "match", "mutable", "null", "of", "override", "partial", "promise", "reference", "relocatable", "sealed", "sizeof", "static", "supports", "switch", "try", "typedef", "typeof", "unchecked",
		"address", "bool", "byte", "bytes", "calldata", "contract", "enum", "error", "event", "external", "fallback", "function", "import", "internal", "library", "mapping", "memory", "modifier", "payable", "private", "public", "pure", "receive", "return", "returns", "storage", "string", "struct", "this", "view", "virtual",
		"break", "continue", "do", "else", "for", "if", "revert", "while",
		"emit", "new", "delete", "super", "assembly", "constructor", "indexed", "anonymous", "override", "virtual", "abstract", "constant", "immutable", "payable", "external", "internal", "private", "public", "pure", "view", "nonpayable", "indexed", "anonymous", "override", "virtual", "abstract", "constant", "immutable", "payable", "external", "internal", "private", "public", "pure", "view", "nonpayable",
		"block", "blockhash", "gasleft", "msg", "now", "tx", "abi", "assert", "require", "revert", "addmod", "mulmod", "keccak256", "sha256", "ripemd160", "ecrecover",
		"int", "int8", "int16", "int24", "int32", "int40", "int48", "int56", "int64", "int72", "int80", "int88", "int96", "int104", "int112", "int120", "int128", "int136", "int144", "int152", "int160", "int168", "int176", "int184", "int192", "int200", "int208", "int216", "int224", "int232", "int240", "int248", "int256",
		"uint", "uint8", "uint16", "uint24", "uint32", "uint40", "uint48", "uint56", "uint64", "uint72", "uint80", "uint88", "uint96", "uint104", "uint112", "uint120", "uint128", "uint136", "uint144", "uint152", "uint160", "uint168", "uint176", "uint184", "uint192", "uint200", "uint208", "uint216", "uint224", "uint232", "uint240", "uint248", "uint256",
		"byte1", "byte2", "byte3", "byte4", "byte5", "byte6", "byte7", "byte8", "byte9", "byte10", "byte11", "byte12", "byte13", "byte14", "byte15", "byte16", "byte17", "byte18", "byte19", "byte20", "byte21", "byte22", "byte23", "byte24", "byte25", "byte26", "byte27", "byte28", "byte29", "byte30", "byte31", "byte32",
		"bytes1", "bytes2", "bytes3", "bytes4", "bytes5", "bytes6", "bytes7", "bytes8", "bytes9", "bytes10", "bytes11", "bytes12", "bytes13", "bytes14", "bytes15", "bytes16", "bytes17", "bytes18", "bytes19", "bytes20", "bytes21", "bytes22", "bytes23", "bytes24", "bytes25", "bytes26", "bytes27", "bytes28", "bytes29", "bytes30", "bytes31", "bytes32",
	}

	for _, keyword := range keywords {
		if strings.ToLower(w) == keyword {
			return fmt.Errorf("reserved keyword: %s", w)
		}
	}

	return nil
}

// sanitizeKeyword renames reserved Solidity keywords by prefixing with underscore
func sanitizeKeyword(w string) string {
	keywords := []string{
		"abstract", "after", "alias", "apply", "auto", "case", "catch", "copyof", "default", "define", "final", "immutable", "implements", "in", "inline", "let", "macro", "match", "mutable", "null", "of", "override", "partial", "promise", "reference", "relocatable", "sealed", "sizeof", "static", "supports", "switch", "try", "typedef", "typeof", "unchecked",
		"address", "bool", "byte", "bytes", "calldata", "contract", "enum", "error", "event", "external", "fallback", "function", "import", "internal", "library", "mapping", "memory", "modifier", "payable", "private", "public", "pure", "receive", "return", "returns", "storage", "string", "struct", "this", "view", "virtual",
		"break", "continue", "do", "else", "for", "if", "revert", "while",
		"emit", "new", "delete", "super", "assembly", "constructor", "indexed", "anonymous", "override", "virtual", "abstract", "constant", "immutable", "payable", "external", "internal", "private", "public", "pure", "view", "nonpayable", "indexed", "anonymous", "override", "virtual", "abstract", "constant", "immutable", "payable", "external", "internal", "private", "public", "pure", "view", "nonpayable",
		"block", "blockhash", "gasleft", "msg", "now", "tx", "abi", "assert", "require", "revert", "addmod", "mulmod", "keccak256", "sha256", "ripemd160", "ecrecover",
		"int", "int8", "int16", "int24", "int32", "int40", "int48", "int56", "int64", "int72", "int80", "int88", "int96", "int104", "int112", "int120", "int128", "int136", "int144", "int152", "int160", "int168", "int176", "int184", "int192", "int200", "int208", "int216", "int224", "int232", "int240", "int248", "int256",
		"uint", "uint8", "uint16", "uint24", "uint32", "uint40", "uint48", "uint56", "uint64", "uint72", "uint80", "uint88", "uint96", "uint104", "uint112", "uint120", "uint128", "uint136", "uint144", "uint152", "uint160", "uint168", "uint176", "uint184", "uint192", "uint200", "uint208", "uint216", "uint224", "uint232", "uint240", "uint248", "uint256",
		"byte1", "byte2", "byte3", "byte4", "byte5", "byte6", "byte7", "byte8", "byte9", "byte10", "byte11", "byte12", "byte13", "byte14", "byte15", "byte16", "byte17", "byte18", "byte19", "byte20", "byte21", "byte22", "byte23", "byte24", "byte25", "byte26", "byte27", "byte28", "byte29", "byte30", "byte31", "byte32",
		"bytes1", "bytes2", "bytes3", "bytes4", "bytes5", "bytes6", "bytes7", "bytes8", "bytes9", "bytes10", "bytes11", "bytes12", "bytes13", "bytes14", "bytes15", "bytes16", "bytes17", "bytes18", "bytes19", "bytes20", "bytes21", "bytes22", "bytes23", "bytes24", "bytes25", "bytes26", "bytes27", "bytes28", "bytes29", "bytes30", "bytes31", "bytes32",
	}

	for _, keyword := range keywords {
		if strings.ToLower(w) == keyword {
			return "_" + w
		}
	}

	return w
} 