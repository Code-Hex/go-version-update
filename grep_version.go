package update

import (
	"context"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sync/errgroup"
)

type info struct {
	Path    string
	Version string
}

// GrepVersion will detect go files for "version" variables in the project.
func GrepVersion(basedir string) ([]*info, error) {
	path, err := filepath.Abs(basedir)
	if err != nil {
		return nil, err
	}

	return grepVersion(context.Background(), path)
}

func isMainGoFile(f os.FileInfo) bool {
	name := f.Name()
	if f.IsDir() || strings.HasSuffix(name, "_test.go") {
		return false
	}
	return !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".go")
}

func grepVersion(ctx context.Context, basedir string) ([]*info, error) {
	g, ctx := errgroup.WithContext(ctx)

	paths := make(chan string)
	g.Go(findGoFiles(ctx, basedir, paths))

	vpaths := make(chan *info)
	g.Go(detectVersionFile(ctx, paths, vpaths))

	vfiles := make([]*info, 0)
	for vpath := range vpaths {
		vfiles = append(vfiles, vpath)
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return vfiles, nil
}

func detectVersionFile(ctx context.Context, paths <-chan string, vpaths chan<- *info) func() error {
	return func() error {
		defer close(vpaths)
		for path := range paths {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				info, err := getVersionFile(path)
				if err != nil {
					if _, ok := err.(*notfound); !ok {
						return err
					}
				} else {
					vpaths <- info
				}
			}
		}
		return nil
	}
}

// walk in go projects.
// but do not search below these directories: .git, vendor, internal
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
				if isMainGoFile(info) {
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

func getVersionFile(path string) (*info, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return nil, err
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

		return &info{Path: path, Version: lit.Value}, nil
	}

	return nil, &notfound{err: errors.New("Not found")}
}

type notfound struct {
	err error
}

func (n *notfound) Error() string {
	return n.err.Error()
}
