package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	b, err := os.ReadFile("srcs/dashboard/server_missing_test.go")
	if err != nil {
		panic(err)
	}
	content := string(b)

    // TestHandleScale redeclared in this block
	// TestHandleScaleStream redeclared in this block
    // We just rename them to TestHandleScale2 and TestHandleScaleStream2 where they are duplicated.

    // Find the second occurrence and rename
    firstScale := strings.Index(content, "func TestHandleScale(")
    if firstScale != -1 {
        secondScale := strings.Index(content[firstScale+len("func TestHandleScale("):], "func TestHandleScale(")
        if secondScale != -1 {
            secondScale += firstScale + len("func TestHandleScale(")
            content = content[:secondScale] + "func TestHandleScale2(" + content[secondScale+len("func TestHandleScale("):]
        }
    }

    firstStream := strings.Index(content, "func TestHandleScaleStream(")
    if firstStream != -1 {
        secondStream := strings.Index(content[firstStream+len("func TestHandleScaleStream("):], "func TestHandleScaleStream(")
        if secondStream != -1 {
            secondStream += firstStream + len("func TestHandleScaleStream(")
            content = content[:secondStream] + "func TestHandleScaleStream2(" + content[secondStream+len("func TestHandleScaleStream("):]
        }
    }

	err = os.WriteFile("srcs/dashboard/server_missing_test.go", []byte(content), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Println("Patched server_missing_test.go")
}
