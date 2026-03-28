package inventory

import "github.com/google/uuid"

type InventoryCreated struct {
	InventoryID    uuid.UUID `json:"inventory_id"`
	CompanyBranchID uuid.UUID `json:"company_branch_id"`
	Month          int       `json:"month"`
	Year           int       `json:"year"`
	TemplateID     *uuid.UUID `json:"template_id"`
}

type InventoryStateChanged struct {
	InventoryID      uuid.UUID   `json:"inventory_id"`
	FromState        string      `json:"from_state"`
	ToState          string      `json:"to_state"`
	ActorID          uuid.UUID   `json:"actor_id"`
	ReviewMessage    *string     `json:"review_message,omitempty"`
	RecipientUserIDs []uuid.UUID `json:"recipient_user_ids,omitempty"`
	RecipientEmails  []string    `json:"recipient_emails,omitempty"`
}

type EmissionVariablesFilled struct {
	InventoryID uuid.UUID   `json:"inventory_id"`
	EmissionID  uuid.UUID   `json:"emission_id"`
}

type EvidenceAdded struct {
	InventoryID uuid.UUID   `json:"inventory_id"`
	EmissionID  uuid.UUID   `json:"emission_id"`
	LinkID      uuid.UUID   `json:"link_id"`
}

type InventoryAudited struct {
	InventoryID      uuid.UUID `json:"inventory_id"`
	AuditorID        uuid.UUID `json:"auditor_id"`
	CompanyAdminEmail string   `json:"company_admin_email"`
	CompanyBranchID  uuid.UUID `json:"company_branch_id"`
	Timestamp        int64     `json:"timestamp"`
}