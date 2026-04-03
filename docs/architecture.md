# Architecture

## 1. システム構成（テキスト図）

```text
[Browser (React + Vite)]
    |
    | HTTPS JSON (REST)
    v
[Go API (chi on Render)]
    |
    | SQL (sqlc / repository)
    v
[PostgreSQL (Supabase)]

Import path:
Browser -> POST /v1/recipes/import -> Go API -> fetch source URL (HTTP GET) -> JSON-LD parse -> preview response
```

## 2. コンポーネント責務

### Frontend
- 画面状態: `useState`
- サーバー状態: `TanStack Query`
- API クライアント: `fetch` ベース
- 全リクエストで `Authorization: Bearer <JWT>` を付与

### Backend
- `handler`: HTTP I/O, validation, status code
- `service`: JSON-LD import の外部通信/変換
- `repository`: SQL 実行, トランザクション, DB エラー変換
- `model`: API/ドメイン構造体

### DB
- 正規化された3テーブル
- タグ/材料は 1:N
- URL重複を DB 制約で強制

## 3. データフロー

### 3.1 Import preview（保存しない）
1. `POST /v1/recipes/import` で `sourceUrl` 受信
2. バックエンドが対象 URL の HTML を最大 5MB 読み込み
3. `<script type="application/ld+json">` を抽出
4. `@type=Recipe` を探索
5. `title`, `recipeIngredient`, `recipeCategory/recipeCuisine/keywords`, `totalTime` を抽出
6. `rawJsonLd` と warning を含む preview を返却

### 3.2 Save recipe
1. フロントが preview を編集
2. `POST /v1/recipes`
3. `recipes` INSERT
4. `recipe_ingredients`, `recipe_tags` を挿入（同一 recipe 内で重複除去）
5. `201` で作成済みデータ返却

### 3.3 Search recipes
1. `GET /v1/recipes?q=...&tags=...&ingredients=...&maxMinutes=...`
2. SQL の WHERE で全条件 AND
3. tags/ingredients は `GROUP BY recipe_id HAVING COUNT(DISTINCT ...) = 入力数`
4. `created_at DESC` で返却

## 4. 技術選定理由（具体）

### React + TypeScript + Vite
- 開発起動が高速（Vite）
- API 型との整合を TypeScript で担保
- MVP の UI 速度に対して十分

### Go + chi
- ミドルウェア/ルーティングが軽量
- 単一バイナリで Render 配置が簡単
- `net/http` ベースで過剰抽象を避けられる

### PostgreSQL + Supabase
- JSONB 保存（`source_recipe_jsonld`）が素直
- 複合 unique・GIN・部分一致設計が実用的
- 低運用コストで開始可能

### sqlc 前提
- SQL を一次情報として管理できる
- 文字列組み立てを減らし、型付きクエリで事故を減らす

## 5. 非機能の最小方針
- API タイムアウト: 15秒
- Import 外部取得タイムアウト: 10秒
- ログ: リクエストID単位で追跡
- 初期性能目標: 1ユーザー数千件で検索応答 < 300ms（DB キャッシュ温状態）
