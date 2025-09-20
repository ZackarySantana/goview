package stats

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var pkgContents1 = fstest.MapFS{
	"pkg/foo.go": {Data: []byte("package pkg_name // A comment \n func Foo() {}")},
	"pkg/foo_test.go": {Data: []byte(`package pkg_name
		import "testing"
		func TestFoo(t *testing.T) {}
		func ExampleFoo() {}
		func BenchmarkFoo(b *testing.B) {}
	`)},
	"pkg/README.md":  {Data: []byte("# Just a readme")},
	"README.md":      {Data: []byte("# Root readme")},
	"other/file.txt": {Data: []byte("Not a Go file.")},
}

var expectedPkg1 = Package{
	Name:       "pkg_name",
	Directory:  "pkg",
	TestFiles:  []string{"foo_test.go"},
	GoFiles:    []string{"foo.go"},
	OtherFiles: []string{"README.md"},
}

var pkgContents2 = fstest.MapFS{
	"dir/a.go":             {Data: []byte("package dir\n func A() {}")},
	"dir/b_test.go":        {Data: []byte("package dir\n func TestB(t *testing.T) {}")},
	"dir/subdir/c.go":      {Data: []byte("package subdir\n func C() {}")},
	"dir/subdir/d_test.go": {Data: []byte("package subdir\n func TestD(t *testing.T) {}")},
	"dir/notes.txt":        {Data: []byte("Some notes.")},
}

var expectedPkg2 = Package{
	Name:           "dir",
	Directory:      "dir",
	TestFiles:      []string{"b_test.go"},
	GoFiles:        []string{"a.go"},
	OtherFiles:     []string{"notes.txt"},
	Subdirectories: []string{"subdir"},
}

func TestParsePackage(t *testing.T) {
	for name, tc := range map[string]struct {
		fsys     fstest.MapFS
		dir      string
		expected Package
	}{
		"Parses": {
			fsys:     pkgContents1,
			dir:      "pkg",
			expected: expectedPkg1,
		},
		"ParsesAndTrimsPackageSuffixOf_test": {
			fsys:     pkgContents2,
			dir:      "dir",
			expected: expectedPkg2,
		},
	} {
		t.Run(name, func(t *testing.T) {
			result, err := ParsePackage(tc.fsys, tc.dir)
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tc.expected, *result)
		})
	}
}
