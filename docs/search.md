# Search Design (AND Conditions)

## 1. 要件再確認
`GET /v1/recipes` の検索条件は **すべて AND**。

- `q`: `title` + `note` の部分一致
- `tags`: 複数指定 AND
- `ingredients`: 複数指定 AND
- `maxMinutes`: 上限

式としては以下:

```text
match(q)
AND matchAll(tags)
AND matchAll(ingredients)
AND within(maxMinutes)
```

## 2. 実装SQL（実行可能）

```sql
WITH filtered AS (
  SELECT r.*
  FROM recipes r
  WHERE r.user_id = $1
    AND (
      $2 = ''
      OR r.title ILIKE '%' || $2 || '%'
      OR COALESCE(r.note, '') ILIKE '%' || $2 || '%'
    )
    AND (
      $3::int IS NULL
      OR (r.total_minutes IS NOT NULL AND r.total_minutes <= $3)
    )
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
ORDER BY f.created_at DESC;
```

パラメータ:
- `$1`: user_id (uuid)
- `$2`: q (text, 空文字可)
- `$3`: maxMinutes (int or null)
- `$4`: tags (text[])
- `$5`: ingredients (text[])
- `$6`: limit
- `$7`: offset

## 3. tags AND の実装方法

入力: `tags=["和食", "平日夜"]`

判定ロジック:
1. `recipe_tags` で `name = ANY(tags)` を抽出
2. `recipe_id` ごとに集約
3. `COUNT(DISTINCT name) == 入力タグ数` を満たす recipe のみ残す

これにより「和食 OR 平日夜」ではなく「和食 AND 平日夜」を保証できる。

## 4. ingredients AND の実装方法
- tags と同じ戦略を `recipe_ingredients` に適用
- `COUNT(DISTINCT ri.name) == array_length(input_ingredients)`

## 5. 正規化ルール（Go側）
SQLに渡す前に実施:
- trim
- 空文字除去
- 重複除去

例: ` [" 玉ねぎ ", "", "玉ねぎ", "鶏もも肉"] -> ["玉ねぎ", "鶏もも肉"]`

この正規化を行わないと、`HAVING COUNT(DISTINCT ...)` が意図とずれる。

## 6. パフォーマンス考慮

### 6.1 主要インデックス
```sql
CREATE INDEX IF NOT EXISTS idx_recipes_user_created
  ON recipes (user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_recipe_tags_recipe_name
  ON recipe_tags (recipe_id, name);

CREATE INDEX IF NOT EXISTS idx_recipe_ingredients_recipe_name
  ON recipe_ingredients (recipe_id, name);
```

### 6.2 検索規模の目安
- 1ユーザー数千件では本SQLで実用範囲
- 件数増加時は以下を検討
  - `pg_trgm` + `GIN(title gin_trgm_ops, note gin_trgm_ops)`
  - `EXISTS` ベースへの書き換え比較
  - 総件数 API（count）の別クエリ化

### 6.3 ページング
- `limit` は `1..100` に制限
- MVP は offset paging
- 将来的に cursor paging へ移行可能

## 7. 具体クエリ例

```sql
-- q="鶏", tags=["和食","平日夜"], ingredients=["玉ねぎ"], maxMinutes=30
SELECT *
FROM recipes r
WHERE r.user_id = '11111111-1111-1111-1111-111111111111'
  AND (r.title ILIKE '%鶏%' OR COALESCE(r.note, '') ILIKE '%鶏%')
  AND r.total_minutes <= 30
  AND r.id IN (
    SELECT recipe_id
    FROM recipe_tags
    WHERE name IN ('和食', '平日夜')
    GROUP BY recipe_id
    HAVING COUNT(DISTINCT name) = 2
  )
  AND r.id IN (
    SELECT recipe_id
    FROM recipe_ingredients
    WHERE name IN ('玉ねぎ')
    GROUP BY recipe_id
    HAVING COUNT(DISTINCT name) = 1
  );
```

