package update

import (
	"context"
	"fmt"
	"io"
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

type Mode int

const (
	Stdout Mode = iota
	ReWrite
	FindGoGile
)

func NextVersion(version, basedir string) error {
	return nextVersion(context.Background(), version, basedir)
}

func nextVersion(ctx context.Context, version, basedir string) error {
	g, ctx := errgroup.WithContext(ctx)

	paths := make(chan string)
	g.Go(findGoFiles(ctx, basedir, paths))

	f := make(chan File)
	defer close(f)
	g.Go(updateVersion(ctx, version, paths))

	return g.Wait()
}

func findGoFiles(ctx context.Context, basedir string, paths chan<- string) func() error {
	return func() error {
		defer close(paths)
		return filepath.Walk(basedir, func(path string, info os.FileInfo, err error) error {
			if info == nil {
				return err
			}

			if info.IsDir() {
				dirname := info.Name()
				if dirname == ".git" || dirname == "vendor" || dirname == "internal" {
					return filepath.SkipDir
				}
			} else {
				if strings.HasSuffix(info.Name(), "_test.go") {
					return nil
				}

				if strings.HasSuffix(info.Name(), ".go") {
					select {
					case paths <- path:
					case <-ctx.Done():
						return ctx.Err()
					}
				}
			}

			return nil
		})
	}
}

func updateVersion(ctx context.Context, version string, paths <-chan string) func() error {
	// ここに列挙したパスのスライスを作成
	return func() error {
		for path := range paths {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				// Todo:
				//   - アップデートできるファイルの列挙 channel で送る?
				//   - ファイルへの書き込み
				if err := NextVersionGoFile(os.Stdout, version, path); err != nil {
					return err
				}
			}
		}
		return nil
	}
}

func NextVersionGoFile(output io.Writer, version, path string) error {
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
		if err := format.Node(output, fset, f); err != nil {
			return err
		}
	}

	return nil
}
