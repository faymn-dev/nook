package main

import (
	"context"
	"github.com/urfave/cli/v3"
	"log"
	"os"
)

const numWorkers = 5
const defaultPermissions = 0755

func main() {
	cmd := &cli.Command{
		Name:           "nook",
		Usage:          "a minimal static site generator",
		DefaultCommand: "build",
		Commands: []*cli.Command{
			{
				Name: "build",
				Arguments: []cli.Argument{
					&cli.StringArg{
						Name:  "input-directory",
						Value: ".",
					},
				},
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "copy-markdown",
						Value: false,
						Usage: "copy markdown files to output directory (great for language models)",
					},
					&cli.BoolFlag{
						Name:  "clean",
						Value: false,
						Usage: "clean output directory",
					},
					&cli.StringFlag{
						Name:    "output-directory",
						Aliases: []string{"out"},
						Value:   "public",
						Usage:   "where to output generated content and copied assets",
					},
					// TODO support metadata inside of the markdown files, which can override h1
				},
				Usage:  "generate static website",
				Action: generateContentAction,
			},
			{
				Name:  "theme",
				Usage: "download preset nook themes",
				Arguments: []cli.Argument{
					&cli.StringArg{
						Name:  "name",
						Value: "classic.css",
					},
				},
				Action: downloadThemeAction,
				Commands: []*cli.Command{
					{
						Name:    "list",
						Aliases: []string{"ls"},
						Usage:   "list available themes",
						Action:  listThemesAction,
					},
				},
			},
			{
				Name:   "serve",
				Usage:  "serve the contents of the public directory",
				Action: servePublicAction,
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
