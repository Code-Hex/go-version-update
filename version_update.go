package update

import (
	"fmt"
	"io"
	"strings"

	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
)

// Version for go-version-update
const Version = "1.0.0"

// NextVersion will replace values of "version" attributes.
// fi:      Specify the output destination of formatted code.
// path:    target file.
func NextVersion(fi io.Writer, version, path string) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	for _, d := range f.Decls {
		td, ok := d.(*ast.GenDecl)
		if !ok {
			continue
		}

		if !(td.Tok == token.CONST || td.Tok == token.VAR) || len(td.Specs) == 0 {
			continue
		}

		val, ok := td.Specs[0].(*ast.ValueSpec)
		if !ok || len(val.Names) == 0 {
			continue
		}

		if strings.ToLower(val.Names[0].Name) != "version" {
			continue
		}

		lit, ok := val.Values[0].(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			continue
		}

		lit.Value = fmt.Sprintf(`"%s"`, version)
		if err := format.Node(fi, fset, f); err != nil {
			return err
		}
		break
	}

	return nil
}
