package stats

import (
	"fmt"
	"io/fs"
	"path"
)

type PackageAndTests struct {
	Pkg   *Package
	Tests []*TestFile
}

type Module struct {
	GoMod    *GoMod
	Packages []*PackageAndTests
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
