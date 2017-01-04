package modifier

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go/ast"
	"go/format"
	"go/parser"
	"go/token"

	"golang.org/x/sync/errgroup"
)

type File struct {
	src  string
	path string
}

func NextVersion(version, dir string) {
	if err := changeVersion(context.Background(), version, dir); err != nil {
		panic(err)
	}
}

func changeVersion(ctx context.Context, version, root string) error {
	g, ctx := errgroup.WithContext(ctx)

	paths := make(chan string)
	g.Go(func() error {
		defer close(paths)
		return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if info == nil {
				return err
			}

			if info.IsDir() {
				dirname := info.Name()
				if dirname == ".git" || dirname == "vendor" || dirname == "internal" {
					return filepath.SkipDir
				}
			} else {
				if strings.HasSuffix(info.Name(), ".go") {
					select {
					case paths <- path:
						fmt.Println(path)
					case <-ctx.Done():
						return ctx.Err()
					}
				}
			}

			return nil
		})
	})

	f := make(chan File)
	defer close(f)
	g.Go(modifyVersion(ctx, version, paths))

	return g.Wait()
}

func modifyVersion(ctx context.Context, version string, paths chan string) func() error {
	return func() error {
		fset := token.NewFileSet()
		for path := range paths {

			f, err := parser.ParseFile(fset, path, nil, 0)
			if err != nil {
				return err
			}

			for _, d := range f.Decls {
				if td, ok := d.(*ast.GenDecl); ok {
					if (td.Tok == token.CONST || td.Tok == token.VAR) && len(td.Specs) > 0 {
						if val, ok := td.Specs[0].(*ast.ValueSpec); ok && len(val.Names) > 0 {
							ident := strings.ToLower(val.Names[0].Name)
							if ident == "version" {
								if lit, ok := val.Values[0].(*ast.BasicLit); ok {
									if lit.Kind == token.STRING {
										lit.Value = fmt.Sprintf(`"%s"`, version)
										if err := format.Node(os.Stdout, fset, f); err != nil {
											return err
										}
									}
								}
							}
						}
					}
				}
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
		}
		return nil
	}
}
