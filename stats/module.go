package stats

import (
	"fmt"
	"io/fs"
	"os"
	"path"
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

	directory string
	ignores   gitignore.Matcher
}

func ParseModule(filesystem fs.FS, dirPath string, ignores gitignore.Matcher) (*Module, error) {
	module := &Module{
		directory: dirPath,
		ignores:   ignores,
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
