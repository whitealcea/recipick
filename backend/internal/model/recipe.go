package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type TagType string

const (
	TagTypeSystem TagType = "system"
	TagTypeUser   TagType = "user"
)

type RecipeTag struct {
	Name    string  `json:"name"`
	TagType TagType `json:"tagType"`
}

type Recipe struct {
	ID               uuid.UUID       `json:"id"`
	UserID           uuid.UUID       `json:"-"`
	SourceURL        string          `json:"sourceUrl"`
	Title            string          `json:"title"`
	Note             *string         `json:"note,omitempty"`
	Servings         *string         `json:"servings,omitempty"`
	TotalMinutes     *int            `json:"totalMinutes,omitempty"`
	Ingredients      []string        `json:"ingredients"`
	Tags             []RecipeTag     `json:"tags"`
	SourceRecipeJSON json.RawMessage `json:"sourceRecipeJsonLd,omitempty"`
	CreatedAt        time.Time       `json:"createdAt"`
	UpdatedAt        time.Time       `json:"updatedAt"`
}

type ImportPreview struct {
	SourceURL      string          `json:"sourceUrl"`
	Title          *string         `json:"title,omitempty"`
	Servings       *string         `json:"servings,omitempty"`
	TotalMinutes   *int            `json:"totalMinutes,omitempty"`
	Ingredients    []string        `json:"ingredients"`
	SystemTags     []string        `json:"systemTags"`
	RawJSONLD      json.RawMessage `json:"rawJsonLd,omitempty"`
	JSONLDFound    bool            `json:"jsonLdFound"`
	ImportWarnings []string        `json:"warnings"`
}
