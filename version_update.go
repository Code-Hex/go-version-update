package update

import (
	"fmt"
	"os"
	"strings"

	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
)

const Version = "0.0.1"

type Mode int

const (
	Write Mode = iota
	Stdout
)

func NextVersion(mode Mode, version, path string) error {

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, 0)
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
		switch mode {
		case 0:
			fi, err := os.Open(path)
			if err != nil {
				return err
			}
			defer fi.Close()
			if err := format.Node(fi, fset, f); err != nil {
				return err
			}
		default:
			if err := format.Node(os.Stdout, fset, f); err != nil {
				return err
			}
		}
	}

	return nil
}
