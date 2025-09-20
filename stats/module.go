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

	"github.com/go-git/go-git/v6/plumbing/format/gitignore"
)

type PackageAndTests struct {
	Pkg   *Package
	Tests []*TestFile
}

type Module struct {
	GoMod    *GoMod
	Packages []*PackageAndTests

	ignores gitignore.Matcher
}

func (m *Module) Reload(filesystem fs.FS, fsPath string) error {
	fsPath = path.Clean(fsPath)
	if fsPath == "go.mod" {
		return m.reloadGoMod(filesystem, fsPath)
	}

	if strings.HasSuffix(fsPath, "_test.go") {
		return m.reloadGoTestFile(filesystem, fsPath)
	}

	if strings.HasSuffix(fsPath, ".go") {
		return m.reloadGoFile(filesystem, fsPath)
	}

	// otherwise, it's a directory and we should parse the whole package again.

	return nil
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
		if !isDirectory(directoryName, pkgAndTests.Pkg.Directory) {
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

func isDirectory(dir1, dir2 string) bool {
	if dir1 == dir2 {
		return true
	}
	return (dir1 == "." && dir2 == "") || (dir1 == "" && dir2 == ".")
}

func ParseModule(filesystem fs.FS, dirPath string, ignores gitignore.Matcher) (*Module, error) {
	module := &Module{
		ignores: ignores,
	}

	goModPath := path.Join(dirPath, "go.mod")
	goModFile, err := filesystem.Open(goModPath)
	if err == nil {
		defer goModFile.Close()
		module.GoMod, err = ParseGoMod(goModFile)
		if err != nil {
			return nil, fmt.Errorf("failed to parse '%s': %w", goModPath, err)
		}
	}

	module.Packages, err = module.getPackageAndTests(filesystem, dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get packages and tests: %w", err)
	}

	return module, nil
}

func (m *Module) getPackageAndTests(filesystem fs.FS, dirPath string) ([]*PackageAndTests, error) {
	pkg, err := ParsePackage(filesystem, dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse package in '%s': %w", dirPath, err)
	}

	pkgAndTests := &PackageAndTests{
		Pkg: pkg,
	}

	for _, testFileName := range pkg.TestFiles {
		path := path.Join(dirPath, testFileName)
		testFile, err := filesystem.Open(path)
		if err != nil {
			return nil, fmt.Errorf("failed to open test file '%s': %w", path, err)
		}

		parsedTestFile, err := ParseTestFile(testFile)
		testFile.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to parse test file '%s': %w", path, err)
		}

		parsedTestFile.Name = testFileName
		parsedTestFile.Path = path
		pkgAndTests.Tests = append(pkgAndTests.Tests, parsedTestFile)
	}

	pkgAndTestsList := []*PackageAndTests{pkgAndTests}

	for _, subDir := range pkg.Subdirectories {
		if m.ignores.Match(strings.Split(subDir, string(os.PathSeparator)), false) {
			continue
		}
		subDirPackagesAndTests, err := m.getPackageAndTests(filesystem, path.Join(dirPath, subDir))
		if err != nil {
			return nil, fmt.Errorf("failed to get packages and tests in subdirectory '%s': %w", subDir, err)
		}
		pkgAndTestsList = append(pkgAndTestsList, subDirPackagesAndTests...)
	}

	return pkgAndTestsList, nil
}
