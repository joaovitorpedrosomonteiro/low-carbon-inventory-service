package query

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/joaovitorpedrosomonteiro/low-carbon-inventory-service/internal/domain/inventory"
	"github.com/joaovitorpedrosomonteiro/low-carbon-inventory-service/internal/domain/valueobject"
	"github.com/shopspring/decimal"
)

type GetInventory struct {
	ID uuid.UUID
}

type ListInventories struct {
	Cursor string
	Limit  int
}

type GetSummary struct {
	InventoryID uuid.UUID
}

type GetDashboard struct {
	CompanyBranchID uuid.UUID
}

type ListTemplates struct {
	Cursor string
	Limit  int
}

type InventoryQueryHandler struct {
	repo          inventory.InventoryRepository
	templateRepo  inventory.EmissionTemplateRepository
	gwpRepo       GWPFactorRepository
}

type GWPFactorRepository interface {
	GetByGasType(ctx context.Context, gasType valueobject.GasType) (valueobject.ConversionFactor, error)
}

func NewInventoryQueryHandler(repo inventory.InventoryRepository, templateRepo inventory.EmissionTemplateRepository, gwpRepo GWPFactorRepository) *InventoryQueryHandler {
	return &InventoryQueryHandler{
		repo:         repo,
		templateRepo: templateRepo,
		gwpRepo:      gwpRepo,
	}
}

func (h *InventoryQueryHandler) HandleGetInventory(ctx context.Context, q GetInventory) (*inventory.Inventory, error) {
	inv, err := h.repo.GetByID(ctx, q.ID)
	if err != nil {
		return nil, fmt.Errorf("inventory not found: %w", err)
	}
	return inv, nil
}

func (h *InventoryQueryHandler) HandleListInventories(ctx context.Context, q ListInventories) ([]*inventory.Inventory, string, int, error) {
	if q.Limit == 0 {
		q.Limit = 20
	}
	if q.Limit > 100 {
		q.Limit = 100
	}

	return h.repo.List(ctx, q.Cursor, q.Limit)
}

func (h *InventoryQueryHandler) HandleGetSummary(ctx context.Context, q GetSummary) (*Summary, error) {
	inv, err := h.repo.GetByID(ctx, q.InventoryID)
	if err != nil {
		return nil, fmt.Errorf("inventory not found: %w", err)
	}

	summary := &Summary{
		InventoryID:   inv.ID,
		TotalEmission: decimal.Zero,
		CO2Equivalent: decimal.Zero,
		ByGasType:     make(map[string]decimal.Decimal),
	}

	for _, e := range inv.Emissions {
		if !e.IsComplete() {
			continue
		}

		total, err := e.TotalEmissionTons()
		if err != nil {
			continue
		}

		summary.TotalEmission = summary.TotalEmission.Add(total)

		gwp, err := h.gwpRepo.GetByGasType(ctx, e.GasType)
		if err == nil && !gwp.Value.IsZero() {
			co2Eq := total.Mul(gwp.Value)
			summary.CO2Equivalent = summary.CO2Equivalent.Add(co2Eq)
			summary.ByGasType[e.GasType.String()] = summary.ByGasType[e.GasType.String()].Add(co2Eq)
		}
	}

	return summary, nil
}

func (h *InventoryQueryHandler) HandleGetDashboard(ctx context.Context, q GetDashboard) (*DashboardData, error) {
	limit := 100
	inventories, _, total, err := h.repo.List(ctx, "", limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list inventories: %w", err)
	}

	dashboard := &DashboardData{
		TotalInventories:     total,
		ByState:              make(map[string]int),
		EmissionsByScope:     make(map[string]map[string]decimal.Decimal),
		RecentInventories:    make([]RecentInventory, 0),
	}

	for _, inv := range inventories {
		state := inv.State.String()
		dashboard.ByState[state]++

		isRelevant := inv.CompanyBranchID == q.CompanyBranchID
		if isRelevant {
			recent := RecentInventory{
				ID:        inv.ID,
				Name:      inv.Name,
				Month:     inv.Month.Int(),
				Year:      inv.Year.Int(),
				State:     state,
				Version:   inv.Version,
			}
			dashboard.RecentInventories = append(dashboard.RecentInventories, recent)
		}
	}

	return dashboard, nil
}

func (h *InventoryQueryHandler) HandleListTemplates(ctx context.Context, q ListTemplates) ([]*inventory.EmissionTemplate, string, int, error) {
	if q.Limit == 0 {
		q.Limit = 20
	}
	if q.Limit > 100 {
		q.Limit = 100
	}

	return h.templateRepo.List(ctx, q.Cursor, q.Limit)
}

type Summary struct {
	InventoryID   uuid.UUID
	TotalEmission decimal.Decimal
	CO2Equivalent decimal.Decimal
	ByGasType     map[string]decimal.Decimal
}

type DashboardData struct {
	TotalInventories  int                         `json:"total_inventories"`
	ByState           map[string]int              `json:"by_state"`
	EmissionsByScope  map[string]map[string]decimal.Decimal `json:"emissions_by_scope"`
	RecentInventories []RecentInventory           `json:"recent_inventories"`
}

type RecentInventory struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Month   int       `json:"month"`
	Year    int       `json:"year"`
	State   string    `json:"state"`
	Version int       `json:"version"`
}