package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

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
