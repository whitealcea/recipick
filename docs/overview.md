# Recipick Overview

## 1. プロジェクト概要
Recipick は「個人用レシピ URL ブックマークアプリ」です。ユーザーはレシピページの URL を登録し、可能であれば JSON-LD からタイトル・材料・タグ候補を取得して、あとから検索できます。

- フロントエンド: React + TypeScript + Vite
- バックエンド: Go + chi
- DB: PostgreSQL (Supabase)
- Frontend deploy: Vercel
- Backend deploy: Render
- Deploy 手順: `docs/deployment.md`

このプロジェクトは **MVP（最小実装）優先** です。複雑な権限やレコメンド機能は対象外とし、登録・更新・検索の品質を最短で担保します。

## 2. ゴール

### 2.1 ユーザー価値
- 気になるレシピ URL を一か所に集約できる
- 材料とタグで実用的に絞り込める
- JSON-LD が取れるページは入力を省力化できる

### 2.2 プロダクトゴール
- URLベースの一元管理
- 材料・タグ・所要時間で AND 検索
- JSON-LD 取得失敗時でも手動登録可能
- 同一ユーザー内の URL 重複登録を防止

## 3. 想定ユースケース

### UC-01: URL から下書きを作る
1. ユーザーが `POST /recipes/import` に URL を送る
2. サーバーが HTML から `application/ld+json` を抽出
3. `@type=Recipe` の JSON-LD を見つけた場合、preview を返す
4. 見つからない場合も `200` で warning を返す

### UC-02: レシピを保存する
1. ユーザーが preview を確認・編集
2. `POST /recipes` で保存
3. 同一 `user_id + source_url` が存在する場合 `409` を返す

### UC-03: レシピを検索する
1. 一覧画面で `q, tags, ingredients, maxMinutes` を指定
2. `GET /recipes` が AND 条件で絞り込む
3. 結果を新しい順に表示

### UC-04: 編集・削除する
- `PATCH /recipes/{id}` で部分更新
- `DELETE /recipes/{id}` で削除

## 4. MVP スコープ

### 4.1 含む
- レシピ CRUD
- Import preview
- AND 検索
- `Authorization: Bearer <JWT>` によるユーザー認証

### 4.2 含まない
- OAuth / 認証基盤
- 画像保存
- レシピ手順の構造化保存
- 同義語検索・全文検索の高度最適化

## 5. 成功条件（実装完了の判定）
- API 6本が OpenAPI 仕様どおりに動作
- DB に `recipes`, `recipe_ingredients`, `recipe_tags` が作成される
- URL 重複時に `409 duplicate_source_url`
- 検索が全条件 AND で動く
- Import API が保存せず preview のみ返す
