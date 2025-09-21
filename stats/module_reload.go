package stats

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"slices"
	"strings"
)

type ReloadType int

const (
	ReloadTypeUpdate ReloadType = iota
	ReloadTypeRemove
	ReloadTypeRename
	ReloadTypeCreate
)

// Reload reloads the whole module. This is a simple but inefficient way to handle file changes.
func (m *Module) Reload(filesystem fs.FS, fsPath string, reloadType ReloadType) error {
	newModule, err := ParseModule(filesystem, m.directory, m.ignores)
	if err != nil {
		return fmt.Errorf("reloading module: %w", err)
	}
	*m = *newModule
	fmt.Println("Reloading module")

	return nil
}

// FineGrainedReload reloads only the affected parts of the module based on the file change event.
// This is under development and does not cover all edge cases yet.
func (m *Module) FineGrainedReload(filesystem fs.FS, fsPath string, reloadType ReloadType) error {
	fsPath = path.Clean(fsPath)
	fmt.Println("Reloading path:", fsPath, "with type:", reloadType)

	if fsPath == "go.mod" {
		return m.reloadGoMod(filesystem, fsPath)
	}

	if strings.HasSuffix(fsPath, "_test.go") {
		return m.reloadGoTestFile(filesystem, fsPath)
	}

	if strings.HasSuffix(fsPath, ".go") {
		return m.reloadGoFile(filesystem, fsPath)
	}

	// TODO: We should check if it's a file or directory.
	// it might be a file or a directory, it's hard to tell with fsnotify events
	return m.reloadDirectory(filesystem, fsPath)

}

func (m *Module) reloadGoMod(filesystem fs.FS, fileName string) error {
	goModFile, err := filesystem.Open(fileName)
	if err != nil {
		return fmt.Errorf("opening file '%s': %w", fileName, err)
	}
	goMod, err := ParseGoMod(goModFile)
	if err != nil {
		return fmt.Errorf("parsing file '%s': %w", fileName, err)
	}
	m.GoMod = goMod
	return nil
}

func (m *Module) reloadGoTestFile(filesystem fs.FS, path string) error {
	return m.reloadFileHelper(filesystem, path,
		func(pkgAndTests *PackageAndTests, fileName string) error {
			if slices.Contains(pkgAndTests.Pkg.TestFiles, fileName) {
				pkgAndTests.Pkg.TestFiles = slices.DeleteFunc(pkgAndTests.Pkg.TestFiles, func(t string) bool {
					return t == fileName
				})
				pkgAndTests.Tests = slices.DeleteFunc(pkgAndTests.Tests, func(t *TestFile) bool {
					return t.Name == fileName
				})
			}
			return nil
		},
		func(pkgAndTests *PackageAndTests, fileName string, contents io.Reader) error {
			pkgAndTests.Pkg.TestFiles = append(pkgAndTests.Pkg.TestFiles, fileName)

			parsedTestFile, err := ParseTestFile(contents)
			if err != nil {
				return fmt.Errorf("parsing file '%s': %w", path, err)
			}

			parsedTestFile.Name = fileName
			parsedTestFile.Path = path
			if !slices.Contains(pkgAndTests.Pkg.TestFiles, fileName) {
				pkgAndTests.Tests = append(pkgAndTests.Tests, parsedTestFile)
			} else {
				for i, t := range pkgAndTests.Tests {
					if t.Name == fileName {
						pkgAndTests.Tests[i] = parsedTestFile
						break
					}
				}
			}
			return nil
		},
	)
}

func (m *Module) reloadGoFile(filesystem fs.FS, path string) error {
	return m.reloadFileHelper(filesystem, path,
		func(pkgAndTests *PackageAndTests, fileName string) error {
			if slices.Contains(pkgAndTests.Pkg.GoFiles, fileName) {
				pkgAndTests.Pkg.GoFiles = slices.DeleteFunc(pkgAndTests.Pkg.GoFiles, func(t string) bool {
					return t == fileName
				})
			}
			return nil
		},
		func(pkgAndTests *PackageAndTests, fileName string, contents io.Reader) error {
			if !slices.Contains(pkgAndTests.Pkg.GoFiles, fileName) {
				pkgAndTests.Pkg.GoFiles = append(pkgAndTests.Pkg.GoFiles, fileName)
			}
			return nil
		},
	)
}

func (m *Module) reloadFileHelper(filesystem fs.FS, path string, noLongExists func(pkgAndTests *PackageAndTests, fileName string) error, exists func(pkgAndTests *PackageAndTests, fileName string, contents io.Reader) error) error {
	lastIdx := strings.LastIndex(path, "/")
	if lastIdx == -1 {
		lastIdx = 0
	}
	directoryName := path[:lastIdx]
	fileName := path[strings.LastIndex(path, "/")+1:]

	for _, pkgAndTests := range m.Packages {
		if !isSamePath(directoryName, pkgAndTests.Pkg.Directory) {
			continue
		}
		file, err := filesystem.Open(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return noLongExists(pkgAndTests, fileName)
			}
			return fmt.Errorf("opening file '%s': %w", path, err)
		}
		defer file.Close()
		return exists(pkgAndTests, fileName, file)
	}

	return nil
}

func isSamePath(path1, path2 string) bool {
	if path1 == path2 {
		return true
	}
	return (path1 == "." && path2 == "") || (path1 == "" && path2 == ".")
}

func (m *Module) reloadDirectory(filesystem fs.FS, dirPath string) error {
	pkgAndTests, err := m.getPackageAndTests(filesystem, dirPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			for i, pt := range m.Packages {
				if isSamePath(pt.Pkg.Directory, dirPath) {
					m.Packages = slices.Delete(m.Packages, i, i+1)
					return nil
				}
			}
			return nil
		}
		return fmt.Errorf("getting package and tests for directory '%s': %w", dirPath, err)
	}

	for _, newPkg := range pkgAndTests {
		found := false
		for i, existingPkg := range m.Packages {
			if isSamePath(existingPkg.Pkg.Directory, newPkg.Pkg.Directory) {
				m.Packages[i] = existingPkg
				found = true
				break
			}
		}
		if !found {
			m.Packages = append(m.Packages, newPkg)
		}
	}

	return nil
}
