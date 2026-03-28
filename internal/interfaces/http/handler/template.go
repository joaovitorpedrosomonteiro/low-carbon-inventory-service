package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/joaovitorpedrosomonteiro/low-carbon-inventory-service/internal/application/command"
	"github.com/joaovitorpedrosomonteiro/low-carbon-inventory-service/internal/application/query"
)

type TemplateHandler struct {
	cmdHandler   *command.InventoryCommandHandler
	queryHandler *query.InventoryQueryHandler
}

func NewTemplateHandler(cmdHandler *command.InventoryCommandHandler, queryHandler *query.InventoryQueryHandler) *TemplateHandler {
	return &TemplateHandler{
		cmdHandler:   cmdHandler,
		queryHandler: queryHandler,
	}
}

func (h *TemplateHandler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
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
		Name            string `json:"name"`
		EmissionIDs     []string `json:"emission_ids"`
		SupportingLinks []struct {
			Name        string `json:"name"`
			Path        string `json:"path"`
			StorageType string `json:"storage_type"`
		} `json:"supporting_links"`
	}

	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	cmd := command.CreateTemplate{
		Name:        req.Name,
		EmissionIDs: make([]uuid.UUID, 0),
	}

	var links []command.LinkDTO
	for _, l := range req.SupportingLinks {
		links = append(links, command.LinkDTO{
			Name:        l.Name,
			Path:        l.Path,
			StorageType: l.StorageType,
		})
	}
	cmd.SupportingLinks = links

	template, err := h.cmdHandler.HandleCreateTemplate(r.Context(), cmd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(template)
}

func (h *TemplateHandler) ListTemplates(w http.ResponseWriter, r *http.Request) {
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

	q := query.ListTemplates{Cursor: cursor, Limit: limit}
	templates, nextCursor, total, err := h.queryHandler.HandleListTemplates(r.Context(), q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := struct {
		Data       interface{} `json:"data"`
		Pagination struct {
			NextCursor string `json:"next_cursor"`
			HasMore    bool   `json:"has_more"`
			TotalCount int    `json:"total_count"`
		} `json:"pagination"`
	}{
		Data: templates,
	}

	response.Pagination.NextCursor = nextCursor
	response.Pagination.HasMore = nextCursor != ""
	response.Pagination.TotalCount = total

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func HandleHealthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func HandleReadyz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("READY"))
}