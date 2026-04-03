-- name: CreateRecipe :one
INSERT INTO recipes (
  user_id, source_url, title, note, servings, total_minutes, source_recipe_jsonld
) VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetRecipeByID :one
SELECT *
FROM recipes
WHERE user_id = $1 AND id = $2;

-- name: DeleteRecipe :execrows
DELETE FROM recipes
WHERE user_id = $1 AND id = $2;

-- name: ReplaceIngredientsDelete :exec
DELETE FROM recipe_ingredients
WHERE recipe_id = $1;

-- name: AddIngredient :exec
INSERT INTO recipe_ingredients (recipe_id, name, sort_order)
VALUES ($1, $2, $3);

-- name: ReplaceTagsDelete :exec
DELETE FROM recipe_tags
WHERE recipe_id = $1;

-- name: AddTag :exec
INSERT INTO recipe_tags (recipe_id, name, tag_type, sort_order)
VALUES ($1, $2, $3, $4);

-- name: SearchRecipes :many
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
SELECT * FROM filtered
ORDER BY created_at DESC;
