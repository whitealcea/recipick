# recipick

レシピ管理アプリケーションです。`backend` (Go) と `frontend` (React + Vite) で構成されています。

## ディレクトリ構成

- `backend`: API サーバー
- `frontend`: Web フロントエンド
- `docs`: 設計・仕様ドキュメント

## セットアップ

### 1. Frontend

```bash
cd frontend
npm install
npm run dev
```

### 2. Backend

```bash
cd backend
go mod download
go run ./cmd/api
```

## ドキュメント

- 全体概要: `docs/overview.md`
- アーキテクチャ: `docs/architecture.md`
- API 定義: `docs/api.yaml`

