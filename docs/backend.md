# Backend Implementation Guide

## 1. 目的
Go + chi + PostgreSQL で、レシピの CRUD / 検索 / Import preview を実装する。MVPなので、責務を `handler / service / repository` に限定する。

## 2. 推奨ディレクトリ構成

```text
backend/
  cmd/
    api/
      main.go
  internal/
    api/
      server.go
    handler/
      recipe_handler.go
    service/
      jsonld_importer.go
    repository/
      recipe_repository.go
    model/
      recipe.go
  db/
    migrations/
      0001_init.sql
    query/
      recipes.sql
  openapi/
    openapi.yaml
  sqlc.yaml
```

## 3. レイヤー設計

### 3.1 handler
- 役割
  - HTTP リクエストの decode / validation
  - Query param parse
  - status code と JSON response 組み立て
- やらないこと
  - SQL 実行
  - 外部HTTP呼び出し

### 3.2 service
- `JSONLDImporter` のみ
- URL fetch, HTML parse, JSON-LD parse
- import は preview 生成に限定

### 3.3 repository
- DB トランザクション管理
- SQL 実行
- DBエラーをドメインエラーへ変換
  - `23505` -> `ErrDuplicateSourceURL`

## 4. chi ルーティング例

```go
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
```

`internal/api/server.go` では `/v1` に mount する。

## 5. sqlc 前提の設計

### 5.1 方針
- 検索クエリは `db/query/recipes.sql` に定義
- SQL をまず固定し、Go 側で query builder を増やさない
- 返却型は sqlc 生成型を repository で `model.Recipe` に詰め替える

### 5.2 `sqlc.yaml` サンプル

```yaml
version: '2'
sql:
  - engine: postgresql
    schema: db/migrations
    queries: db/query
    gen:
      go:
        package: db
        out: internal/db
        sql_package: database/sql
```

### 5.3 repository での扱い
- Create/Update は transaction を使用
- ingredients/tags は差分更新でなく「全削除→再投入」
  - MVP でロジックを単純化
- `normalizeValues` で空文字と重複を除外

## 6. バリデーションルール（実装）
- `sourceUrl`, `title` 必須（POST）
- `totalMinutes >= 0`
- `maxMinutes >= 0`
- `Authorization: Bearer <JWT>` 必須（`sub` は UUID）

## 7. エラーハンドリング

統一フォーマット:

```json
{
  "error": {
    "code": "validation_error",
    "message": "totalMinutes must be >= 0",
    "details": {"field": "totalMinutes"}
  }
}
```

主な code:
- `invalid_request`
- `validation_error`
- `not_found`
- `duplicate_source_url`
- `internal_error`

## 8. ローカル起動手順（例）
1. PostgreSQL 起動
2. migration 適用
3. `go run ./cmd/api`
4. `GET /health` が `200 ok` を返すこと
