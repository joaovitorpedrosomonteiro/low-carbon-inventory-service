package inventory

import (
	"fmt"

	"github.com/google/uuid"
)

type EmissionTemplate struct {
	ID             uuid.UUID   `json:"id"`
	Name           string      `json:"name"`
	Emissions      []*Emission `json:"emissions"`
	SupportingLinks []*Link    `json:"supporting_links"`
	InventoryCount int         `json:"inventory_count"`
	IsFrozen       bool        `json:"is_frozen"`
}

func NewEmissionTemplate(id uuid.UUID, name string, emissions []*Emission, supportingLinks []*Link) (*EmissionTemplate, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("id cannot be nil")
	}
	if name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}
	if emissions == nil {
		emissions = make([]*Emission, 0)
	}
	if supportingLinks == nil {
		supportingLinks = make([]*Link, 0)
	}

	return &EmissionTemplate{
		ID:             id,
		Name:           name,
		Emissions:      emissions,
		SupportingLinks: supportingLinks,
		InventoryCount: 0,
		IsFrozen:       false,
	}, nil
}

func (t *EmissionTemplate) Freeze() error {
	if t.InventoryCount > 0 {
		t.IsFrozen = true
	}
	return nil
}

func (t *EmissionTemplate) IncrementInventoryCount() {
	if !t.IsFrozen {
		t.InventoryCount++
		if t.InventoryCount > 0 {
			t.IsFrozen = true
		}
	}
}

func (t EmissionTemplate) CanModify() bool {
	return !t.IsFrozen
}