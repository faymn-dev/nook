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
						Name:  "title",
						Value: "nook",
						Usage: "title of your website",
					},
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
					&cli.StringSliceFlag{
						Name:  "ignore",
						Usage: "folders and files to exclude from the build",
					},
					// TODO support metadata inside of the markdown files, which can override titles
					// TODO download theme command (literally go and fetch from GitHub)
					// TODO ignore files
				},
				Usage:  "generate static website",
				Action: generateContent,
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
