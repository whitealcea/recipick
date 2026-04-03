CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS recipes (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL,
  source_url TEXT NOT NULL,
  title TEXT NOT NULL,
  note TEXT NULL,
  servings TEXT NULL,
  total_minutes INT NULL CHECK (total_minutes >= 0),
  source_recipe_jsonld JSONB NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (user_id, source_url)
);

CREATE TABLE IF NOT EXISTS recipe_ingredients (
  id BIGSERIAL PRIMARY KEY,
  recipe_id UUID NOT NULL REFERENCES recipes(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  sort_order INT NOT NULL DEFAULT 0,
  UNIQUE (recipe_id, name)
);

CREATE TABLE IF NOT EXISTS recipe_tags (
  id BIGSERIAL PRIMARY KEY,
  recipe_id UUID NOT NULL REFERENCES recipes(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  tag_type TEXT NOT NULL CHECK (tag_type IN ('system', 'user')),
  sort_order INT NOT NULL DEFAULT 0,
  UNIQUE (recipe_id, name)
);

CREATE INDEX IF NOT EXISTS idx_recipes_user_created ON recipes (user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_recipes_title_note_fts ON recipes USING GIN (to_tsvector('simple', coalesce(title,'') || ' ' || coalesce(note,'')));
CREATE INDEX IF NOT EXISTS idx_recipe_ingredients_recipe_name ON recipe_ingredients (recipe_id, name);
CREATE INDEX IF NOT EXISTS idx_recipe_tags_recipe_name ON recipe_tags (recipe_id, name);
