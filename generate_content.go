package main

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/faymn-dev/nook/internals/language"
	"github.com/faymn-dev/nook/internals/node"
	"github.com/urfave/cli/v3"
)

func generateContent(_ context.Context, cmd *cli.Command) error {
	inputDir := cmd.StringArg("input-directory")
	outputDir := cmd.String("output-directory")

	title := cmd.String("title")

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
	styles := []string{}

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

		ext := filepath.Ext(d.Name())
		if ext == ".md" {
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

		if ext == ".css" {
			styles = append(styles, p)
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

	head := createDocumentHead(title, styles)

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

				err = os.WriteFile(replaceExtension(item.outputFile, ".html"), renderDocument(head, string(markdown)), permission)
				if err != nil {
					log.Printf("failed to write %s: %v", item.inputFile, err)
					return
				}
			}
		})
	}

	wg.Wait()

	log.Printf("finished building %s\n", title)
	return nil
}

type queueItem struct {
	inputFile  string
	outputFile string
}

func renderDocument(head []node.Renderer, markdown string) []byte {
	body, _ := language.Parse(language.Tokenize(markdown))
	return []byte(node.NewDocument(head, []node.Renderer{body}).ToHTML())
}

func createDocumentHead(title string, styles []string) []node.Renderer {
	head := []node.Renderer{}
	head = append(head, node.NewHTMLNode("title", nil, node.TextNode(title)))
	for _, style := range styles {
		head = append(head, node.NewHTMLNode("link", node.HTMLProps{"rel": "stylesheet", "href": "/" + style}))
	}
	return head
}

func replaceExtension(src string, newExt string) string {
	ext := filepath.Ext(src)
	return src[:len(src)-len(ext)] + newExt
}

func copyFile(src string, dest string) error {
	source, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer source.Close()

	destination, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	if err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	err = destination.Sync()
	if err != nil {
		return fmt.Errorf("failed to sync destination file: %w", err)
	}

	return nil
}

// adapted from
// https://stackoverflow.com/questions/33450980/how-to-remove-all-contents-of-a-directory-using-golang
func removeContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()

	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}

	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}

	return nil
}
