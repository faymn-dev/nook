package main

import (
	"context"
	"github.com/urfave/cli/v3"
	"log"
	"os"
)

const numWorkers = 5
const permission = 0755

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
					&cli.StringFlag{
						Name:    "output-directory",
						Aliases: []string{"out"},
						Value:   "public",
						Usage:   "where to output generated content and copied assets",
					},
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
					// TODO blockquotes
					// TODO support metadata inside of the markdown files, which can override h1
					// TODO download theme command (literally go and fetch latest from GitHub)
				},
				Usage:  "generate static website",
				Action: generateContent,
			},
			{
				Name:  "theme",
				Usage: "download preset nook themes",
				Arguments: []cli.Argument{
					&cli.StringArg{
						Name:  "name",
						Value: "classic",
					},
				},
				Action: downloadTheme,
				Commands: []*cli.Command{
					{
						Name:    "list",
						Aliases: []string{"ls"},
						Usage:   "list available themes",
						Action:  listThemes,
					},
				},
			},
			{
				Name:   "serve",
				Usage:  "serve the contents of the public directory",
				Action: servePublic,
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
