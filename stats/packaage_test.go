package stats

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePackage(t *testing.T) {
	for name, tc := range map[string]struct {
		fsys     fstest.MapFS
		dir      string
		expected Package
	}{
		"Parses": {
			fsys: fstest.MapFS{
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
			},
			dir: "pkg",
			expected: Package{
				Name:       "pkg_name",
				Directory:  "pkg",
				TestFiles:  []string{"foo_test.go"},
				GoFiles:    []string{"foo.go"},
				OtherFiles: []string{"README.md"},
			},
		},
		"ParsesAndTrimsPackageSuffixOf_test": {
			fsys: fstest.MapFS{
				"dir/x_test.go": {Data: []byte(`package mypkg_test`)},
			},
			dir: "dir",
			expected: Package{
				Name:      "mypkg",
				Directory: "dir",
				TestFiles: []string{"x_test.go"},
			},
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
