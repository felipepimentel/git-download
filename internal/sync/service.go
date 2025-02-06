package sync

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pimentel/git-download/internal/metadata"
)

// Service handles repository synchronization operations
type Service struct {
	metadataFile string
}

// NewService creates a new sync service
func NewService(metadataFile string) *Service {
	return &Service{
		metadataFile: metadataFile,
	}
}

// SyncRepository synchronizes a single repository
func (s *Service) SyncRepository(repo *metadata.RepositoryMetadata) error {
	fmt.Printf("Syncing '%s'...\n", repo.Name)

	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(repo.Destination, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Download ZIP file
	zipFile, err := DownloadZIP(repo.URL, repo.Branch, repo.RefType)
	if err != nil {
		return fmt.Errorf("failed to download repository: %w", err)
	}
	defer Cleanup(zipFile.Name())

	// Extract ZIP file
	if err := ExtractZIP(zipFile, repo.Destination); err != nil {
		return fmt.Errorf("failed to extract repository: %w", err)
	}

	// Update last sync time
	repo.LastSync = time.Now()

	fmt.Printf("Successfully synced '%s'\n", repo.Name)
	return nil
}

// SyncAll synchronizes all repositories in the metadata file
func (s *Service) SyncAll() error {
	meta, err := metadata.LoadMetadata(s.metadataFile)
	if err != nil {
		return fmt.Errorf("failed to load metadata: %w", err)
	}

	for i := range meta.Repositories {
		if err := s.SyncRepository(&meta.Repositories[i]); err != nil {
			fmt.Printf("Error syncing '%s': %v\n", meta.Repositories[i].Name, err)
			continue
		}

		// Update metadata after successful sync
		meta.AddRepository(meta.Repositories[i])
		if err := meta.SaveMetadata(s.metadataFile); err != nil {
			fmt.Printf("Warning: failed to update metadata for '%s': %v\n", meta.Repositories[i].Name, err)
		}
	}

	return nil
}

// GetRepositoryByName retrieves a repository by its name
func (s *Service) GetRepositoryByName(name string) (*metadata.RepositoryMetadata, error) {
	meta, err := metadata.LoadMetadata(s.metadataFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load metadata: %w", err)
	}

	repo, found := meta.GetRepository(name)
	if !found {
		return nil, fmt.Errorf("repository '%s' not found", name)
	}

	return repo, nil
}

// RemoveRepository removes a repository and optionally its local files
func (s *Service) RemoveRepository(name string, deleteLocal bool) error {
	meta, err := metadata.LoadMetadata(s.metadataFile)
	if err != nil {
		return fmt.Errorf("failed to load metadata: %w", err)
	}

	repo, found := meta.GetRepository(name)
	if !found {
		return fmt.Errorf("repository '%s' not found", name)
	}

	if deleteLocal && repo.Destination != "" {
		if err := os.RemoveAll(repo.Destination); err != nil {
			return fmt.Errorf("failed to delete local files: %w", err)
		}
	}

	if !meta.RemoveRepository(name) {
		return fmt.Errorf("failed to remove repository from metadata")
	}

	if err := meta.SaveMetadata(s.metadataFile); err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	return nil
}

// AddRepository adds a new repository to be tracked
func (s *Service) AddRepository(repo metadata.RepositoryMetadata) error {
	meta, err := metadata.LoadMetadata(s.metadataFile)
	if err != nil {
		return fmt.Errorf("failed to load metadata: %w", err)
	}

	// Set default values if not specified
	if repo.Branch == "" {
		repo.Branch = "main"
	}
	if repo.RefType == "" {
		repo.RefType = "branch"
	}
	if repo.Destination == "" {
		repo.Destination = filepath.Join(".", repo.Name)
	}

	meta.AddRepository(repo)

	if err := meta.SaveMetadata(s.metadataFile); err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	return nil
} 