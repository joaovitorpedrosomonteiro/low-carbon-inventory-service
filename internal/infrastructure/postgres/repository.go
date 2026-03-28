package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joaovitorpedrosomonteiro/low-carbon-inventory-service/internal/domain/inventory"
	"github.com/joaovitorpedrosomonteiro/low-carbon-inventory-service/internal/domain/valueobject"
	"github.com/shopspring/decimal"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Create(ctx context.Context, inv *inventory.Inventory) error {
	query := `
		INSERT INTO inventories (id, name, month, year, state, template_id, company_branch_id, gwp_standard_id, review_message, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.pool.Exec(ctx, query,
		inv.ID, inv.Name, inv.Month.Int(), inv.Year.Int(), inv.State.String(),
		inv.TemplateID, inv.CompanyBranchID, inv.GWPStandardID, inv.ReviewMessage, inv.Version,
	)
	return err
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*inventory.Inventory, error) {
	query := `
		SELECT id, name, month, year, state, template_id, company_branch_id, gwp_standard_id, review_message, version
		FROM inventories WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)

	var inv inventory.Inventory
	var templateID, gwpStandardID *uuid.UUID
	var reviewMessage *string

	err := row.Scan(&inv.ID, &inv.Name, &inv.Month, &inv.Year, &inv.State, &templateID, &inv.CompanyBranchID, &gwpStandardID, &reviewMessage, &inv.Version)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("inventory not found")
		}
		return nil, err
	}

	inv.TemplateID = templateID
	inv.GWPStandardID = gwpStandardID
	inv.ReviewMessage = reviewMessage

	inv.Emissions, _ = r.listEmissionsByInventory(ctx, inv.ID)

	return &inv, nil
}

func (r *PostgresRepository) Update(ctx context.Context, inv *inventory.Inventory) error {
	query := `
		UPDATE inventories 
		SET name = $1, month = $2, year = $3, state = $4, template_id = $5, gwp_standard_id = $6, 
		    review_message = $7, version = $8
		WHERE id = $9 AND version = $10
	`

	_, err := r.pool.Exec(ctx, query,
		inv.Name, inv.Month.Int(), inv.Year.Int(), inv.State.String(),
		inv.TemplateID, inv.GWPStandardID, inv.ReviewMessage, inv.Version, inv.ID, inv.Version-1,
	)
	if err != nil {
		return fmt.Errorf("failed to update inventory: %w", err)
	}

	return nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM inventories WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *PostgresRepository) List(ctx context.Context, cursor string, limit int) ([]*inventory.Inventory, string, int, error) {
	query := `
		SELECT id, name, month, year, state, template_id, company_branch_id, gwp_standard_id, review_message, version
		FROM inventories
		ORDER BY created_at DESC
		LIMIT $1
	`

	rows, err := r.pool.Query(ctx, query, limit+1)
	if err != nil {
		return nil, "", 0, err
	}
	defer rows.Close()

	var inventories []*inventory.Inventory
	for rows.Next() {
		var inv inventory.Inventory
		var templateID, gwpStandardID *uuid.UUID

		err := rows.Scan(&inv.ID, &inv.Name, &inv.Month, &inv.Year, &inv.State, &templateID, &inv.CompanyBranchID, &gwpStandardID, &inv.ReviewMessage, &inv.Version)
		if err != nil {
			return nil, "", 0, err
		}
		inv.TemplateID = templateID
		inv.GWPStandardID = gwpStandardID
		inventories = append(inventories, &inv)
	}

	var nextCursor string
	if len(inventories) > limit {
		inventories = inventories[:limit]
		nextCursor = inventories[len(inventories)-1].ID.String()
	}

	return inventories, nextCursor, len(inventories), nil
}

func (r *PostgresRepository) GetByCompanyBranchAndPeriod(ctx context.Context, companyBranchID uuid.UUID, month int, year int) (*inventory.Inventory, error) {
	query := `
		SELECT id, name, month, year, state, template_id, company_branch_id, gwp_standard_id, review_message, version
		FROM inventories 
		WHERE company_branch_id = $1 AND month = $2 AND year = $3
	`

	row := r.pool.QueryRow(ctx, query, companyBranchID, month, year)

	var inv inventory.Inventory
	var templateID, gwpStandardID *uuid.UUID

	err := row.Scan(&inv.ID, &inv.Name, &inv.Month, &inv.Year, &inv.State, &templateID, &inv.CompanyBranchID, &gwpStandardID, &inv.ReviewMessage, &inv.Version)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	inv.TemplateID = templateID
	inv.GWPStandardID = gwpStandardID

	return &inv, nil
}

func (r *PostgresRepository) listEmissionsByInventory(ctx context.Context, inventoryID uuid.UUID) ([]*inventory.Emission, error) {
	query := `
		SELECT e.id, e.name, e.formula, e.category_id, e.reliability_job_id
		FROM emissions e
		WHERE e.inventory_id = $1
	`

	rows, err := r.pool.Query(ctx, query, inventoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var emissions []*inventory.Emission
	for rows.Next() {
		var e inventory.Emission
		if err := rows.Scan(&e.ID, &e.Name, &e.Formula, &e.Category, &e.ReliabilityJobID); err != nil {
			continue
		}
		emissions = append(emissions, &e)
	}

	return emissions, nil
}

type PostgresTemplateRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresTemplateRepository(pool *pgxpool.Pool) *PostgresTemplateRepository {
	return &PostgresTemplateRepository{pool: pool}
}

func (r *PostgresTemplateRepository) Create(ctx context.Context, template *inventory.EmissionTemplate) error {
	query := `
		INSERT INTO emission_templates (id, name, inventory_count, is_frozen)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.pool.Exec(ctx, query, template.ID, template.Name, template.InventoryCount, template.IsFrozen)
	return err
}

func (r *PostgresTemplateRepository) GetByID(ctx context.Context, id uuid.UUID) (*inventory.EmissionTemplate, error) {
	query := `
		SELECT id, name, inventory_count, is_frozen
		FROM emission_templates WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)

	var template inventory.EmissionTemplate
	err := row.Scan(&template.ID, &template.Name, &template.InventoryCount, &template.IsFrozen)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("template not found")
		}
		return nil, err
	}

	return &template, nil
}

func (r *PostgresTemplateRepository) Update(ctx context.Context, template *inventory.EmissionTemplate) error {
	query := `
		UPDATE emission_templates 
		SET name = $1, inventory_count = $2, is_frozen = $3
		WHERE id = $4
	`

	_, err := r.pool.Exec(ctx, query, template.Name, template.InventoryCount, template.IsFrozen, template.ID)
	return err
}

func (r *PostgresTemplateRepository) List(ctx context.Context, cursor string, limit int) ([]*inventory.EmissionTemplate, string, int, error) {
	query := `
		SELECT id, name, inventory_count, is_frozen
		FROM emission_templates
		ORDER BY created_at DESC
		LIMIT $1
	`

	rows, err := r.pool.Query(ctx, query, limit+1)
	if err != nil {
		return nil, "", 0, err
	}
	defer rows.Close()

	var templates []*inventory.EmissionTemplate
	for rows.Next() {
		var t inventory.EmissionTemplate
		if err := rows.Scan(&t.ID, &t.Name, &t.InventoryCount, &t.IsFrozen); err != nil {
			continue
		}
		templates = append(templates, &t)
	}

	var nextCursor string
	if len(templates) > limit {
		templates = templates[:limit]
		nextCursor = templates[len(templates)-1].ID.String()
	}

	return templates, nextCursor, len(templates), nil
}

type GWPRepository struct {
	pool *pgxpool.Pool
}

func NewGWPRepository(pool *pgxpool.Pool) *GWPRepository {
	return &GWPRepository{pool: pool}
}

func (r *GWPRepository) GetByGasType(ctx context.Context, gasType valueobject.GasType) (valueobject.ConversionFactor, error) {
	query := `
		SELECT cf.value, cf.gas_type
		FROM conversion_factors cf
		JOIN gwp_standards gs ON cf.gwp_standard_id = gs.id
		WHERE cf.gas_type = $1 AND gs.is_default = true
		LIMIT 1
	`

	row := r.pool.QueryRow(ctx, query, gasType.String())

	var value decimal.Decimal
	var gType string

	err := row.Scan(&value, &gType)
	if err == pgx.ErrNoRows {
		return valueobject.ConversionFactor{}, fmt.Errorf("GWP factor not found for gas type")
	}
	if err != nil {
		return valueobject.ConversionFactor{}, err
	}

	gt, _ := valueobject.NewGasTypeFromFormula(gType)
	return valueobject.NewConversionFactor(gt, value)
}