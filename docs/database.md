# Database Design

## 1. DDL（実行可能）

```sql
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

CREATE INDEX IF NOT EXISTS idx_recipes_user_created
  ON recipes (user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_recipes_title_note_fts
  ON recipes USING GIN (
    to_tsvector('simple', coalesce(title, '') || ' ' || coalesce(note, ''))
  );

CREATE INDEX IF NOT EXISTS idx_recipe_ingredients_recipe_name
  ON recipe_ingredients (recipe_id, name);

CREATE INDEX IF NOT EXISTS idx_recipe_tags_recipe_name
  ON recipe_tags (recipe_id, name);
```

## 2. テーブルとカラム説明

### 2.1 recipes
- `id`: レシピID（UUID）
- `user_id`: ユーザーID（JWT の `sub` claim）
- `source_url`: 元ページURL
- `title`: レシピ名（必須）
- `note`: 任意メモ
- `servings`: 2人分などの表示文字列
- `total_minutes`: 所要時間（分）
- `source_recipe_jsonld`: 抽出した JSON-LD 生データ
- `created_at`, `updated_at`: 監査用時刻

### 2.2 recipe_ingredients
- `recipe_id`: 親レシピ
- `name`: 材料名
- `sort_order`: 画面表示順

### 2.3 recipe_tags
- `recipe_id`: 親レシピ
- `name`: タグ名
- `tag_type`: `system` or `user`
- `sort_order`: 画面表示順

## 3. ユニーク制約

### 3.1 同一URL重複禁止
- 制約: `UNIQUE (user_id, source_url)`
- 効果: 同一ユーザーで同じ URL を保存不可
- API 挙動: DB 一意制約違反 `23505` を `409 duplicate_source_url` へ変換

## 4. インデックス設計意図
- `idx_recipes_user_created`: ユーザー別一覧 + 新着順
- `idx_recipe_tags_recipe_name`: tags AND サブクエリを高速化
- `idx_recipe_ingredients_recipe_name`: ingredients AND サブクエリを高速化
- `idx_recipes_title_note_fts`: 将来の全文検索最適化に備える

注: 現行検索は `ILIKE` を使用するため、`idx_recipes_title_note_fts` は直接利用されません。MVP ではまず正確性を優先し、必要時に trigram index へ拡張します。

## 5. 代表クエリ

### 5.1 重複チェックを伴うINSERT（実運用は repository 経由）
```sql
INSERT INTO recipes (user_id, source_url, title)
VALUES ('11111111-1111-1111-1111-111111111111', 'https://example.com/r1', '豚汁')
RETURNING id;
```

### 5.2 レシピ削除（関連材料・タグは cascade）
```sql
DELETE FROM recipes
WHERE user_id = '11111111-1111-1111-1111-111111111111'
  AND id = 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa';
```
