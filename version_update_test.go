package update

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestNextVersion(t *testing.T) {

	cases := []struct {
		previous string
		new      string
		src      string
		expected string
	}{
		{"1.2.3", "4.0", createCode("var", "1.2.3"), createCode("var", "4.0")},
		{"1.30", "2.2000001", createCode("var", "1.30"), createCode("var", "2.2000001")},
		{"1.2beta", "1.2", createCode("const", "1.2beta"), createCode("const", "1.2")},
	}

	for _, cased := range cases {
		path := filepath.Join("_testdata", "test.go")
		fi, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			t.Fatalf("Could not create testcode: %s", err.Error())
		}
		fi.WriteString(cased.src)
		fi.Close()

		var buf bytes.Buffer
		if err := NextVersion(&buf, cased.new, path); err != nil {
			t.Fatalf("Failed to format testcode: %s", err.Error())
		}

		if cased.expected != buf.String() {
			t.Errorf("Got: %s, expected: %s", buf.String(), cased.expected)
		}
	}

	os.Remove(filepath.Join("_testdata", "test.go"))
}

func createCode(types, version string) string {
	return fmt.Sprintf(`package main

import "fmt"

%s version = "%s"

// This is main
func main() {
	fmt.Println("Hello, World")
}
`, types, version)
}
