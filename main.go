package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "gitzip <commit1> <commit2> <zipfile>",
	Short: "Create a zip containing changed files between two commits",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		commit1 := args[0]
		commit2 := args[1]
		zipFile := args[2]

		// 1. Get list of changed files (excluding deleted ones)
		changedFiles, err := getChangedFiles(commit1, commit2)
		if err != nil {
			return fmt.Errorf("failed to get changed files: %w", err)
		}

		if len(changedFiles) == 0 {
			fmt.Println("No changed files between the specified commits.")
			return nil
		}

		// 2. Get repo root
		repoRoot, err := getRepoRoot()
		if err != nil {
			return fmt.Errorf("failed to get git repo root: %w", err)
		}

		// 3. Create zip
		if err := createZipFromFiles(repoRoot, changedFiles, zipFile); err != nil {
			return fmt.Errorf("failed to create zip: %w", err)
		}

		fmt.Printf("Created zip %q with %d files.\n", zipFile, len(changedFiles))
		return nil
	},
}

func getChangedFiles(commit1, commit2 string) ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", commit1, commit2, "--diff-filter=d")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(out), "\n")
	var files []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}

func getRepoRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func createZipFromFiles(repoRoot string, files []string, zipPath string) error {
	// Remove existing zip if any (like your PowerShell script does)
	if _, err := os.Stat(zipPath); err == nil {
		if err := os.Remove(zipPath); err != nil {
			return fmt.Errorf("failed to remove existing zip: %w", err)
		}
	}

	zipFile, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for _, relPath := range files {
		srcPath := filepath.Join(repoRoot, relPath)

		info, err := os.Stat(srcPath)
		if err != nil {
			// skip missing files (just in case)
			fmt.Fprintf(os.Stderr, "warning: cannot stat %s: %v\n", srcPath, err)
			continue
		}
		if info.IsDir() {
			// normally git diff --name-only won't return dirs, but just in case
			continue
		}

		// Use the repo-relative path inside the zip
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return fmt.Errorf("failed to create header for %s: %w", srcPath, err)
		}

		header.Name = filepath.ToSlash(relPath) // ensure forward slashes in zip
		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return fmt.Errorf("failed to create entry for %s: %w", srcPath, err)
		}

		file, err := os.Open(srcPath)
		if err != nil {
			return fmt.Errorf("failed to open %s: %w", srcPath, err)
		}

		if _, err := io.Copy(writer, file); err != nil {
			file.Close()
			return fmt.Errorf("failed to write %s to zip: %w", srcPath, err)
		}
		file.Close()
	}

	return nil
}
