package inventory

import (
	"context"

	"github.com/google/uuid"
)

type InventoryRepository interface {
	Create(ctx context.Context, inv *Inventory) error
	GetByID(ctx context.Context, id uuid.UUID) (*Inventory, error)
	Update(ctx context.Context, inv *Inventory) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, cursor string, limit int) ([]*Inventory, string, int, error)
	GetByCompanyBranchAndPeriod(ctx context.Context, companyBranchID uuid.UUID, month int, year int) (*Inventory, error)
}

type EmissionTemplateRepository interface {
	Create(ctx context.Context, template *EmissionTemplate) error
	GetByID(ctx context.Context, id uuid.UUID) (*EmissionTemplate, error)
	Update(ctx context.Context, template *EmissionTemplate) error
	List(ctx context.Context, cursor string, limit int) ([]*EmissionTemplate, string, int, error)
}