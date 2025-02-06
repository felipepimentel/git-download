package metadata

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// RepositoryMetadata represents a single repository's sync information
type RepositoryMetadata struct {
	Name        string    `json:"name"`
	URL         string    `json:"url"`
	Branch      string    `json:"branch"`
	RefType     string    `json:"refType"` // "branch" or "tag"
	LastSync    time.Time `json:"lastSync"`
	Destination string    `json:"destination"`
}

// SyncMetadata represents the root metadata structure
type SyncMetadata struct {
	Repositories []RepositoryMetadata `json:"repositories"`
}

// LoadMetadata loads the sync metadata from the specified file
func LoadMetadata(filename string) (*SyncMetadata, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return &SyncMetadata{Repositories: make([]RepositoryMetadata, 0)}, nil
		}
		return nil, err
	}

	var metadata SyncMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, err
	}

	return &metadata, nil
}

// SaveMetadata saves the sync metadata to the specified file
func (sm *SyncMetadata) SaveMetadata(filename string) error {
	data, err := json.MarshalIndent(sm, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

// AddRepository adds a new repository to the metadata
func (sm *SyncMetadata) AddRepository(repo RepositoryMetadata) {
	// Check if repository already exists
	for i, r := range sm.Repositories {
		if r.Name == repo.Name {
			sm.Repositories[i] = repo
			return
		}
	}
	sm.Repositories = append(sm.Repositories, repo)
}

// RemoveRepository removes a repository from the metadata
func (sm *SyncMetadata) RemoveRepository(name string) bool {
	for i, repo := range sm.Repositories {
		if repo.Name == name {
			sm.Repositories = append(sm.Repositories[:i], sm.Repositories[i+1:]...)
			return true
		}
	}
	return false
}

// GetRepository retrieves a repository by name
func (sm *SyncMetadata) GetRepository(name string) (*RepositoryMetadata, bool) {
	for _, repo := range sm.Repositories {
		if repo.Name == name {
			return &repo, true
		}
	}
	return nil, false
} 