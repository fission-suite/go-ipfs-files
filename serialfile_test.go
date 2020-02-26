package files

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func isFullPathHidden(p string) bool {
	return strings.HasPrefix(p, ".") || strings.Contains(p, "/.")
}

func TestSerialFile(t *testing.T) {
	t.Run("Hidden", func(t *testing.T) { testSerialFile(t, true) })
	t.Run("NotHidden", func(t *testing.T) { testSerialFile(t, false) })
}

func testSerialFile(t *testing.T, hidden bool) {
	tmppath, err := ioutil.TempDir("", "files-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmppath)

	expected := map[string]string{
		"1":      "Some text!\n",
		"2":      "beep",
		"3":      "",
		"4":      "boop",
		"5":      "",
		"5/a":    "foobar",
		".6":     "thing",
		"7":      "",
		"7/.foo": "bla",
		".8":     "",
		".8/foo": "bla",
	}

	for p, c := range expected {
		path := filepath.Join(tmppath, p)
		if c != "" {
			continue
		}
		if err := os.MkdirAll(path, 0777); err != nil {
			t.Fatal(err)
		}
	}

	expectedHiddenPaths := make([]string, 0, 4)
	expectedRegularPaths := make([]string, 0, 6)
	for p, c := range expected {
		path := filepath.Join(tmppath, p)
		if isFullPathHidden(path) {
			expectedHiddenPaths = append(expectedHiddenPaths, p)
		} else {
			expectedRegularPaths = append(expectedRegularPaths, p)
		}
		if c == "" {
			continue
		}
		if err := ioutil.WriteFile(path, []byte(c), 0666); err != nil {
			t.Fatal(err)
		}
	}

	stat, err := os.Stat(tmppath)
	if err != nil {
		t.Fatal(err)
	}

	sf, err := NewSerialFile(tmppath, hidden, stat)
	if err != nil {
		t.Fatal(err)
	}
	defer sf.Close()

	rootFound := false
	actualRegularPaths := make([]string, 0, len(expectedRegularPaths))
	actualHiddenPaths := make([]string, 0, len(expectedHiddenPaths))
	err = Walk(sf, func(path string, nd Node) error {
		defer nd.Close()

		// root node.
		if path == "" {
			if rootFound {
				return fmt.Errorf("found root twice")
			}
			if sf != nd {
				return fmt.Errorf("wrong root")
			}
			rootFound = true
			return nil
		}
		if isFullPathHidden(path) {
			actualHiddenPaths = append(actualHiddenPaths, path)
		} else {
			actualRegularPaths = append(actualRegularPaths, path)
		}
		if !hidden && isFullPathHidden(path) {
			return fmt.Errorf("found a hidden file")
		}

		data, ok := expected[path]
		if !ok {
			return fmt.Errorf("expected something at %q", path)
		}
		delete(expected, path)

		switch nd := nd.(type) {
		case *Symlink:
			return fmt.Errorf("didn't expect a symlink")
		case Directory:
			if data != "" {
				return fmt.Errorf("expected a directory at %q", path)
			}
		case File:
			actual, err := ioutil.ReadAll(nd)
			if err != nil {
				return err
			}
			if string(actual) != data {
				return fmt.Errorf("expected %q, got %q", data, string(actual))
			}
		}
		return nil
	})
	if !rootFound {
		t.Fatal("didn't find the root")
	}
	for _, regular := range expectedRegularPaths {
		if idx := sort.SearchStrings(actualRegularPaths, regular); idx < 0 {
			t.Errorf("missed regular path %q", regular)
		}
	}
	if hidden && len(actualHiddenPaths) != len(expectedHiddenPaths) {
		for _, missing := range expectedHiddenPaths {
			if idx := sort.SearchStrings(actualHiddenPaths, missing); idx < 0 {
				t.Errorf("missed %q", missing)
			}
		}
	}
}
