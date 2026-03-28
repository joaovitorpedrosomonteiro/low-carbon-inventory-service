package inventory

import (
	"fmt"

	"github.com/google/uuid"
)

type Category struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Scope  *Scope    `json:"scope"`
	ScopeID uuid.UUID `json:"scope_id"`
}

func NewCategory(id uuid.UUID, name string, scope *Scope) (*Category, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("id cannot be nil")
	}
	if name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}
	if scope == nil {
		return nil, fmt.Errorf("scope cannot be nil")
	}

	return &Category{
		ID:     id,
		Name:   name,
		Scope:  scope,
		ScopeID: scope.ID,
	}, nil
}

func (c Category) IsValid() bool {
	return c.ID != uuid.Nil && c.Name != "" && c.Scope != nil
}