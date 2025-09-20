package stats

import (
	"fmt"
	"io/fs"
	"path"
	"slices"
	"strings"
)

type PackageAndTests struct {
	Pkg   *Package
	Tests []*TestFile
}

type Module struct {
	GoMod    *GoMod
	Packages []*PackageAndTests
}

func (m *Module) Reload(filesystem fs.FS, path string) error {
	if strings.HasSuffix(path, "go.mod") {
		fmt.Println("Reloading go.mod")
		return m.reloadGoMod(filesystem, path)
	}

	if strings.HasSuffix(path, ".go") {
		fmt.Println("Reloading go file:", path)
		return m.reloadGoFile(filesystem, path)
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

func (m *Module) reloadGoFile(filesystem fs.FS, path string) error {
	directoryName := path[:strings.LastIndex(path, "/")]
	fileName := path[strings.LastIndex(path, "/")+1:]
	for _, pkgAndTests := range m.Packages {
		fmt.Println("Checking package:", pkgAndTests.Pkg.Directory)
		if pkgAndTests.Pkg.Directory != directoryName {
			continue
		}
		if !slices.Contains(pkgAndTests.Pkg.GoFiles, fileName) {
			continue
		}
		testFile, err := filesystem.Open(path)
		if err != nil {
			return fmt.Errorf("opening file '%s': %w", path, err)
		}
		parsedTestFile, err := ParseTestFile(testFile)
		if err != nil {
			return fmt.Errorf("parsing file '%s': %w", path, err)
		}
		parsedTestFile.Name = path
		parsedTestFile.Path = path

		for _, existingTest := range pkgAndTests.Tests {
			if existingTest.Name == parsedTestFile.Name {
				*existingTest = *parsedTestFile
				return nil
			}
		}

		return nil
	}

	return nil
}

func ParseModule(filesystem fs.FS, dirPath string) (*Module, error) {
	module := &Module{}

	goModPath := path.Join(dirPath, "go.mod")
	goModFile, err := filesystem.Open(goModPath)
	if err == nil {
		defer goModFile.Close()
		module.GoMod, err = ParseGoMod(goModFile)
		if err != nil {
			return nil, fmt.Errorf("failed to parse '%s': %w", goModPath, err)
		}
	}

	module.Packages, err = getPackageAndTests(filesystem, dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get packages and tests: %w", err)
	}

	return module, nil
}

func getPackageAndTests(filesystem fs.FS, dirPath string) ([]*PackageAndTests, error) {
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
		subDirPackagesAndTests, err := getPackageAndTests(filesystem, path.Join(dirPath, subDir))
		if err != nil {
			return nil, fmt.Errorf("failed to get packages and tests in subdirectory '%s': %w", subDir, err)
		}
		pkgAndTestsList = append(pkgAndTestsList, subDirPackagesAndTests...)
	}

	return pkgAndTestsList, nil
}
