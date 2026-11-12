package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var deletedFileFlag string

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "gitzip <commit1> <commit2> <zipfile>",
	Short: "Create a zip containing changed files between two commits directly from Git objects",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		commit1 := args[0]
		commit2 := args[1]
		zipFile := args[2]

		// 1. Get list of changed/added/modified files (excluding deleted ones)
		changedFiles, err := getChangedFiles(commit1, commit2)
		if err != nil {
			return fmt.Errorf("failed to get changed files: %w", err)
		}

		// 2. Get list of deleted files
		deletedFiles, err := getDeletedFiles(commit1, commit2)
		if err != nil {
			return fmt.Errorf("failed to get deleted files: %w", err)
		}

		if len(changedFiles) == 0 && len(deletedFiles) == 0 {
			fmt.Println("No changed or deleted files between the specified commits.")
			return nil
		}

		// 3. Process changed files into ZIP from git objects
		if len(changedFiles) > 0 {
			commitTime, err := getCommitTime(commit2)
			if err != nil {
				// fallback to current time if commit timestamp fails
				commitTime = time.Now()
			}

			if err := createZipFromGitObjects(commit2, changedFiles, zipFile, commitTime); err != nil {
				return fmt.Errorf("failed to create zip: %w", err)
			}
			fmt.Printf("Created zip %q with %d file(s) from commit %s.\n", zipFile, len(changedFiles), commit2)
		} else {
			fmt.Println("No modified or added files to zip.")
		}

		// 4. Handle deleted files
		if len(deletedFiles) > 0 {
			outDeletedPath := deletedFileFlag
			if outDeletedPath == "" {
				ext := filepath.Ext(zipFile)
				base := strings.TrimSuffix(zipFile, ext)
				outDeletedPath = base + "_deleted.txt"
			}

			content := strings.Join(deletedFiles, "\n") + "\n"
			if err := os.WriteFile(outDeletedPath, []byte(content), 0644); err != nil {
				return fmt.Errorf("failed to write deleted files list: %w", err)
			}

			fmt.Printf("\nFound %d deleted file(s):\n", len(deletedFiles))
			for _, f := range deletedFiles {
				fmt.Printf("  - %s\n", f)
			}
			fmt.Printf("List of deleted files saved to %q.\n", outDeletedPath)
		} else {
			fmt.Println("No files were deleted between the specified commits.")
		}

		return nil
	},
}

func init() {
	rootCmd.Flags().StringVarP(&deletedFileFlag, "deleted-file", "d", "", "Custom path for deleted files list (default: <zipfile>_deleted.txt)")
}

// getChangedFiles returns files modified, added, renamed, etc. (excluding deleted)
func getChangedFiles(commit1, commit2 string) ([]string, error) {
	return runGitDiff(commit1, commit2, "--diff-filter=d")
}

// getDeletedFiles returns only files that were deleted
func getDeletedFiles(commit1, commit2 string) ([]string, error) {
	return runGitDiff(commit1, commit2, "--diff-filter=D")
}

func runGitDiff(commit1, commit2, filter string) ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", filter, commit1, commit2)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	out, err := cmd.Output()
	if err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg != "" {
			return nil, fmt.Errorf("%s (%w)", errMsg, err)
		}
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

func getCommitTime(commit string) (time.Time, error) {
	cmd := exec.Command("git", "log", "-1", "--format=%ct", commit)
	out, err := cmd.Output()
	if err != nil {
		return time.Time{}, err
	}

	sec, err := strconv.ParseInt(strings.TrimSpace(string(out)), 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(sec, 0), nil
}

func createZipFromGitObjects(commit string, files []string, zipPath string, commitTime time.Time) error {
	// Remove existing zip if any
	if _, err := os.Stat(zipPath); err == nil {
		if err := os.Remove(zipPath); err != nil {
			return fmt.Errorf("failed to remove existing zip: %w", err)
		}
	}

	// Ensure destination directory exists if path contains directories
	dir := filepath.Dir(zipPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory for zip: %w", err)
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
		// Prepare zip header
		header := &zip.FileHeader{
			Name:     filepath.ToSlash(relPath),
			Method:   zip.Deflate,
			Modified: commitTime,
		}
		header.SetMode(0644)

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return fmt.Errorf("failed to create zip entry for %s: %w", relPath, err)
		}

		// Stream content directly from Git object: git cat-file blob <commit>:<path>
		gitObjectRef := fmt.Sprintf("%s:%s", commit, filepath.ToSlash(relPath))
		cmd := exec.Command("git", "cat-file", "blob", gitObjectRef)
		cmd.Stdout = writer

		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			errMsg := strings.TrimSpace(stderr.String())
			fmt.Fprintf(os.Stderr, "warning: cannot read git object %s (%s): %v\n", gitObjectRef, errMsg, err)
			continue
		}
	}

	return nil
}
