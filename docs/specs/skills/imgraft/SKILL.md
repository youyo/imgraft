# imgraft SKILL

## 目的

`imgraft` は、LLM coding-agent から **再利用可能な透過画像アセット** を生成するための CLI です。

このスキルは、エージェントが `imgraft` を安定して使うためのルールを定義します。

---

## 位置づけ

`imgraft` は汎用画像生成ツールではありません。

次の用途を優先します。

- UI 素材
- アイコン
- ロゴたたき台
- マスコット
- ステッカー風アセット
- 他の画像や UI に合成する部品

背景込みの完成作品より、**透明 PNG 素材**を安定して作ることを目的とします。

---

## 最重要ルール

### 1. prompt は短くする

良い prompt は、**被写体 + スタイル + 制約** の短い組み合わせです。

よい例:

```text
blue robot mascot, simple, modern
tech logo, geometric, high contrast
cute cat sticker, pastel colors
settings icon, line style
```

避けたい例:

```text
a beautiful robot standing in a futuristic city at night with neon buildings and dramatic reflections
```

長すぎる prompt は、素材生成用途ではノイズになりやすいです。

---

### 2. 背景を指定しない

transparent モードでは、背景は `imgraft` 側が制御します。

避けるべき例:

```text
robot in a forest
city background
space scene
night skyline
```

この種の背景指定は、透過素材生成の成功率を下げます。

---

### 3. まずは transparent 前提で使う

`imgraft` は transparent がデフォルトです。通常は `--no-transparent` を付けません。

通常例:

```bash
imgraft "blue robot mascot, simple, modern"
```

例外的に背景込み画像が必要な場合だけ:

```bash
imgraft "scene illustration" --no-transparent
```

---

### 4. 出力は必ず JSON をパースする

`imgraft` の stdout は JSON です。エージェントは人間向けログではなく JSON を正としてください。

見るべき項目:

- `success`
- `model`
- `images[0].path`
- `warnings`
- `error.code`
- `error.message`

---

## 推奨コマンドパターン

### 基本

```bash
imgraft "blue robot mascot, simple, modern"
```

### 参照画像付き（ローカルファイル）

```bash
imgraft "convert to icon style" --ref ./input.png
```

### 参照画像付き（URL）

```bash
imgraft "convert to sticker style" --ref https://example.com/input.png
```

### 参照画像複数枚

```bash
imgraft "redesign to match style" --ref ./subject.png --ref ./style.png
```

### 高品質寄り

```bash
imgraft "dragon logo, bold, geometric" --model pro
```

### 出力先ディレクトリ指定

```bash
imgraft "cute cat icon, pastel" --dir ./out
```

### 単一ファイル名指定

```bash
imgraft "app icon, line style" --output ./asset.png
```

### JSON 整形出力

```bash
imgraft "settings icon" --pretty
```

---

## モデル選択ルール

通常は `flash` を使って問題ありません。

- `flash`: 既定。速度と安定性優先
- `pro`: 高品質寄り。失敗時は `flash` fallback の可能性あり

推奨:

- まず `flash`
- 品質が欲しいときだけ `pro`

`model` フィールドには、最終的に使われた実モデル名が入ります。`pro` 指定でも fallback が起きた場合は `flash` 系の実モデル名になります。

---

## 参照画像の使い方

参照画像は以下を推奨します。

- 被写体が分かりやすい
- 枠や余計な UI がない
- 画質が極端に悪くない
- 背景が極端に複雑すぎない

避けたいもの:

- GIF、SVG、PDF
- 20MB 超の巨大画像
- localhost / private IP の URL

---

## 失敗時の扱い

### `success == false`

次を確認します。

- `error.code`
- `error.message`

主な対応方針:

| コード | 対応 |
|--------|------|
| `RATE_LIMIT_EXCEEDED` | 少し待って再実行 |
| `FILE_NOT_FOUND` | 参照画像パスを修正 |
| `REFERENCE_FETCH_FAILED` | URL を見直す |
| `REFERENCE_URL_FORBIDDEN` | private URL を使っていないか確認 |
| `UNSUPPORTED_IMAGE_FORMAT` | PNG/JPEG/WebP に変換して再実行 |
| `AUTH_REQUIRED` | `imgraft auth login` を実行 |
| `AUTH_INVALID` | `imgraft auth whoami` で API key を確認 |

---

## warnings の扱い

代表例:

- `model fallback: pro -> flash` — pro を指定したが flash を使用した
- `background removal quality may be low` — 透過品質が低い可能性あり
- `using http reference URL` — https の使用を推奨

`warnings` は即失敗ではありません。必要に応じて再実行条件の判断材料に使います。

---

## エージェント向け判断ルール

### 透過素材を作りたいとき

そのまま `imgraft` を使う。

### 背景込みの作品が欲しいとき

`imgraft` より他の画像生成手段の方が向いている可能性が高い。どうしても `imgraft` を使う場合だけ `--no-transparent` を付ける。

### ロゴ・アイコン・UI 素材

`imgraft` に最も向いている。積極的に使う。

### 写真品質や高精度切り抜き

`imgraft` は向いていない。純粋な asset generator として扱う。

---

## 実行後の取り回し

エージェントは次の流れで扱う。

1. `imgraft` を実行
2. stdout JSON を parse
3. `success` を確認
4. `images[0].path` を取得
5. 後続処理でそのパスを使う

---

## JSON 出力スキーマ参照

```json
{
  "success": true,
  "model": "gemini-3.1-flash-image-preview",
  "backend": "google_ai_studio",
  "images": [
    {
      "index": 0,
      "path": "/abs/path/imgraft-20260324-153012-001.png",
      "filename": "imgraft-20260324-153012-001.png",
      "width": 1024,
      "height": 1024,
      "mime_type": "image/png",
      "sha256": "abc123...",
      "transparent_applied": true
    }
  ],
  "rate_limit": {
    "provider": "google_ai_studio",
    "limit_type": null,
    "requests_limit": null,
    "requests_remaining": null,
    "requests_used": null,
    "reset_at": null,
    "retry_after_seconds": null
  },
  "warnings": [],
  "error": {
    "code": null,
    "message": null
  }
}
```

---

## 推奨テンプレート

```text
Generate a reusable transparent visual asset with imgraft.

Rules:
- Keep the prompt short.
- Focus on subject and style.
- Do not describe the background.
- Prefer simple, high-contrast, reusable assets.
- Parse stdout as JSON and use images[0].path.
```

---

## やってはいけないこと

- 長文 prompt を投げる
- 背景を詳細に指示する（transparent モード時）
- stdout の JSON ではなく stderr を正として扱う
- 参照画像エラーを無視する
- `warnings` を `success: false` と混同する
- 完成作品向けの複雑なシーン生成に `imgraft` を無理に使う
