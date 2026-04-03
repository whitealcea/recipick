# Import Design

## 1. 概要
`POST /v1/recipes/import` は URL から JSON-LD を取得して preview を返す API。**保存はしない**。

## 2. 入出力

### Request
```json
{
  "sourceUrl": "https://example.com/recipes/nikujaga"
}
```

### Response（成功時）
```json
{
  "sourceUrl": "https://example.com/recipes/nikujaga",
  "title": "肉じゃが",
  "servings": "2人分",
  "totalMinutes": 35,
  "ingredients": ["じゃがいも", "牛こま切れ肉", "玉ねぎ"],
  "systemTags": ["和食", "煮物"],
  "rawJsonLd": {"@type": "Recipe", "name": "肉じゃが"},
  "jsonLdFound": true,
  "warnings": []
}
```

### Response（JSON-LD なし）
```json
{
  "sourceUrl": "https://example.com/no-jsonld",
  "ingredients": [],
  "systemTags": [],
  "jsonLdFound": false,
  "warnings": ["json-ld script not found"]
}
```

## 3. 取得フロー
1. `sourceUrl` の URI 形式を検証
2. HTTP GET (`User-Agent: recipick-mvp/1.0`)
3. ステータス 2xx 以外は warning で終了
4. HTML を最大 5MB 読み込み
5. `script[type="application/ld+json"]` を抽出
6. JSON parse し、`@graph` を flatten
7. `@type` が `Recipe` の最初のオブジェクトを採用
8. フィールド抽出して preview 返却

## 4. パース仕様（具体）

### 4.1 title
- `name` を文字列で取得

### 4.2 servings
- `recipeYield` を文字列化

### 4.3 totalMinutes
- `totalTime`（ISO8601 duration, 例 `PT35M`）を分へ変換

### 4.4 ingredients
- `recipeIngredient` を文字列配列化
- 空要素除外
- 重複除外

### 4.5 systemTags
- `recipeCategory`
- `recipeCuisine`
- `keywords`（`,` split）
を統合し、trim + 重複除外

### 4.6 rawJsonLd
- 採用した Recipe オブジェクトをそのまま返す

## 5. fallback仕様（失敗しても登録可能）
Import API の失敗時はエラー終了を最小化し、`warnings` で返す。

例:
- `invalid sourceUrl`
- `failed to fetch HTML: ...`
- `source responded with status 404`
- `json-ld script not found`
- `json-ld found but Recipe type not found`

フロントは `jsonLdFound=false` でも新規作成フォームを開き、手入力で `POST /recipes` できるようにする。

## 6. 実装注意点
- Import API 内では DB へ書き込まない
- 外部サイト依存のためタイムアウト必須（10秒）
- HTML 取り込み上限を設ける（5MB）

