# go-version-update
[![Go Report Card](https://goreportcard.com/badge/github.com/Code-Hex/go-version-update)](https://goreportcard.com/report/github.com/Code-Hex/go-version-update) [![GoDoc](https://godoc.org/github.com/Code-Hex/go-version-update?status.svg)](https://godoc.org/github.com/Code-Hex/go-version-update)  
Update the version string of the go project.  
Rewrite the value of the `const` or `var` variables matching `/version/i` to the specified version string.

# Synopsis

```go
package main

import update "github.com/Code-Hex/go-version-update"

func main() {

    // Find go files are described version variables.
    founds, err := update.GrepVersion(opts.RelPath)
    if err != nil {
        panic(err)
    }

    // Re write value of version variables.
    for _, path := range founds {
        // Re write to "1.2.3"
        if err := update.NextVersion(os.Stdout, "1.2.3", path); err != nil {
            panic(err)
        }
    }
}
```

You can also modify the version of your project using the `goversion` command.

## goversion
```
Usage: goversion [options] 
  Options:
  -h,  --help                   print usage and exit
  -f,  --format <version>       rewrite version of code
  -d,  --dir                    target directory
```

# Installation
- Library install

    go get github.com/Code-Hex/go-version-update

- `goversion` command install

    go get github.com/Code-Hex/go-version-update/cmd/goversion

# License
[MIT](https://github.com/Code-Hex/go-version-update/blob/master/LICENSE)
# Author
[codehex](https://twitter.com/CodeHex)