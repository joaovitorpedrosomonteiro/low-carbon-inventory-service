package command

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/joaovitorpedrosomonteiro/low-carbon-inventory-service/internal/domain/inventory"
	"github.com/joaovitorpedrosomonteiro/low-carbon-inventory-service/internal/domain/valueobject"
	"github.com/shopspring/decimal"
)

type CreateInventory struct {
	Name            string    `json:"name"`
	Month           int       `json:"month"`
	Year            int       `json:"year"`
	CompanyBranchID uuid.UUID `json:"company_branch_id"`
	TemplateID      *uuid.UUID `json:"template_id"`
	GWPStandardID   *uuid.UUID `json:"gwp_standard_id"`
}

type TransitionState struct {
	InventoryID   uuid.UUID   `json:"inventory_id"`
	ToState       string      `json:"to_state"`
	ActorID       uuid.UUID   `json:"actor_id"`
	ReviewMessage *string     `json:"review_message"`
	Version       int         `json:"version"`
}

type FillVariables struct {
	InventoryID uuid.UUID              `json:"inventory_id"`
	EmissionID  uuid.UUID              `json:"emission_id"`
	Variables   map[string]interface{} `json:"variables"`
	Version     int                    `json:"version"`
}

type AddEvidence struct {
	InventoryID  uuid.UUID `json:"inventory_id"`
	EmissionID   uuid.UUID `json:"emission_id"`
	Name         string    `json:"name"`
	Path         string    `json:"path"`
	StorageType  string    `json:"storage_type"`
	Version      int       `json:"version"`
}

type CreateTemplate struct {
	Name            string `json:"name"`
	EmissionIDs     []uuid.UUID `json:"emission_ids"`
	SupportingLinks []LinkDTO  `json:"supporting_links"`
}

type LinkDTO struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	StorageType string `json:"storage_type"`
}

type StoreReliabilityJob struct {
	InventoryID  uuid.UUID `json:"inventory_id"`
	EmissionID   uuid.UUID `json:"emission_id"`
	ReliabilityJobID string `json:"reliability_job_id"`
	Version      int       `json:"version"`
}

type InventoryCommandHandler struct {
	repo        inventory.InventoryRepository
	templateRepo inventory.EmissionTemplateRepository
	publisher   Publisher
}

type Publisher interface {
	Publish(ctx context.Context, topic string, data interface{}) error
}

func NewInventoryCommandHandler(repo inventory.InventoryRepository, templateRepo inventory.EmissionTemplateRepository, publisher Publisher) *InventoryCommandHandler {
	return &InventoryCommandHandler{
		repo:          repo,
		templateRepo: templateRepo,
		publisher:    publisher,
	}
}

func (h *InventoryCommandHandler) HandleCreateInventory(ctx context.Context, cmd CreateInventory) (*inventory.Inventory, error) {
	month, err := valueobject.NewMonth(cmd.Month)
	if err != nil {
		return nil, fmt.Errorf("invalid month: %w", err)
	}

	year, err := valueobject.NewYear(cmd.Year)
	if err != nil {
		return nil, fmt.Errorf("invalid year: %w", err)
	}

	existing, err := h.repo.GetByCompanyBranchAndPeriod(ctx, cmd.CompanyBranchID, cmd.Month, cmd.Year)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("inventory already exists for this company branch and period")
	}

	inv, err := inventory.NewInventory(
		uuid.New(),
		cmd.Name,
		month,
		year,
		cmd.CompanyBranchID,
		cmd.TemplateID,
		cmd.GWPStandardID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create inventory: %w", err)
	}

	if err := h.repo.Create(ctx, inv); err != nil {
		return nil, fmt.Errorf("failed to save inventory: %w", err)
	}

	event := inventory.InventoryCreated{
		InventoryID:    inv.ID,
		CompanyBranchID: cmd.CompanyBranchID,
		Month:          cmd.Month,
		Year:           cmd.Year,
		TemplateID:     cmd.TemplateID,
	}
	h.publisher.Publish(ctx, "inventory.created", event)

	return inv, nil
}

func (h *InventoryCommandHandler) HandleTransitionState(ctx context.Context, cmd TransitionState) (*inventory.Inventory, error) {
	inv, err := h.repo.GetByID(ctx, cmd.InventoryID)
	if err != nil {
		return nil, fmt.Errorf("inventory not found: %w", err)
	}

	newState, err := valueobject.NewInventoryState(cmd.ToState)
	if err != nil {
		return nil, fmt.Errorf("invalid state: %w", err)
	}

	fromState := inv.State.String()

	if err := inv.TransitionTo(newState, cmd.ActorID, cmd.Version); err != nil {
		return nil, fmt.Errorf("failed to transition state: %w", err)
	}

	if err := h.repo.Update(ctx, inv); err != nil {
		return nil, fmt.Errorf("failed to update inventory: %w", err)
	}

	event := inventory.InventoryStateChanged{
		InventoryID:   inv.ID,
		FromState:     fromState,
		ToState:       cmd.ToState,
		ActorID:       cmd.ActorID,
		ReviewMessage: cmd.ReviewMessage,
	}
	h.publisher.Publish(ctx, "inventory.state_changed", event)

	return inv, nil
}

func (h *InventoryCommandHandler) HandleFillVariables(ctx context.Context, cmd FillVariables) (*inventory.Inventory, error) {
	inv, err := h.repo.GetByID(ctx, cmd.InventoryID)
	if err != nil {
		return nil, fmt.Errorf("inventory not found: %w", err)
	}

	variables := make(map[string]decimal.Decimal)
	for k, v := range cmd.Variables {
		var dec decimal.Decimal
		switch val := v.(type) {
		case float64:
			dec = decimal.NewFromFloat(val)
		case string:
			dec, _ = decimal.NewFromString(val)
		default:
			data, _ := json.Marshal(v)
			dec, _ = decimal.NewFromString(string(data))
		}
		variables[k] = dec
	}

	if err := inv.FillVariables(cmd.EmissionID, variables, cmd.Version); err != nil {
		return nil, fmt.Errorf("failed to fill variables: %w", err)
	}

	if err := h.repo.Update(ctx, inv); err != nil {
		return nil, fmt.Errorf("failed to update inventory: %w", err)
	}

	event := inventory.EmissionVariablesFilled{
		InventoryID: inv.ID,
		EmissionID:  cmd.EmissionID,
	}
	h.publisher.Publish(ctx, "inventory.variables_filled", event)

	return inv, nil
}

func (h *InventoryCommandHandler) HandleAddEvidence(ctx context.Context, cmd AddEvidence) (*inventory.Inventory, error) {
	inv, err := h.repo.GetByID(ctx, cmd.InventoryID)
	if err != nil {
		return nil, fmt.Errorf("inventory not found: %w", err)
	}

	storageType := inventory.StorageType(cmd.StorageType)
	link, err := inventory.NewLink(uuid.New(), cmd.Name, cmd.Path, storageType)
	if err != nil {
		return nil, fmt.Errorf("failed to create link: %w", err)
	}

	if err := inv.AddEvidence(cmd.EmissionID, link, cmd.Version); err != nil {
		return nil, fmt.Errorf("failed to add evidence: %w", err)
	}

	if err := h.repo.Update(ctx, inv); err != nil {
		return nil, fmt.Errorf("failed to update inventory: %w", err)
	}

	event := inventory.EvidenceAdded{
		InventoryID: inv.ID,
		EmissionID:  cmd.EmissionID,
		LinkID:      link.ID,
	}
	h.publisher.Publish(ctx, "inventory.evidence_added", event)

	return inv, nil
}

func (h *InventoryCommandHandler) HandleCreateTemplate(ctx context.Context, cmd CreateTemplate) (*inventory.EmissionTemplate, error) {
	template, err := inventory.NewEmissionTemplate(
		uuid.New(),
		cmd.Name,
		make([]*inventory.Emission, 0),
		make([]*inventory.Link, 0),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create template: %w", err)
	}

	for _, linkDTO := range cmd.SupportingLinks {
		storageType := inventory.StorageType(linkDTO.StorageType)
		link, err := inventory.NewLink(uuid.New(), linkDTO.Name, linkDTO.Path, storageType)
		if err != nil {
			continue
		}
		template.SupportingLinks = append(template.SupportingLinks, link)
	}

	if err := h.templateRepo.Create(ctx, template); err != nil {
		return nil, fmt.Errorf("failed to save template: %w", err)
	}

	return template, nil
}

func (h *InventoryCommandHandler) HandleStoreReliabilityJob(ctx context.Context, cmd StoreReliabilityJob) (*inventory.Inventory, error) {
	inv, err := h.repo.GetByID(ctx, cmd.InventoryID)
	if err != nil {
		return nil, fmt.Errorf("inventory not found: %w", err)
	}

	if err := inv.StoreReliabilityJobID(cmd.EmissionID, cmd.ReliabilityJobID, cmd.Version); err != nil {
		return nil, fmt.Errorf("failed to store reliability job: %w", err)
	}

	if err := h.repo.Update(ctx, inv); err != nil {
		return nil, fmt.Errorf("failed to update inventory: %w", err)
	}

	return inv, nil
}

func parseEventTimestamp() int64 {
	return time.Now().Unix()
}