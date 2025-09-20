package stats

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"strings"
)

type TestFile struct {
	// These fields are set by the caller.
	Name, Path string

	Tests      []Test
	Examples   []Example
	Benchmarks []Benchmark
}

type Test struct {
	Name string
}

type Example struct {
	Name string
}

type Benchmark struct {
	Name string
}

func ParseTestFile(r io.Reader) (*TestFile, error) {
	testFile := &TestFile{}

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", r, parser.AllErrors)
	if err != nil {
		return nil, err
	}

	for _, decl := range node.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		if functionHasSignature(fn, "Test", "T") {
			testFile.Tests = append(testFile.Tests, Test{Name: fn.Name.Name})
		} else if functionHasSignature(fn, "Example", "") {
			testFile.Examples = append(testFile.Examples, Example{Name: fn.Name.Name})
		} else if functionHasSignature(fn, "Benchmark", "B") {
			testFile.Benchmarks = append(testFile.Benchmarks, Benchmark{Name: fn.Name.Name})
		}
	}

	return testFile, nil
}

func functionHasSignature(fn *ast.FuncDecl, prefix string, paramType string) bool {
	if fn == nil || !strings.HasPrefix(fn.Name.Name, prefix) {
		return false
	}
	if paramType == "" {
		return fn.Type.Params == nil || len(fn.Type.Params.List) == 0
	}
	if fn.Type.Params == nil || len(fn.Type.Params.List) != 1 {
		return false
	}
	if starExpr, ok := fn.Type.Params.List[0].Type.(*ast.StarExpr); ok {
		if selExpr, ok := starExpr.X.(*ast.SelectorExpr); ok {
			return selExpr.Sel.Name == paramType
		}
	}
	return false
}
