# Deployment Guide (Vercel + Render)

## 1. 前提
- Frontend: `frontend` (React + Vite) を Vercel にデプロイ
- Backend: `backend` (Go + chi) を Render にデプロイ
- DB は Supabase PostgreSQL を利用（`DATABASE_URL` を Render に設定）

推奨順序:
1. Backend (Render) を先にデプロイして URL を確定する
2. Frontend (Vercel) に `VITE_API_BASE_URL` を設定してデプロイする

## 2. Backend を Render にデプロイ

### 2.1 Render サービス作成
1. Render Dashboard で `New +` -> `Web Service`
2. 対象リポジトリを接続
3. 設定:
   - `Root Directory`: `backend`
   - `Build Command`: `go build -o bin/api ./cmd/api`
   - `Start Command`: `./bin/api`
   - `Health Check Path`: `/health`

### 2.2 環境変数
Render の Environment に以下を設定:
- `DATABASE_URL`: Supabase の接続文字列（`sslmode=require` を含める）
- `AUTH_JWT_ISSUER`: 例 `https://<supabase-project-ref>.supabase.co/auth/v1`
- `AUTH_JWT_JWKS_URL`: 例 `https://<supabase-project-ref>.supabase.co/auth/v1/.well-known/jwks.json`
- `AUTH_JWT_AUDIENCE`: 例 `authenticated`（利用しない場合は未設定可）
- `PORT`: 未設定で可（Render が自動で注入）

### 2.3 デプロイ確認
- デプロイ後、`https://<render-service>.onrender.com/health` が `ok` を返すこと
- API ベース URL は `https://<render-service>.onrender.com/v1`

## 3. Frontend を Vercel にデプロイ

### 3.1 Vercel プロジェクト作成
1. Vercel Dashboard で `Add New...` -> `Project`
2. 対象リポジトリを Import
3. 設定:
   - `Root Directory`: `frontend`
   - Framework Preset: `Vite`（自動検出）
   - Install Command: `npm ci`
   - Build Command: `npm run build`
   - Output Directory: `dist`

### 3.2 環境変数
Vercel の Environment Variables に以下を設定:
- `VITE_API_BASE_URL=https://<render-service>.onrender.com/v1`
- `VITE_SUPABASE_URL=https://<supabase-project-ref>.supabase.co`
- `VITE_SUPABASE_ANON_KEY=<supabase-anon-key>`

### 3.3 Supabase Auth (Google OAuth) 設定
Supabase Dashboard の `Authentication` -> `URL Configuration` / `Providers` で以下を設定:
- Google Provider を有効化し、Google Cloud 側の Client ID / Secret を登録
- Site URL: `https://<vercel-domain>`
- Redirect URLs:
  - `https://<vercel-domain>/login`
  - `http://localhost:5173/login`（ローカル開発用）

### 3.3 デプロイ確認
- Vercel URL を開き、レシピ一覧取得・作成・更新・削除が通ること
- ブラウザ DevTools で API リクエスト先が Render URL になっていること

## 4. CORS の注意点（必須）
現状の backend は `backend/internal/api/server.go` で `AllowedOrigins` がローカルホスト固定です。  
Vercel の公開 URL から API を叩くには、以下のどちらかが必要です。

1. `AllowedOrigins` に Vercel の URL を追加して再デプロイ
2. 環境変数ベース（例: `CORS_ALLOWED_ORIGINS`）で許可 Origin を切り替える実装に変更して再デプロイ

## 5. トラブルシュート
- `CORS error`:
  - backend 側の `AllowedOrigins` 未設定が原因。`https://<project>.vercel.app` を許可する
- `500 / DB 接続エラー`:
  - `DATABASE_URL` の誤り、または `sslmode=require` の不足を確認
- `404 on /v1/...`:
  - Frontend の `VITE_API_BASE_URL` が `/v1` 付きになっているか確認
- `401 invalid_token`:
  - `Authorization` ヘッダに Bearer JWT が付与されているか確認
  - `AUTH_JWT_ISSUER` / `AUTH_JWT_JWKS_URL` / `AUTH_JWT_AUDIENCE` が Supabase 設定と一致するか確認
- `Build failed (frontend)`:
  - Node バージョン要件 `>=24 <25` を満たすよう Vercel の Node 設定を合わせる
