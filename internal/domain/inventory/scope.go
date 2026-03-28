package inventory

import (
	"fmt"

	"github.com/google/uuid"
)

type Scope struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

func NewScope(id uuid.UUID, name string) (*Scope, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("id cannot be nil")
	}
	if name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}

	return &Scope{
		ID:   id,
		Name: name,
	}, nil
}

func (s Scope) IsValid() bool {
	return s.ID != uuid.Nil && s.Name != ""
}