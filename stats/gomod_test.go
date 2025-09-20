package stats

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseGoMod(t *testing.T) {
	for name, tc := range map[string]struct {
		input    string
		expected GoMod
	}{
		"Parses": {
			input: `module github.com/user/project
			go 1.18

			require (
				github.com/some/dependency v1.2.3
				github.com/another/dependency v0.9.0 // indirect
			)
			
			tool github.com/some/tool v0.1.0`,
			expected: GoMod{
				ModuleName: "github.com/user/project",
				GoVersion:  "1.18",
				Dependencies: []Dependency{
					{Name: "github.com/some/dependency", Version: "v1.2.3", Indirect: false},
					{Name: "github.com/another/dependency", Version: "v0.9.0", Indirect: true},
				},
				Tools: []Tool{{Name: "github.com/some/tool", Version: "v0.1.0"}},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			result, err := ParseGoMod(strings.NewReader(tc.input))
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tc.expected, *result)
		})
	}
}
