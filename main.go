package main

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/faymn-dev/nook/internals/language"
	"github.com/urfave/cli/v3"
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
						Name:  "clean",
						Value: false,
						Usage: "clean output directory",
					},
					&cli.BoolFlag{
						Name:    "copy-markdown",
						Aliases: []string{"cp"},
						Value:   false,
						Usage:   "copy markdown files to output directory (great for language models)",
					},
				},
				Usage:  "generate static content",
				Action: generateStaticContent,
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

type queueItem struct {
	inputFile  string
	outputFile string
}

func generateStaticContent(_ context.Context, cmd *cli.Command) error {
	inputDir := cmd.StringArg("input-directory")
	outputDir := cmd.String("output-directory")

	if cmd.Bool("clean") {
		if err := removeContents(outputDir); err != nil {
			return fmt.Errorf("failed to clean output directory: %w", err)
		}
	}

	// ensure output directory exists
	err := os.MkdirAll(outputDir, permission)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	queue := []queueItem{}

	err = filepath.WalkDir(inputDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if p == inputDir { // skip input directory
			return nil
		}

		if strings.HasPrefix(p, ".") || strings.HasPrefix(p, outputDir) {
			return nil
		}

		// get path relative to input dir
		rel, err := filepath.Rel(inputDir, p)
		if err != nil {
			return err
		}
		outputPath := filepath.Join(outputDir, rel)

		if d.IsDir() { // "copy" directory, but not its contents
			return os.MkdirAll(outputPath, permission)
		}

		if filepath.Ext(d.Name()) == ".md" {
			if cmd.Bool("copy-markdown") {
				err := copyFile(p, outputPath)
				if err != nil {
					return err
				}
			}

			queue = append(queue, queueItem{
				inputFile:  p,
				outputFile: outputPath,
			})
			return nil
		}

		// otherwise copy a static asset
		err = copyFile(p, outputPath)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	c := make(chan queueItem, len(queue))
	for _, item := range queue {
		c <- item
	}
	close(c)

	for range numWorkers {
		wg.Go(func() {
			for item := range c {
				markdown, err := os.ReadFile(item.inputFile)
				if err != nil {
					log.Printf("failed to read %s: %v", item.inputFile, err)
					return
				}

				html := []byte(language.RenderDocument(string(markdown)))
				err = os.WriteFile(replaceExtension(item.outputFile, ".html"), html, permission)
				if err != nil {
					log.Printf("failed to write %s: %v", item.inputFile, err)
					return
				}
			}
		})
	}

	wg.Wait()

	return nil
}

func servePublic(context.Context, *cli.Command) error {
	mux := http.NewServeMux()

	mux.Handle("/", http.FileServer(http.Dir("public")))

	server := http.Server{
		Addr:    ":8888",
		Handler: mux,
	}

	return server.ListenAndServe()
}
