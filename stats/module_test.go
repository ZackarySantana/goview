package stats

import (
	"fmt"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/go-git/go-git/v6/plumbing/format/gitignore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseModule(t *testing.T) {
	for name, tc := range map[string]struct {
		fsys     fstest.MapFS
		dir      string
		expected *Module
	}{
		"Parses": {
			fsys: fstest.MapFS{
				"go.mod":                      {Data: []byte(goModContent)},
				"testfile_test.go":            {Data: []byte(testFileContent)},
				"subdir/file.go":              {Data: []byte("package subdir")},
				"subdir/file_test.go":         {Data: []byte(strings.ReplaceAll(testFileContent, "pkg_name", "subdir"))},
				"subdir/another/README.md":    {Data: []byte("#readme?")},
				"subdir/another/file_test.go": {Data: []byte(strings.ReplaceAll(testFileContent, "pkg_name", "uniquePkgName"))},
			},
			dir: ".",
			expected: &Module{
				GoMod: &expectedGoMod,
				Packages: []*PackageAndTests{
					{
						Pkg: &Package{
							Name:           "pkg_name",
							Directory:      ".",
							TestFiles:      []string{"testfile_test.go"},
							OtherFiles:     []string{"go.mod"},
							Subdirectories: []string{"subdir"},
						},
						Tests: []*TestFile{{
							Name:       "testfile_test.go",
							Path:       "testfile_test.go",
							Tests:      expectedTestFile.Tests,
							Examples:   expectedTestFile.Examples,
							Benchmarks: expectedTestFile.Benchmarks,
						}},
					},
					{
						Pkg: &Package{
							Name:           "subdir",
							Directory:      "subdir",
							TestFiles:      []string{"file_test.go"},
							GoFiles:        []string{"file.go"},
							Subdirectories: []string{"another"},
						},
						Tests: []*TestFile{{
							Name:       "file_test.go",
							Path:       "subdir/file_test.go",
							Tests:      expectedTestFile.Tests,
							Examples:   expectedTestFile.Examples,
							Benchmarks: expectedTestFile.Benchmarks,
						}},
					},
					{
						Pkg: &Package{
							Name:       "uniquePkgName",
							Directory:  "subdir/another",
							TestFiles:  []string{"file_test.go"},
							OtherFiles: []string{"README.md"},
						},
						Tests: []*TestFile{{
							Name:       "file_test.go",
							Path:       "subdir/another/file_test.go",
							Tests:      expectedTestFile.Tests,
							Examples:   expectedTestFile.Examples,
							Benchmarks: expectedTestFile.Benchmarks,
						}},
					},
				},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			result, err := ParseModule(tc.fsys, tc.dir, gitignore.NewMatcher(nil))
			fmt.Println(tc.fsys)
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tc.expected, result)
		})
	}
}
