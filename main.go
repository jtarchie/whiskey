package main

import (
	"github.com/alecthomas/kong"
)

type CLI struct {
	Extract  Extract  `cmd:"" help:"Extract information from a set of images"`
	Organize Organize `cmd:"" help:"Organize a set of images"`
}

func main() {
	cli := &CLI{}
	ctx := kong.Parse(cli)
	// Call the Run() method of the selected parsed command.
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
