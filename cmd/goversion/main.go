package main

import (
	"fmt"
	"os"

	update "github.com/Code-Hex/go-version-update"
	"github.com/k0kubun/pp"
)

func main() {
	founds, err := update.GrepVersion("~")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	pp.Println(founds)
}
