// Binary gen_ts generates TypeScript type definitions from proto source files.
//
// It performs a lightweight parse of .proto files to extract enum and message
// definitions, then emits a single TypeScript file with interface and union-type
// declarations that mirror the proto schema.
//
// Usage:
//
//	gen_ts <proto_file>... > proto_types.ts
//
// Field names are converted from snake_case to camelCase. Types are prefixed
// with a CamelCase version of their proto package to avoid name collisions
// across multiple proto files (e.g. ohc.common.Role → CommonRole).
package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// ── Proto AST types ───────────────────────────────────────────────────────────

type enumValue struct{ name string }

type enumDef struct {
	name   string
	values []enumValue
}

type fieldDef struct {
	name     string
	typeName string // already resolved to the TS output name
	repeated bool
	optional bool // all proto3 scalar fields are optional
}

type messageDef struct {
	name   string
	fields []fieldDef
}

type protoFile struct {
	pkg      string // e.g. "ohc.common"
	prefix   string // CamelCase prefix derived from pkg, e.g. "Common"
	enums    []enumDef
	messages []messageDef
}

// ── Helpers ───────────────────────────────────────────────────────────────────

var reWord = regexp.MustCompile(`[A-Za-z0-9]+`)

// snakeToCamel converts a snake_case identifier to lowerCamelCase.
func snakeToCamel(s string) string {
	parts := strings.Split(s, "_")
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

// pkgToPrefix converts a proto package name to a CamelCase prefix.
// e.g. "ohc.common" → "Common", "ohc.api.v1" → "Api"
func pkgToPrefix(pkg string) string {
	parts := strings.Split(pkg, ".")
	// Use the last non-version segment (skip "ohc" and version segments).
	for i := len(parts) - 1; i >= 0; i-- {
		p := parts[i]
		if p == "" || p == "ohc" {
			continue
		}
		// skip version identifiers like "v1", "v2"
		if len(p) >= 2 && (p[0] == 'v' || p[0] == 'V') {
			isVer := true
			for _, c := range p[1:] {
				if c < '0' || c > '9' {
					isVer = false
					break
				}
			}
			if isVer {
				continue
			}
		}
		return strings.ToUpper(p[:1]) + p[1:]
	}
	return "Proto"
}

// qualifiedToTS resolves a (possibly-qualified) proto type name to its
// TypeScript equivalent, given a map of all known types keyed by proto name.
func qualifiedToTS(typeName string, allTypes map[string]string) string {
	// Strip leading dot
	typeName = strings.TrimPrefix(typeName, ".")

	// Check exact match first
	if ts, ok := allTypes[typeName]; ok {
		return ts
	}
	// Try matching by the last segment (unqualified name)
	parts := strings.Split(typeName, ".")
	unqualified := parts[len(parts)-1]
	if ts, ok := allTypes[unqualified]; ok {
		return ts
	}
	// Fall back to CamelCase of the unqualified name
	return strings.ToUpper(unqualified[:1]) + unqualified[1:]
}

// protoTypeToTS converts a primitive proto type name to a TypeScript type string.
func protoTypeToTS(t string) string {
	switch t {
	case "string":
		return "string"
	case "bool":
		return "boolean"
	case "int32", "int64", "uint32", "uint64",
		"sint32", "sint64", "fixed32", "fixed64", "sfixed32", "sfixed64",
		"float", "double":
		return "number"
	case "bytes":
		return "string" // base64-encoded in JSON
	default:
		return "" // handled separately for message/enum types
	}
}

// ── Parser ────────────────────────────────────────────────────────────────────

// stripComment removes inline // and /* */ comments from a line.
func stripComment(line string) string {
	if idx := strings.Index(line, "//"); idx >= 0 {
		line = line[:idx]
	}
	if idx := strings.Index(line, "/*"); idx >= 0 {
		line = line[:idx]
	}
	return strings.TrimSpace(line)
}

var (
	rePkg     = regexp.MustCompile(`^package\s+([\w.]+)\s*;`)
	reEnum    = regexp.MustCompile(`^enum\s+(\w+)\s*\{`)
	reMsg     = regexp.MustCompile(`^message\s+(\w+)\s*\{`)
	reEnumVal = regexp.MustCompile(`^(\w+)\s*=\s*\d+\s*;`)
	reField   = regexp.MustCompile(`^(repeated\s+|optional\s+)?([\w.]+)\s+(\w+)\s*=\s*\d+`)
)

func parseProto(path string) protoFile {
	f, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: cannot open %s: %v\n", path, err)
		return protoFile{}
	}
	defer f.Close()

	var pf protoFile
	var inEnum, inMsg bool
	var curEnum enumDef
	var curMsg messageDef
	depth := 0

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		raw := scanner.Text()
		line := stripComment(raw)
		if line == "" {
			continue
		}

		// Count braces to track nesting.
		opens := strings.Count(line, "{")
		closes := strings.Count(line, "}")

		// Package declaration.
		if m := rePkg.FindStringSubmatch(line); m != nil {
			pf.pkg = m[1]
			pf.prefix = pkgToPrefix(pf.pkg)
			continue
		}

		// Skip option/import/syntax/edition lines.
		if strings.HasPrefix(line, "option ") ||
			strings.HasPrefix(line, "import ") ||
			strings.HasPrefix(line, "syntax ") ||
			strings.HasPrefix(line, "edition ") {
			continue
		}

		// Enum start.
		if m := reEnum.FindStringSubmatch(line); m != nil && depth == 0 {
			inEnum = true
			curEnum = enumDef{name: m[1]}
			depth += opens - closes
			continue
		}

		// Message start.
		if m := reMsg.FindStringSubmatch(line); m != nil && depth == 0 {
			inMsg = true
			curMsg = messageDef{name: m[1]}
			depth += opens - closes
			continue
		}

		// Closing brace at top level.
		if closes > 0 && depth == 1 {
			depth -= closes
			if depth <= 0 {
				depth = 0
				if inEnum {
					pf.enums = append(pf.enums, curEnum)
					inEnum = false
				}
				if inMsg {
					pf.messages = append(pf.messages, curMsg)
					inMsg = false
				}
			}
			continue
		}

		depth += opens - closes

		// Inside enum: collect values.
		if inEnum && depth == 1 {
			if m := reEnumVal.FindStringSubmatch(line); m != nil {
				curEnum.values = append(curEnum.values, enumValue{name: m[1]})
			}
			continue
		}

		// Inside message: collect fields (only at depth 1 to skip oneofs/maps).
		if inMsg && depth == 1 {
			if m := reField.FindStringSubmatch(line); m != nil {
				prefix := strings.TrimSpace(m[1])
				repeated := strings.HasPrefix(prefix, "repeated")
				optional := strings.HasPrefix(prefix, "optional")
				rawType := m[2]
				fieldName := m[3]
				// Skip reserved/map/oneof pseudo-fields.
				if fieldName == "reserved" || rawType == "map" {
					continue
				}
				curMsg.fields = append(curMsg.fields, fieldDef{
					name:     fieldName,
					typeName: rawType, // resolved later
					repeated: repeated,
					optional: optional,
				})
			}
		}
	}
	return pf
}

// ── Code generator ────────────────────────────────────────────────────────────

func generate(files []protoFile) {
	// First pass: build a map from proto qualified name → TS type name.
	// This lets us resolve cross-file references.
	allTypes := map[string]string{}
	for _, pf := range files {
		for _, e := range pf.enums {
			tsName := pf.prefix + e.name
			allTypes[e.name] = tsName
			allTypes[pf.pkg+"."+e.name] = tsName
		}
		for _, m := range pf.messages {
			tsName := pf.prefix + m.name
			allTypes[m.name] = tsName
			allTypes[pf.pkg+"."+m.name] = tsName
		}
	}

	fmt.Println("// GENERATED FILE – do not edit manually.")
	fmt.Println("// Generated at build time from proto definitions in srcs/proto/.")
	fmt.Println("// Run:  bazel build //srcs/proto:proto_types_ts")
	fmt.Println()

	for _, pf := range files {
		if len(pf.enums) == 0 && len(pf.messages) == 0 {
			continue
		}
		fmt.Printf("// ── %s ──\n\n", pf.pkg)

		// Emit enums.
		for _, e := range pf.enums {
			tsName := pf.prefix + e.name
			if len(e.values) == 0 {
				continue
			}
			// Sort enum values for deterministic output.
			vals := make([]string, len(e.values))
			for i, v := range e.values {
				vals[i] = v.name
			}
			sort.Strings(vals)
			fmt.Printf("export type %s =\n", tsName)
			for _, v := range vals {
				fmt.Printf("  | %q\n", v)
			}
			fmt.Println(";")
			fmt.Println()
		}

		// Emit messages.
		for _, msg := range pf.messages {
			tsName := pf.prefix + msg.name
			fmt.Printf("export interface %s {\n", tsName)
			for _, field := range msg.fields {
				camel := snakeToCamel(field.name)
				tsType := protoTypeToTS(field.typeName)
				if tsType == "" {
					// Message or enum type.
					tsType = qualifiedToTS(field.typeName, allTypes)
				}
				if field.repeated {
					tsType = tsType + "[]"
				}
				optional := "?"
				if field.repeated {
					optional = "?" // repeated fields are still optional in JSON
				}
				fmt.Printf("  %s%s: %s;\n", camel, optional, tsType)
			}
			fmt.Println("}")
			fmt.Println()
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: gen_ts <proto_file>...")
		os.Exit(1)
	}

	var files []protoFile
	for _, arg := range os.Args[1:] {
		// Resolve symlinks / runfiles paths.
		resolved := arg
		if _, err := os.Stat(resolved); err != nil {
			// Try relative to cwd.
			cwd, _ := os.Getwd()
			resolved = filepath.Join(cwd, arg)
		}
		pf := parseProto(resolved)
		if pf.pkg != "" {
			files = append(files, pf)
		}
	}

	generate(files)
}
