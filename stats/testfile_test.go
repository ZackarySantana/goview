package stats

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTestFile(t *testing.T) {
	for name, tc := range map[string]struct {
		input    string
		expected TestFile
	}{
		"Parses": {
			input: `package mypackage

			import "testing"

			func Testfake() {
				// not a real test function
			}

			func TestMyFunction(t *testing.T) {
				// test code
			}
				
			func ExampleMyFunction() {
				// example code
			}

			func BenchmarkMyFunction(b *testing.B) {
				// benchmark code
			}`,
			expected: TestFile{
				Tests:      []Test{{Name: "TestMyFunction"}},
				Examples:   []Example{{Name: "ExampleMyFunction"}},
				Benchmarks: []Benchmark{{Name: "BenchmarkMyFunction"}},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Run("ParseTestFile", func(t *testing.T) {
				result, err := ParseTestFile(strings.NewReader(tc.input))
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tc.expected, *result)
			})
		})
	}
}
