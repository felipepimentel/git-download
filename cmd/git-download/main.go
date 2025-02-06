package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/pimentel/git-download/internal/metadata"
	"github.com/pimentel/git-download/internal/sync"
)

const metadataFile = ".syncmeta.json"

var rootCmd = &cobra.Command{
	Use:   "git-download",
	Short: "A tool to download and sync Git repositories without using git clone",
	Long: `git-download is a CLI tool that allows you to download and synchronize
public Git repositories using ZIP files instead of git clone. It maintains a metadata
file to track repositories and their sync status.`,
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new repository sync",
	Long:  `Initialize a new repository for syncing by providing the repository URL, branch/tag, and destination.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		url, _ := cmd.Flags().GetString("url")
		ref, _ := cmd.Flags().GetString("ref")
		refType, _ := cmd.Flags().GetString("ref-type")
		dest, _ := cmd.Flags().GetString("destination")
		name, _ := cmd.Flags().GetString("name")

		if url == "" {
			return fmt.Errorf("URL is required")
		}

		// Validate ref type
		if refType != "" && refType != "branch" && refType != "tag" {
			return fmt.Errorf("ref-type must be either 'branch' or 'tag'")
		}

		// If name is not provided, extract it from the URL
		if name == "" {
			name = filepath.Base(url)
		}

		// Load or create metadata
		meta, err := metadata.LoadMetadata(metadataFile)
		if err != nil {
			return fmt.Errorf("failed to load metadata: %w", err)
		}

		// Add repository to metadata
		repo := metadata.RepositoryMetadata{
			Name:        name,
			URL:         url,
			Branch:      ref,
			RefType:     refType,
			Destination: dest,
			LastSync:    time.Time{}, // Zero time indicates never synced
		}
		meta.AddRepository(repo)

		// Save metadata
		if err := meta.SaveMetadata(metadataFile); err != nil {
			return fmt.Errorf("failed to save metadata: %w", err)
		}

		fmt.Printf("Repository '%s' initialized successfully\n", name)
		return nil
	},
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize repositories",
	Long:  `Synchronize one or all tracked repositories by downloading and extracting their latest ZIP files.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")

		// Load metadata
		meta, err := metadata.LoadMetadata(metadataFile)
		if err != nil {
			return fmt.Errorf("failed to load metadata: %w", err)
		}

		if name != "" {
			// Sync specific repository
			repo, found := meta.GetRepository(name)
			if !found {
				return fmt.Errorf("repository '%s' not found", name)
			}
			return syncRepository(repo, meta)
		}

		// Sync all repositories
		for _, repo := range meta.Repositories {
			if err := syncRepository(&repo, meta); err != nil {
				fmt.Printf("Error syncing '%s': %v\n", repo.Name, err)
			}
		}

		return nil
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show sync status",
	Long:  `Display the sync status of all tracked repositories.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		meta, err := metadata.LoadMetadata(metadataFile)
		if err != nil {
			return fmt.Errorf("failed to load metadata: %w", err)
		}

		if len(meta.Repositories) == 0 {
			fmt.Println("No repositories are being tracked")
			return nil
		}

		fmt.Println("Tracked repositories:")
		for _, repo := range meta.Repositories {
			lastSync := "Never"
			if !repo.LastSync.IsZero() {
				lastSync = repo.LastSync.Format(time.RFC3339)
			}
			fmt.Printf("- %s:\n", repo.Name)
			fmt.Printf("  URL: %s\n", repo.URL)
			fmt.Printf("  %s: %s\n", repo.RefType, repo.Branch)
			fmt.Printf("  Destination: %s\n", repo.Destination)
			fmt.Printf("  Last Sync: %s\n\n", lastSync)
		}

		return nil
	},
}

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a repository",
	Long:  `Remove a repository from being tracked. Optionally delete local files.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		deleteLocal, _ := cmd.Flags().GetBool("delete-local")

		if name == "" {
			return fmt.Errorf("repository name is required")
		}

		meta, err := metadata.LoadMetadata(metadataFile)
		if err != nil {
			return fmt.Errorf("failed to load metadata: %w", err)
		}

		repo, found := meta.GetRepository(name)
		if !found {
			return fmt.Errorf("repository '%s' not found", name)
		}

		if deleteLocal && repo.Destination != "" {
			if err := os.RemoveAll(repo.Destination); err != nil {
				fmt.Printf("Warning: failed to delete local files: %v\n", err)
			}
		}

		if meta.RemoveRepository(name) {
			if err := meta.SaveMetadata(metadataFile); err != nil {
				return fmt.Errorf("failed to save metadata: %w", err)
			}
			fmt.Printf("Repository '%s' removed successfully\n", name)
		}

		return nil
	},
}

func syncRepository(repo *metadata.RepositoryMetadata, meta *metadata.SyncMetadata) error {
	fmt.Printf("Syncing '%s'...\n", repo.Name)

	// Download ZIP file
	zipFile, err := sync.DownloadZIP(repo.URL, repo.Branch, repo.RefType)
	if err != nil {
		return fmt.Errorf("failed to download repository: %w", err)
	}
	defer sync.Cleanup(zipFile.Name())

	// Extract ZIP file
	if err := sync.ExtractZIP(zipFile, repo.Destination); err != nil {
		return fmt.Errorf("failed to extract repository: %w", err)
	}

	// Update last sync time
	repo.LastSync = time.Now()
	meta.AddRepository(*repo)
	if err := meta.SaveMetadata(metadataFile); err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	fmt.Printf("Successfully synced '%s'\n", repo.Name)
	return nil
}

func init() {
	// Init command flags
	initCmd.Flags().String("url", "", "Repository URL (required)")
	initCmd.Flags().String("ref", "main", "Repository reference (branch or tag)")
	initCmd.Flags().String("ref-type", "branch", "Reference type ('branch' or 'tag')")
	initCmd.Flags().String("destination", "", "Local destination path")
	initCmd.Flags().String("name", "", "Repository name (defaults to URL basename)")

	// Sync command flags
	syncCmd.Flags().String("name", "", "Repository name to sync (if empty, syncs all)")

	// Remove command flags
	removeCmd.Flags().String("name", "", "Repository name to remove (required)")
	removeCmd.Flags().Bool("delete-local", false, "Delete local files")

	// Add commands to root
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(removeCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
} 