package inventory

import (
	"fmt"

	"github.com/google/uuid"
)

type StorageType string

const (
	StorageLocal StorageType = "local"
	StorageGCS   StorageType = "gcs"
)

type Link struct {
	ID          uuid.UUID   `json:"id"`
	Name        string      `json:"name"`
	Path        string      `json:"path"`
	StorageType StorageType `json:"storage_type"`
}

func NewLink(id uuid.UUID, name, path string, storageType StorageType) (*Link, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("id cannot be nil")
	}
	if name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}
	if path == "" {
		return nil, fmt.Errorf("path cannot be empty")
	}
	if storageType != StorageLocal && storageType != StorageGCS {
		return nil, fmt.Errorf("invalid storage type")
	}

	return &Link{
		ID:          id,
		Name:        name,
		Path:        path,
		StorageType: storageType,
	}, nil
}

func (l Link) IsValid() bool {
	return l.ID != uuid.Nil && l.Name != "" && l.Path != "" &&
		(l.StorageType == StorageLocal || l.StorageType == StorageGCS)
}