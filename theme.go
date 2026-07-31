package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/urfave/cli/v3"
)

func downloadThemeAction(ctx context.Context, cmd *cli.Command) error {
	themes, err := listThemes(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch themes: %w", err)
	}

	themeName := cmd.StringArg("name")
	themeIndex := -1
	for i, theme := range themes {
		if theme.Name == themeName {
			themeIndex = i
			break
		}
	}

	if themeIndex == -1 {
		return fmt.Errorf("theme %q does not exist", themeName)
	}

	client := &http.Client{}

	req, err := http.NewRequestWithContext(ctx, "GET", themes[themeIndex].DownloadUrl, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("response has bad status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	err = os.WriteFile("theme.css", body, defaultPermissions)
	if err != nil {
		return fmt.Errorf("failed to save theme: %w", err)
	}

	return nil
}

type ThemeResponse []struct {
	Name        string `json:"name"`
	DownloadUrl string `json:"download_url"`
}

func listThemes(ctx context.Context) (ThemeResponse, error) {
	client := &http.Client{}

	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/repos/faymn-dev/nook/contents/themes?ref=main", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Add("X-GitHub-Api-Version", "2026-03-10")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("response has bad status code: %d", resp.StatusCode)
	}

	var themes ThemeResponse
	if err := json.NewDecoder(resp.Body).Decode(&themes); err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	return themes, nil
}

func listThemesAction(ctx context.Context, _ *cli.Command) error {
	themes, err := listThemes(ctx)
	if err != nil {
		return err
	}

	for _, theme := range themes {
		fmt.Println(theme.Name)
	}

	return nil
}
