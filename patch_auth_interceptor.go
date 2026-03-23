package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	b, err := os.ReadFile("srcs/orchestration/auth_interceptor.go")
	if err != nil {
		panic(err)
	}
	content := string(b)

	newFunc := `// countSegments explicitly counts the total number of segments, even those exceeding the expected length,
// to accurately enforce strict length constraints and prevent path-based spoofing attacks (e.g. %2F).
func countSegments(s string) int {
	segments := 1
	for i := 0; i < len(s); i++ {
		if s[i] == '/' {
			segments++
		} else if i+2 < len(s) && s[i] == '%' && s[i+1] == '2' && (s[i+2] == 'f' || s[i+2] == 'F') {
			segments++
			i += 2
		}
	}
	return segments
}

// SPIFFEAuthInterceptor validates SPIFFE IDs for incoming gRPC calls.`

	content = strings.Replace(content, "// SPIFFEAuthInterceptor validates SPIFFE IDs for incoming gRPC calls.", newFunc, 1)

	// Replace `strings.Count(rest, "/")` with `countSegments(rest) - 1`
	content = strings.ReplaceAll(content, `strings.Count(rest, "/")`, `countSegments(rest) - 1`)

	err = os.WriteFile("srcs/orchestration/auth_interceptor.go", []byte(content), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Println("Patched auth_interceptor.go")
}
