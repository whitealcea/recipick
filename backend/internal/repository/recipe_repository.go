package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"recipick/backend/internal/model"
)

var (
	ErrNotFound           = errors.New("recipe not found")
	ErrDuplicateSourceURL = errors.New("duplicate source url")
)

type SearchParams struct {
	UserID      uuid.UUID
	Q           string
	Tags        []string
	Ingredients []string
	MaxMinutes  *int
	Limit       int
	Offset      int
}

type CreateRecipeParams struct {
	UserID           uuid.UUID
	SourceURL        string
	Title            string
	Note             *string
	Servings         *string
	TotalMinutes     *int
	Ingredients      []string
	Tags             []model.RecipeTag
	SourceRecipeJSON json.RawMessage
}

type UpdateRecipeParams struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	SourceURL        *string
	Title            *string
	Note             *string
	Servings         *string
	TotalMinutes     *int
	Ingredients      *[]string
	Tags             *[]model.RecipeTag
	SourceRecipeJSON *json.RawMessage
}

type RecipeRepository struct {
	db *sql.DB
}

func NewRecipeRepository(db *sql.DB) *RecipeRepository {
	return &RecipeRepository{db: db}
}

func (r *RecipeRepository) Create(ctx context.Context, p CreateRecipeParams) (*model.Recipe, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO recipes (
			user_id, source_url, title, note, servings, total_minutes, source_recipe_jsonld
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`

	var rec model.Recipe
	rec.UserID = p.UserID
	rec.SourceURL = p.SourceURL
	rec.Title = p.Title
	rec.Note = p.Note
	rec.Servings = p.Servings
	rec.TotalMinutes = p.TotalMinutes
	rec.SourceRecipeJSON = p.SourceRecipeJSON
	if err := tx.QueryRowContext(
		ctx,
		query,
		p.UserID,
		p.SourceURL,
		p.Title,
		p.Note,
		p.Servings,
		p.TotalMinutes,
		nullableJSON(p.SourceRecipeJSON),
	).Scan(&rec.ID, &rec.CreatedAt, &rec.UpdatedAt); err != nil {
		if isUniqueViolation(err) {
			return nil, ErrDuplicateSourceURL
		}
		return nil, err
	}

	if err := replaceIngredientsTx(ctx, tx, rec.ID, p.Ingredients); err != nil {
		return nil, err
	}
	if err := replaceTagsTx(ctx, tx, rec.ID, p.Tags); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return r.GetByID(ctx, p.UserID, rec.ID)
}

func (r *RecipeRepository) GetByID(ctx context.Context, userID, recipeID uuid.UUID) (*model.Recipe, error) {
	query := `
		SELECT
			r.id,
			r.user_id,
			r.source_url,
			r.title,
			r.note,
			r.servings,
			r.total_minutes,
			r.source_recipe_jsonld,
			r.created_at,
			r.updated_at,
			COALESCE((
				SELECT json_agg(ri.name ORDER BY ri.sort_order)
				FROM recipe_ingredients ri
				WHERE ri.recipe_id = r.id
			), '[]'::json) AS ingredients,
			COALESCE((
				SELECT json_agg(json_build_object('name', rt.name, 'tagType', rt.tag_type) ORDER BY rt.sort_order)
				FROM recipe_tags rt
				WHERE rt.recipe_id = r.id
			), '[]'::json) AS tags
		FROM recipes r
		WHERE r.user_id = $1 AND r.id = $2
	`

	var rec model.Recipe
	var sourceJSON []byte
	var ingredientsJSON []byte
	var tagsJSON []byte
	if err := r.db.QueryRowContext(ctx, query, userID, recipeID).Scan(
		&rec.ID,
		&rec.UserID,
		&rec.SourceURL,
		&rec.Title,
		&rec.Note,
		&rec.Servings,
		&rec.TotalMinutes,
		&sourceJSON,
		&rec.CreatedAt,
		&rec.UpdatedAt,
		&ingredientsJSON,
		&tagsJSON,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if len(sourceJSON) > 0 {
		rec.SourceRecipeJSON = sourceJSON
	}
	if err := json.Unmarshal(ingredientsJSON, &rec.Ingredients); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(tagsJSON, &rec.Tags); err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *RecipeRepository) Search(ctx context.Context, p SearchParams) ([]model.Recipe, error) {
	if p.Limit <= 0 || p.Limit > 100 {
		p.Limit = 20
	}
	if p.Offset < 0 {
		p.Offset = 0
	}

	query := `
		WITH filtered AS (
			SELECT r.*
			FROM recipes r
			WHERE r.user_id = $1
			  AND ($2 = '' OR r.title ILIKE '%' || $2 || '%' OR COALESCE(r.note, '') ILIKE '%' || $2 || '%')
			  AND ($3::int IS NULL OR (r.total_minutes IS NOT NULL AND r.total_minutes <= $3))
			  AND (
				COALESCE(array_length($4::text[], 1), 0) = 0
				OR r.id IN (
					SELECT rt.recipe_id
					FROM recipe_tags rt
					WHERE rt.name = ANY($4::text[])
					GROUP BY rt.recipe_id
					HAVING COUNT(DISTINCT rt.name) = array_length($4::text[], 1)
				)
			  )
			  AND (
				COALESCE(array_length($5::text[], 1), 0) = 0
				OR r.id IN (
					SELECT ri.recipe_id
					FROM recipe_ingredients ri
					WHERE ri.name = ANY($5::text[])
					GROUP BY ri.recipe_id
					HAVING COUNT(DISTINCT ri.name) = array_length($5::text[], 1)
				)
			  )
			ORDER BY r.created_at DESC
			LIMIT $6 OFFSET $7
		)
		SELECT
			f.id,
			f.user_id,
			f.source_url,
			f.title,
			f.note,
			f.servings,
			f.total_minutes,
			f.source_recipe_jsonld,
			f.created_at,
			f.updated_at,
			COALESCE((
				SELECT json_agg(ri.name ORDER BY ri.sort_order)
				FROM recipe_ingredients ri
				WHERE ri.recipe_id = f.id
			), '[]'::json) AS ingredients,
			COALESCE((
				SELECT json_agg(json_build_object('name', rt.name, 'tagType', rt.tag_type) ORDER BY rt.sort_order)
				FROM recipe_tags rt
				WHERE rt.recipe_id = f.id
			), '[]'::json) AS tags
		FROM filtered f
		ORDER BY f.created_at DESC
	`

	rows, err := r.db.QueryContext(
		ctx,
		query,
		p.UserID,
		strings.TrimSpace(p.Q),
		p.MaxMinutes,
		pq.Array(normalizeValues(p.Tags)),
		pq.Array(normalizeValues(p.Ingredients)),
		p.Limit,
		p.Offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]model.Recipe, 0)
	for rows.Next() {
		var rec model.Recipe
		var sourceJSON []byte
		var ingredientsJSON []byte
		var tagsJSON []byte
		if err := rows.Scan(
			&rec.ID,
			&rec.UserID,
			&rec.SourceURL,
			&rec.Title,
			&rec.Note,
			&rec.Servings,
			&rec.TotalMinutes,
			&sourceJSON,
			&rec.CreatedAt,
			&rec.UpdatedAt,
			&ingredientsJSON,
			&tagsJSON,
		); err != nil {
			return nil, err
		}
		if len(sourceJSON) > 0 {
			rec.SourceRecipeJSON = sourceJSON
		}
		if err := json.Unmarshal(ingredientsJSON, &rec.Ingredients); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(tagsJSON, &rec.Tags); err != nil {
			return nil, err
		}
		result = append(result, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *RecipeRepository) Update(ctx context.Context, p UpdateRecipeParams) (*model.Recipe, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var existingID uuid.UUID
	if err := tx.QueryRowContext(ctx, `SELECT id FROM recipes WHERE user_id = $1 AND id = $2`, p.UserID, p.ID).Scan(&existingID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	updates := make([]string, 0)
	args := make([]any, 0)
	idx := 1

	if p.SourceURL != nil {
		updates = append(updates, fmt.Sprintf("source_url = $%d", idx))
		args = append(args, *p.SourceURL)
		idx++
	}
	if p.Title != nil {
		updates = append(updates, fmt.Sprintf("title = $%d", idx))
		args = append(args, *p.Title)
		idx++
	}
	if p.Note != nil {
		updates = append(updates, fmt.Sprintf("note = $%d", idx))
		args = append(args, nullableString(p.Note))
		idx++
	}
	if p.Servings != nil {
		updates = append(updates, fmt.Sprintf("servings = $%d", idx))
		args = append(args, nullableString(p.Servings))
		idx++
	}
	if p.TotalMinutes != nil {
		updates = append(updates, fmt.Sprintf("total_minutes = $%d", idx))
		args = append(args, nullableInt(p.TotalMinutes))
		idx++
	}
	if p.SourceRecipeJSON != nil {
		updates = append(updates, fmt.Sprintf("source_recipe_jsonld = $%d", idx))
		args = append(args, nullableJSON(*p.SourceRecipeJSON))
		idx++
	}

	if len(updates) > 0 {
		updates = append(updates, "updated_at = NOW()")
		query := fmt.Sprintf(
			"UPDATE recipes SET %s WHERE user_id = $%d AND id = $%d",
			strings.Join(updates, ", "),
			idx,
			idx+1,
		)
		args = append(args, p.UserID, p.ID)
		if _, err := tx.ExecContext(ctx, query, args...); err != nil {
			if isUniqueViolation(err) {
				return nil, ErrDuplicateSourceURL
			}
			return nil, err
		}
	}

	if p.Ingredients != nil {
		if err := replaceIngredientsTx(ctx, tx, p.ID, *p.Ingredients); err != nil {
			return nil, err
		}
	}
	if p.Tags != nil {
		if err := replaceTagsTx(ctx, tx, p.ID, *p.Tags); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return r.GetByID(ctx, p.UserID, p.ID)
}

func (r *RecipeRepository) Delete(ctx context.Context, userID, recipeID uuid.UUID) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM recipes WHERE user_id = $1 AND id = $2`, userID, recipeID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func replaceIngredientsTx(ctx context.Context, tx *sql.Tx, recipeID uuid.UUID, ingredients []string) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM recipe_ingredients WHERE recipe_id = $1`, recipeID); err != nil {
		return err
	}
	insert := `INSERT INTO recipe_ingredients (recipe_id, name, sort_order) VALUES ($1, $2, $3)`
	for i, name := range normalizeValues(ingredients) {
		if _, err := tx.ExecContext(ctx, insert, recipeID, name, i); err != nil {
			return err
		}
	}
	return nil
}

func replaceTagsTx(ctx context.Context, tx *sql.Tx, recipeID uuid.UUID, tags []model.RecipeTag) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM recipe_tags WHERE recipe_id = $1`, recipeID); err != nil {
		return err
	}
	insert := `INSERT INTO recipe_tags (recipe_id, name, tag_type, sort_order) VALUES ($1, $2, $3, $4)`
	order := 0
	for _, tag := range tags {
		name := strings.TrimSpace(tag.Name)
		if name == "" {
			continue
		}
		tagType := tag.TagType
		if tagType == "" {
			tagType = model.TagTypeUser
		}
		if _, err := tx.ExecContext(ctx, insert, recipeID, name, tagType, order); err != nil {
			return err
		}
		order++
	}
	return nil
}

func normalizeValues(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, v := range values {
		t := strings.TrimSpace(v)
		if t == "" {
			continue
		}
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		result = append(result, t)
	}
	return result
}

func nullableJSON(b json.RawMessage) any {
	if len(b) == 0 {
		return nil
	}
	return b
}

func nullableString(v *string) any {
	if v == nil {
		return nil
	}
	t := strings.TrimSpace(*v)
	if t == "" {
		return nil
	}
	return t
}

func nullableInt(v *int) any {
	if v == nil {
		return nil
	}
	if *v < 0 {
		return nil
	}
	return *v
}

func isUniqueViolation(err error) bool {
	pqErr, ok := err.(*pq.Error)
	if !ok {
		return false
	}
	return pqErr.Code == "23505"
}
