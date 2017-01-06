package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"runtime"

	update "github.com/Code-Hex/go-version-update"
	flags "github.com/jessevdk/go-flags"
)

// Options struct for parse command line arguments
type Options struct {
	Help    bool   `short:"h" long:"help"`
	Version bool   `short:"v" long:"version"`
	Rewrite string `short:"f" long:"format"`
	RelPath string `short:"d" long:"dir"`
}

func (opts *Options) parse(argv []string) error {
	p := flags.NewParser(opts, flags.PrintErrors)
	if _, err := p.ParseArgs(argv); err != nil {
		return opts.usage()
	}
	return nil
}

func (opts *Options) usage() error {
	if procs := os.Getenv("GOMAXPROCS"); procs == "" {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}
	fmt.Fprintf(os.Stdout,
		`Usage: goversion [options] 
  Options:
  -h,  --help                   print usage and exit
  -f,  --format <version>       rewrite version of code
  -d,  --dir                    target directory
`)
	return &ignore{err: errors.New("Hello")}
}

func run() error {
	opts := new(Options)
	if err := opts.parse(os.Args[1:]); err != nil {
		if _, ok := err.(*ignore); ok {
			return nil
		}
		return err
	}
	if opts.Help {
		opts.usage()
		return nil
	}
	if opts.Version {
		fmt.Printf("goversion: %s\n", update.Version)
		return nil
	}
	founds, err := update.GrepVersion(opts.RelPath)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if opts.Rewrite != "" {
		for _, info := range founds {
			fi, err := os.OpenFile(info.Path, os.O_WRONLY, 0644)
			if err != nil {
				return err
			}
			defer fi.Close()

			if err := update.NextVersion(fi, opts.Rewrite, info.Path); err != nil {
				return err
			}
			fmt.Fprintf(&buf, `Modified to "%s" from %s: %s`+"\n", opts.Rewrite, info.Version, info.Path)
		}
		os.Stdout.Write(buf.Bytes())
		return nil
	}

	for _, info := range founds {
		fmt.Fprintf(&buf, "%s: %s\n", info.Version, info.Path)
	}

	os.Stdout.Write(buf.Bytes())
	return nil
}

type ignore struct {
	err error
}

func (i *ignore) Error() string {
	return i.err.Error()
}

func main() {
	if err := run(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}
