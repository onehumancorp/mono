//go:build tools

// Package tools provides build-time tool dependencies that are not imported
// in regular source code.  The //go:build tools constraint ensures this file
// is excluded from normal compilation while still being visible to `go mod tidy`
// so that the tools remain pinned in go.mod.
package tools

import (
	// protoc-gen-go-grpc code generator; must be >= v1.6.2 to support proto edition 2024.
	_ "google.golang.org/grpc/cmd/protoc-gen-go-grpc"
)
