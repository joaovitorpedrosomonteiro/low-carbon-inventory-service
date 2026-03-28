package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/joaovitorpedrosomonteiro/low-carbon-inventory-service/internal/application/command"
	"github.com/joaovitorpedrosomonteiro/low-carbon-inventory-service/internal/application/query"
	"github.com/joaovitorpedrosomonteiro/low-carbon-inventory-service/internal/domain/inventory"
	"github.com/joaovitorpedrosomonteiro/low-carbon-inventory-service/internal/domain/valueobject"
)

type InventoryHandler struct {
	cmdHandler *command.InventoryCommandHandler
	queryHandler *query.InventoryQueryHandler
}

func NewInventoryHandler(cmdHandler *command.InventoryCommandHandler, queryHandler *query.InventoryQueryHandler) *InventoryHandler {
	return &InventoryHandler{
		cmdHandler:   cmdHandler,
		queryHandler: queryHandler,
	}
}

func (h *InventoryHandler) CreateInventory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req struct {
		Name            string    `json:"name"`
		Month           int       `json:"month"`
		Year            int       `json:"year"`
		CompanyBranchID string    `json:"company_branch_id"`
		TemplateID      *string   `json:"template_id"`
		GWPStandardID   *string   `json:"gwp_standard_id"`
	}

	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	companyBranchID, err := uuid.Parse(req.CompanyBranchID)
	if err != nil {
		http.Error(w, "invalid company_branch_id", http.StatusBadRequest)
		return
	}

	cmd := command.CreateInventory{
		Name:            req.Name,
		Month:           req.Month,
		Year:            req.Year,
		CompanyBranchID: companyBranchID,
		TemplateID:      nil,
		GWPStandardID:   nil,
	}

	if req.TemplateID != nil {
		tid, err := uuid.Parse(*req.TemplateID)
		if err == nil {
			cmd.TemplateID = &tid
		}
	}

	if req.GWPStandardID != nil {
		gwpid, err := uuid.Parse(*req.GWPStandardID)
		if err == nil {
			cmd.GWPStandardID = &gwpid
		}
	}

	inv, err := h.cmdHandler.HandleCreateInventory(r.Context(), cmd)
	if err != nil {
		log.Printf("Error creating inventory: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(inv)
}

func (h *InventoryHandler) GetInventory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/v1/inventories/")
	idStr = strings.TrimSuffix(idStr, "/")

	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid inventory id", http.StatusBadRequest)
		return
	}

	q := query.GetInventory{ID: id}
	inv, err := h.queryHandler.HandleGetInventory(r.Context(), q)
	if err != nil {
		http.Error(w, "inventory not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(inv)
}

func (h *InventoryHandler) ListInventories(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cursor := r.URL.Query().Get("cursor")
	limitStr := r.URL.Query().Get("limit")
	limit := 20

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	q := query.ListInventories{Cursor: cursor, Limit: limit}
	inventories, nextCursor, total, err := h.queryHandler.HandleListInventories(r.Context(), q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := struct {
		Data       []*inventory.Inventory `json:"data"`
		Pagination struct {
			NextCursor string `json:"next_cursor"`
			HasMore    bool   `json:"has_more"`
			TotalCount int    `json:"total_count"`
		} `json:"pagination"`
	}{
		Data: inventories,
	}

	response.Pagination.NextCursor = nextCursor
	response.Pagination.HasMore = nextCursor != ""
	response.Pagination.TotalCount = total

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *InventoryHandler) TransitionState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.Split(strings.TrimPrefix(r.URL.Path, "/v1/inventories/"), "/state")[0]
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid inventory id", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req struct {
		ToState       string `json:"to_state"`
		ActorID       string `json:"actor_id"`
		ReviewMessage *string `json:"review_message"`
		Version       int    `json:"version"`
	}

	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	actorID, err := uuid.Parse(req.ActorID)
	if err != nil {
		http.Error(w, "invalid actor_id", http.StatusBadRequest)
		return
	}

	cmd := command.TransitionState{
		InventoryID:   id,
		ToState:       req.ToState,
		ActorID:       actorID,
		ReviewMessage: req.ReviewMessage,
		Version:       req.Version,
	}

	inv, err := h.cmdHandler.HandleTransitionState(r.Context(), cmd)
	if err != nil {
		if strings.Contains(err.Error(), "version conflict") {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(inv)
}

func (h *InventoryHandler) FillVariables(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/v1/inventories/"), "/emissions/")
	if len(parts) != 2 {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	idParts := strings.Split(parts[0], "/")
	inventoryID, err := uuid.Parse(idParts[0])
	if err != nil {
		http.Error(w, "invalid inventory id", http.StatusBadRequest)
		return
	}

	emissionAndVars := strings.Split(parts[1], "/variables")
	if len(emissionAndVars) != 2 {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	emissionID, err := uuid.Parse(emissionAndVars[0])
	if err != nil {
		http.Error(w, "invalid emission id", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req struct {
		Variables map[string]interface{} `json:"variables"`
		Version   int                    `json:"version"`
	}

	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	cmd := command.FillVariables{
		InventoryID: inventoryID,
		EmissionID:  emissionID,
		Variables:   req.Variables,
		Version:     req.Version,
	}

	inv, err := h.cmdHandler.HandleFillVariables(r.Context(), cmd)
	if err != nil {
		if strings.Contains(err.Error(), "version conflict") {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(inv)
}

func (h *InventoryHandler) AddEvidence(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/v1/inventories/"), "/emissions/")
	if len(parts) != 2 {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	idParts := strings.Split(parts[0], "/")
	inventoryID, err := uuid.Parse(idParts[0])
	if err != nil {
		http.Error(w, "invalid inventory id", http.StatusBadRequest)
		return
	}

	evParts := strings.Split(parts[1], "/evidences")
	if len(evParts) != 2 {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	emissionID, err := uuid.Parse(evParts[0])
	if err != nil {
		http.Error(w, "invalid emission id", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req struct {
		Name        string `json:"name"`
		Path        string `json:"path"`
		StorageType string `json:"storage_type"`
		Version     int    `json:"version"`
	}

	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	cmd := command.AddEvidence{
		InventoryID:  inventoryID,
		EmissionID:   emissionID,
		Name:         req.Name,
		Path:         req.Path,
		StorageType:  req.StorageType,
		Version:      req.Version,
	}

	inv, err := h.cmdHandler.HandleAddEvidence(r.Context(), cmd)
	if err != nil {
		if strings.Contains(err.Error(), "version conflict") {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(inv)
}

func (h *InventoryHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/v1/inventories/"), "/summary")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid inventory id", http.StatusBadRequest)
		return
	}

	q := query.GetSummary{InventoryID: id}
	summary, err := h.queryHandler.HandleGetSummary(r.Context(), q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

func (h *InventoryHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	companyBranchIDStr := r.URL.Query().Get("company_branch_id")
	if companyBranchIDStr == "" {
		http.Error(w, "company_branch_id required", http.StatusBadRequest)
		return
	}

	companyBranchID, err := uuid.Parse(companyBranchIDStr)
	if err != nil {
		http.Error(w, "invalid company_branch_id", http.StatusBadRequest)
		return
	}

	q := query.GetDashboard{CompanyBranchID: companyBranchID}
	dashboard, err := h.queryHandler.HandleGetDashboard(r.Context(), q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dashboard)
}

func (h *InventoryHandler) StoreReliabilityJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/v1/inventories/"), "/emissions/")
	if len(parts) != 2 {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	idParts := strings.Split(parts[0], "/")
	inventoryID, err := uuid.Parse(idParts[0])
	if err != nil {
		http.Error(w, "invalid inventory id", http.StatusBadRequest)
		return
	}

	evParts := strings.Split(parts[1], "/reliability-job")
	if len(evParts) != 2 {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	emissionID, err := uuid.Parse(evParts[0])
	if err != nil {
		http.Error(w, "invalid emission id", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req struct {
		ReliabilityJobID string `json:"reliability_job_id"`
		Version          int    `json:"version"`
	}

	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	cmd := command.StoreReliabilityJob{
		InventoryID:      inventoryID,
		EmissionID:       emissionID,
		ReliabilityJobID: req.ReliabilityJobID,
		Version:          req.Version,
	}

	inv, err := h.cmdHandler.HandleStoreReliabilityJob(r.Context(), cmd)
	if err != nil {
		if strings.Contains(err.Error(), "version conflict") {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(inv)
}

func ParseMonth(s string) (valueobject.Month, error) {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid month: %w", err)
	}
	return valueobject.NewMonth(i)
}

func ParseYear(s string) (valueobject.Year, error) {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid year: %w", err)
	}
	return valueobject.NewYear(i)
}