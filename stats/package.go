package stats

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"
)

type Package struct {
	Name string
	// Directory is the relative path within the fs where this package was found.
	Directory string

	// Subdirectories is the list of subdirectories within this package directory.
	Subdirectories []string

	// TestFiles is the list of *_test.go files.
	TestFiles []string

	// GoFiles is the list of non-test .go files.
	GoFiles []string

	// OtherFiles is the list of non-.go files.
	OtherFiles []string
}

func ParsePackage(filesystem fs.FS, dirPath string) (*Package, error) {
	pkg := &Package{
		Directory: path.Clean(dirPath),
	}

	entries, err := fs.ReadDir(filesystem, dirPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			pkg.Subdirectories = append(pkg.Subdirectories, entry.Name())
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, "_test.go") {
			pkg.TestFiles = append(pkg.TestFiles, name)
		} else if strings.HasSuffix(name, ".go") {
			pkg.GoFiles = append(pkg.GoFiles, name)
		} else {
			pkg.OtherFiles = append(pkg.OtherFiles, name)
			continue
		}

		if pkg.Name != "" {
			continue // already set
		}
		pkg.Name, err = detectPackageName(filesystem, path.Join(dirPath, name))
		if err != nil {
			return nil, fmt.Errorf("detect package name in '%s%c%s': %w", dirPath, os.PathSeparator, name, err)
		}
	}

	return pkg, nil
}

func detectPackageName(filesystem fs.FS, path string) (string, error) {
	f, err := filesystem.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if strings.HasPrefix(line, "package ") {
			// This accounts for comments after the package name.
			fields := strings.Fields(strings.TrimPrefix(line, "package "))
			return strings.TrimSuffix(fields[0], "_test"), nil
		}
	}
	return "", sc.Err()
}
