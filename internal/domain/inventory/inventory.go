package inventory

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/joaovitorpedrosomonteiro/low-carbon-inventory-service/internal/domain/valueobject"
	"github.com/shopspring/decimal"
)

var (
	ErrInvalidID            = errors.New("id cannot be nil")
	ErrInvalidName          = errors.New("name cannot be empty")
	ErrInvalidMonth         = errors.New("invalid month")
	ErrInvalidYear          = errors.New("invalid year")
	ErrStateTransition      = errors.New("invalid state transition")
	ErrEmissionNotComplete  = errors.New("cannot transition to for_auditing: not all emissions are complete")
	ErrImmutableField       = errors.New("field is immutable after creation")
	ErrVersionConflict      = errors.New("version conflict: stale version")
	ErrAuditedIsTerminal    = errors.New("audited state is terminal")
	ErrNotEditableState     = errors.New("cannot edit variables in current state")
)

type Inventory struct {
	ID              uuid.UUID                  `json:"id"`
	Name            string                     `json:"name"`
	Month           valueobject.Month          `json:"month"`
	Year            valueobject.Year           `json:"year"`
	State           valueobject.InventoryState `json:"state"`
	Emissions       []*Emission                `json:"emissions"`
	TemplateID      *uuid.UUID                 `json:"template_id"`
	CompanyBranchID uuid.UUID                  `json:"company_branch_id"`
	GWPStandardID   *uuid.UUID                 `json:"gwp_standard_id"`
	ReviewMessage   *string                    `json:"review_message"`
	Version         int                        `json:"version"`
}

func NewInventory(id uuid.UUID, name string, month valueobject.Month, year valueobject.Year, companyBranchID uuid.UUID, templateID *uuid.UUID, gwpStandardID *uuid.UUID) (*Inventory, error) {
	if id == uuid.Nil {
		return nil, ErrInvalidID
	}
	if name == "" {
		return nil, ErrInvalidName
	}
	if !month.IsValid() {
		return nil, ErrInvalidMonth
	}
	if !year.IsValid() {
		return nil, ErrInvalidYear
	}
	if companyBranchID == uuid.Nil {
		return nil, fmt.Errorf("company branch id cannot be nil")
	}

	return &Inventory{
		ID:              id,
		Name:            name,
		Month:           month,
		Year:            year,
		State:           valueobject.ToReportEmissions,
		Emissions:       make([]*Emission, 0),
		TemplateID:      templateID,
		CompanyBranchID: companyBranchID,
		GWPStandardID:   gwpStandardID,
		ReviewMessage:   nil,
		Version:         1,
	}, nil
}

func (inv *Inventory) TransitionTo(newState valueobject.InventoryState, actorID uuid.UUID, expectedVersion int) error {
	if inv.Version != expectedVersion {
		return ErrVersionConflict
	}

	if inv.State.IsTerminal() {
		return ErrAuditedIsTerminal
	}

	if !inv.State.CanTransitionTo(newState) {
		return fmt.Errorf("%w: from %s to %s", ErrStateTransition, inv.State.String(), newState.String())
	}

	if newState == valueobject.ForAuditing && !inv.AllEmissionsComplete() {
		return ErrEmissionNotComplete
	}

	inv.State = newState
	inv.Version++

	return nil
}

func (inv *Inventory) AllEmissionsComplete() bool {
	for _, e := range inv.Emissions {
		if !e.IsComplete() {
			return false
		}
	}
	return true
}

func (inv *Inventory) FillVariables(emissionID uuid.UUID, variables map[string]decimal.Decimal, expectedVersion int) error {
	if inv.Version != expectedVersion {
		return ErrVersionConflict
	}

	if !inv.State.IsEditable() {
		return ErrNotEditableState
	}

	var emission *Emission
	for _, e := range inv.Emissions {
		if e.ID == emissionID {
			emission = e
			break
		}
	}

	if emission == nil {
		return fmt.Errorf("emission not found: %s", emissionID)
	}

	for name, value := range variables {
		emission.SetVariable(name, value)
	}

	inv.Version++

	return nil
}

func (inv *Inventory) AddEvidence(emissionID uuid.UUID, link *Link, expectedVersion int) error {
	if inv.Version != expectedVersion {
		return ErrVersionConflict
	}

	var emission *Emission
	for _, e := range inv.Emissions {
		if e.ID == emissionID {
			emission = e
			break
		}
	}

	if emission == nil {
		return fmt.Errorf("emission not found: %s", emissionID)
	}

	return emission.AddEvidence(link)
}

func (inv *Inventory) SetReviewMessage(message string, expectedVersion int) error {
	if inv.Version != expectedVersion {
		return ErrVersionConflict
	}

	inv.ReviewMessage = &message
	inv.Version++

	return nil
}

func (inv *Inventory) StoreReliabilityJobID(emissionID uuid.UUID, jobID string, expectedVersion int) error {
	if inv.Version != expectedVersion {
		return ErrVersionConflict
	}

	var emission *Emission
	for _, e := range inv.Emissions {
		if e.ID == emissionID {
			emission = e
			break
		}
	}

	if emission == nil {
		return fmt.Errorf("emission not found: %s", emissionID)
	}

	emission.ReliabilityJobID = &jobID
	inv.Version++

	return nil
}

func (inv *Inventory) AddEmission(emission *Emission) {
	inv.Emissions = append(inv.Emissions, emission)
}

func (inv Inventory) IsUniqueConstraintViolated(other *Inventory) bool {
	return inv.CompanyBranchID == other.CompanyBranchID &&
		inv.Month == other.Month &&
		inv.Year == other.Year &&
		inv.ID != other.ID
}