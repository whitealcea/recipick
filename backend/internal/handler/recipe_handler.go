package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"recipick/backend/internal/auth"
	"recipick/backend/internal/model"
	"recipick/backend/internal/repository"
	"recipick/backend/internal/service"
)

type RecipeHandler struct {
	repo     *repository.RecipeRepository
	importer *service.JSONLDImporter
}

func NewRecipeHandler(repo *repository.RecipeRepository, importer *service.JSONLDImporter) *RecipeHandler {
	return &RecipeHandler{repo: repo, importer: importer}
}

func (h *RecipeHandler) Router(authMiddleware func(http.Handler) http.Handler) http.Handler {
	r := chi.NewRouter()
	r.Use(authMiddleware)

	r.Post("/recipes/import", h.ImportRecipe)
	r.Post("/recipes", h.CreateRecipe)
	r.Get("/recipes", h.SearchRecipes)
	r.Get("/recipes/{id}", h.GetRecipe)
	r.Patch("/recipes/{id}", h.UpdateRecipe)
	r.Delete("/recipes/{id}", h.DeleteRecipe)

	return r
}

type importRequest struct {
	SourceURL string `json:"sourceUrl"`
}

func (h *RecipeHandler) ImportRecipe(w http.ResponseWriter, r *http.Request) {
	var req importRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body", nil)
		return
	}
	req.SourceURL = strings.TrimSpace(req.SourceURL)
	if req.SourceURL == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "sourceUrl is required", map[string]string{"field": "sourceUrl"})
		return
	}

	preview := h.importer.ImportPreview(r.Context(), req.SourceURL)
	writeJSON(w, http.StatusOK, preview)
}

type createRecipeRequest struct {
	SourceURL        string            `json:"sourceUrl"`
	Title            string            `json:"title"`
	Note             *string           `json:"note"`
	Servings         *string           `json:"servings"`
	TotalMinutes     *int              `json:"totalMinutes"`
	Ingredients      []string          `json:"ingredients"`
	Tags             []model.RecipeTag `json:"tags"`
	SourceRecipeJSON json.RawMessage   `json:"sourceRecipeJsonLd"`
}

func (h *RecipeHandler) CreateRecipe(w http.ResponseWriter, r *http.Request) {
	var req createRecipeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body", nil)
		return
	}
	req.SourceURL = strings.TrimSpace(req.SourceURL)
	req.Title = strings.TrimSpace(req.Title)
	if req.SourceURL == "" || req.Title == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "sourceUrl and title are required", nil)
		return
	}
	if req.TotalMinutes != nil && *req.TotalMinutes < 0 {
		writeError(w, http.StatusBadRequest, "validation_error", "totalMinutes must be >= 0", map[string]string{"field": "totalMinutes"})
		return
	}

	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusInternalServerError, "internal_error", "missing user context", nil)
		return
	}

	recipe, err := h.repo.Create(r.Context(), repository.CreateRecipeParams{
		UserID:           userID,
		SourceURL:        req.SourceURL,
		Title:            req.Title,
		Note:             req.Note,
		Servings:         req.Servings,
		TotalMinutes:     req.TotalMinutes,
		Ingredients:      req.Ingredients,
		Tags:             req.Tags,
		SourceRecipeJSON: req.SourceRecipeJSON,
	})
	if err != nil {
		h.handleRepoError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, recipe)
}

func (h *RecipeHandler) GetRecipe(w http.ResponseWriter, r *http.Request) {
	recipeID, ok := parseRecipeID(w, r)
	if !ok {
		return
	}
	userID, _ := auth.UserIDFromContext(r.Context())
	recipe, err := h.repo.GetByID(r.Context(), userID, recipeID)
	if err != nil {
		h.handleRepoError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, recipe)
}

func (h *RecipeHandler) SearchRecipes(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	q := r.URL.Query().Get("q")
	tags := csvToSlice(r.URL.Query().Get("tags"))
	ingredients := csvToSlice(r.URL.Query().Get("ingredients"))

	var maxMinutes *int
	if raw := strings.TrimSpace(r.URL.Query().Get("maxMinutes")); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 0 {
			writeError(w, http.StatusBadRequest, "validation_error", "maxMinutes must be a non-negative integer", map[string]string{"field": "maxMinutes"})
			return
		}
		maxMinutes = &v
	}

	limit := 20
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		v, err := strconv.Atoi(raw)
		if err == nil {
			limit = v
		}
	}
	offset := 0
	if raw := strings.TrimSpace(r.URL.Query().Get("offset")); raw != "" {
		v, err := strconv.Atoi(raw)
		if err == nil {
			offset = v
		}
	}

	recipes, err := h.repo.Search(r.Context(), repository.SearchParams{
		UserID:      userID,
		Q:           q,
		Tags:        tags,
		Ingredients: ingredients,
		MaxMinutes:  maxMinutes,
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		h.handleRepoError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"items": recipes})
}

type updateRecipeRequest struct {
	SourceURL        *string            `json:"sourceUrl"`
	Title            *string            `json:"title"`
	Note             *string            `json:"note"`
	Servings         *string            `json:"servings"`
	TotalMinutes     *int               `json:"totalMinutes"`
	Ingredients      *[]string          `json:"ingredients"`
	Tags             *[]model.RecipeTag `json:"tags"`
	SourceRecipeJSON *json.RawMessage   `json:"sourceRecipeJsonLd"`
}

func (h *RecipeHandler) UpdateRecipe(w http.ResponseWriter, r *http.Request) {
	recipeID, ok := parseRecipeID(w, r)
	if !ok {
		return
	}
	var req updateRecipeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body", nil)
		return
	}
	if req.SourceURL != nil {
		t := strings.TrimSpace(*req.SourceURL)
		req.SourceURL = &t
	}
	if req.Title != nil {
		t := strings.TrimSpace(*req.Title)
		req.Title = &t
	}
	if req.TotalMinutes != nil && *req.TotalMinutes < 0 {
		writeError(w, http.StatusBadRequest, "validation_error", "totalMinutes must be >= 0", map[string]string{"field": "totalMinutes"})
		return
	}

	userID, _ := auth.UserIDFromContext(r.Context())
	recipe, err := h.repo.Update(r.Context(), repository.UpdateRecipeParams{
		ID:               recipeID,
		UserID:           userID,
		SourceURL:        req.SourceURL,
		Title:            req.Title,
		Note:             req.Note,
		Servings:         req.Servings,
		TotalMinutes:     req.TotalMinutes,
		Ingredients:      req.Ingredients,
		Tags:             req.Tags,
		SourceRecipeJSON: req.SourceRecipeJSON,
	})
	if err != nil {
		h.handleRepoError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, recipe)
}

func (h *RecipeHandler) DeleteRecipe(w http.ResponseWriter, r *http.Request) {
	recipeID, ok := parseRecipeID(w, r)
	if !ok {
		return
	}
	userID, _ := auth.UserIDFromContext(r.Context())
	if err := h.repo.Delete(r.Context(), userID, recipeID); err != nil {
		h.handleRepoError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *RecipeHandler) handleRepoError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", "recipe not found", nil)
	case errors.Is(err, repository.ErrDuplicateSourceURL):
		writeError(w, http.StatusConflict, "duplicate_source_url", "sourceUrl already exists for this user", nil)
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "internal server error", nil)
	}
}

func parseRecipeID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	rawID := chi.URLParam(r, "id")
	recipeID, err := uuid.Parse(rawID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", "invalid recipe id", map[string]string{"field": "id"})
		return uuid.Nil, false
	}
	return recipeID, true
}

func csvToSlice(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t != "" {
			result = append(result, t)
		}
	}
	return result
}

type errorEnvelope struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func writeError(w http.ResponseWriter, status int, code, message string, details any) {
	writeJSON(w, status, errorEnvelope{Error: errorBody{Code: code, Message: message, Details: details}})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func NewDB(dsn string) (*sql.DB, error) {
	return sql.Open("postgres", dsn)
}
