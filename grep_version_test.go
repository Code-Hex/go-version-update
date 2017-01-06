package update

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestGrepVersion(t *testing.T) {
	files := []string{
		"grep_version.go",
		"version_update.go",
	}
	createGoProject(t, files)

	info, err := GrepVersion("_testdata")
	if err != nil {
		t.Fatalf("Failed to GrepVersion: %s", err.Error())
	}

	for _, i := range info {
		v := fmt.Sprintf(`"%s"`, Version)
		if i.Version != v {
			t.Errorf("got %s, expected: %s", i.Version, v)
		}

		path, err := filepath.Abs(filepath.Join("_testdata", "version_update.go"))
		if err != nil {
			t.Fatal(err)
		}

		if i.Path != path {
			t.Errorf("got %s, expected: %s", i.Path, path)
		}
	}

	removeTestFiles(files)
}

func createGoProject(t *testing.T, files []string) {
	for _, file := range files {
		orig, err := os.Open(file)
		if err != nil {
			t.Fatalf("Could not open source code: %s", err.Error())
		}
		path := filepath.Join("_testdata", file)
		dst, err := os.Create(path)
		if err != nil {
			t.Fatalf("Could not create testcode: %s", err.Error())
		}

		io.Copy(dst, orig)
		orig.Close()
		dst.Close()
	}
}

func removeTestFiles(files []string) {
	// Remove test code
	for _, file := range files {
		os.Remove(filepath.Join("_testdata", file))
	}
}
