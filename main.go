package main

import (
	"log/slog"
	"os"

	"github.com/alecthomas/kong"
)

type CLI struct {
	Extract  Extract  `cmd:"" help:"Extract information from a set of images"`
	Organize Organize `cmd:"" help:"Organize a set of images"`
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))
	
	cli := &CLI{}
	ctx := kong.Parse(cli)
	// Call the Run() method of the selected parsed command.
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
