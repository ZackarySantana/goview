package stats

import (
	"bufio"
	"io"
	"strings"
)

type GoMod struct {
	ModuleName string
	GoVersion  string

	Dependencies []Dependency
	Tools        []Tool
}

type Dependency struct {
	Name, Version string
	Indirect      bool
}

type Tool struct {
	Name, Version string
}

// ParseGoMod extracts the module information from a
func ParseGoMod(r io.Reader) (*GoMod, error) {
	data := &GoMod{}
	insideRequireBlock := false

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			name := strings.TrimSpace(strings.TrimPrefix(line, "module "))
			data.ModuleName = name
			continue
		}
		if strings.HasPrefix(line, "go ") {
			version := strings.TrimSpace(strings.TrimPrefix(line, "go "))
			data.GoVersion = version
			continue
		}
		if strings.HasPrefix(line, "require (") {
			insideRequireBlock = true
			continue
		}
		if insideRequireBlock {
			if strings.HasPrefix(line, ")") {
				insideRequireBlock = false
				continue
			}
		}
		if insideRequireBlock || strings.HasPrefix(line, "require ") {
			// Clean the line from "require" keyword and trim spaces
			depLine := strings.TrimSpace(strings.TrimPrefix(line, "require"))
			fields := strings.Fields(depLine)
			dependency := Dependency{
				Name:    fields[0],
				Version: fields[1],
			}
			if strings.HasSuffix(depLine, "// indirect") {
				dependency.Indirect = true
			}
			data.Dependencies = append(data.Dependencies, dependency)
			continue
		}
		if strings.HasPrefix(line, "tool ") {
			tool := strings.TrimSpace(strings.TrimPrefix(line, "tool "))
			fields := strings.Fields(tool)
			data.Tools = append(data.Tools, Tool{
				Name:    fields[0],
				Version: fields[1],
			})
			continue
		}
	}
	return data, nil
}
