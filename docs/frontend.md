# Frontend Implementation Guide

## 0. Node バージョン管理（Volta）

このプロジェクトの frontend は Volta で Node バージョンを固定する。

初回セットアップ:

```bash
brew install volta
cd frontend
volta install node@24.9.0
npm install
```

`frontend/package.json` の `volta.node` と `engines.node` を変更すると、プロジェクトの Node バージョンを更新できる。

## 1. 目的
React + TypeScript + Vite で、レシピ登録・検索・編集を実装する。状態管理は以下に限定する。

- UI state: `useState`
- Server state: `TanStack Query`
- グローバル state ライブラリは使わない

## 2. 推奨ディレクトリ構成

```text
frontend/
  src/
    api/
      client.ts
      recipes.ts
    components/
      RecipeForm.tsx
      TagInput.tsx
      IngredientInput.tsx
      SearchForm.tsx
      RecipeList.tsx
    hooks/
      useRecipesQuery.ts
      useRecipeDetailQuery.ts
      useImportPreviewMutation.ts
      useCreateRecipeMutation.ts
      useUpdateRecipeMutation.ts
      useDeleteRecipeMutation.ts
    pages/
      RecipeListPage.tsx
      RecipeNewPage.tsx
      RecipeDetailPage.tsx
      RecipeEditPage.tsx
    types/
      api.ts
      form.ts
    lib/
      queryClient.ts
      auth.ts
    App.tsx
    main.tsx
```

## 3. API 呼び出し設計

## 3.1 共通クライアント
`api/client.ts`

- `baseURL = import.meta.env.VITE_API_BASE_URL`
- `Authorization: Bearer <JWT>` をヘッダ付与（トークンは Supabase セッションから取得）
- 非2xxは `ApiError` として throw

## 3.2 認証（Supabase Auth）
- `@supabase/supabase-js` を利用
- `AuthProvider` が `supabase.auth.getSession()` + `onAuthStateChange` でセッション監視
- ログイン手段は Google OAuth のみ
- 未ログイン時は `/login` へリダイレクト
- ログイン成功後は `next` クエリを優先して遷移

必要な環境変数:
- `VITE_SUPABASE_URL`
- `VITE_SUPABASE_ANON_KEY`
- `VITE_API_BASE_URL`

## 3.3 TanStack Query キー
- `['recipes', { q, tags, ingredients, maxMinutes, limit, offset }]`
- `['recipe', recipeId]`

## 3.4 Query/Mutation 例
- 一覧取得: `useQuery`（フォーム入力を query params へ反映）
- インポート: `useMutation`（結果をフォーム初期値へ反映）
- 作成: `useMutation` + 成功時 `invalidateQueries(['recipes'])`
- 更新: `useMutation` + `invalidateQueries(['recipe', id])` と一覧
- 削除: `useMutation` + 一覧 invalidate + 一覧画面に戻す

## 4. 型定義方針

### 4.1 API 契約型
`src/types/api.ts` に API の I/O を定義し、`fetch` レイヤーで必ず使う。

必須型:
- `Recipe`
- `RecipeTag`
- `ImportPreview`
- `SearchRecipesResponse`
- `ApiError`

### 4.2 Form 型
API 型をそのままフォームで持たず、フォーム専用型を分離する。

例:
- `RecipeFormValues`:
  - `sourceUrl: string`
  - `title: string`
  - `note: string`
  - `servings: string`
  - `totalMinutes: string` (入力中は文字列で保持)
  - `ingredientsText: string` (改行区切り)
  - `tags: { name: string; tagType: 'system' | 'user' }[]`

## 5. 画面一覧（MVP）

### 5.1 `/recipes` 一覧
- 検索フォーム
  - `q`（title+note 部分一致）
  - `tags`（カンマ区切り）
  - `ingredients`（カンマ区切り）
  - `maxMinutes`
- 検索結果一覧

### 5.2 `/recipes/new` 新規作成
- URL入力 + Import preview ボタン
- preview 結果をフォームに反映
- 手動修正して保存

### 5.3 `/recipes/:id` 詳細
- レシピ本文表示
- 編集 / 削除導線

### 5.4 `/recipes/:id/edit` 編集
- 既存値ロード
- 部分更新送信

## 6. 実装ルール
- 文字列 trim は submit 前に実施
- 空材料/空タグは送信しない
- API エラーは画面上部にメッセージ表示
- 楽観更新は行わず、Mutation 成功後に再取得（MVP優先）
