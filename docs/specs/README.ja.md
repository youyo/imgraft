# imgraft

imgraft は、LLM coding-agent 向けの **transparent asset generator** です。

汎用の画像生成 CLI ではありません。主目的は、Claude Code、Codex、Gemini CLI などの自動化フローから呼び出し、**再利用しやすい透過 PNG 素材**を安定して生成することです。

## 特徴

- 透過素材生成がデフォルト
- 純粋な Go 実装
- 外部コマンド依存なし
- stdout は固定スキーマの JSON
- `flash` / `pro` のモデル alias をサポート
- 実モデル名は config で差し替え可能
- 参照画像はローカルファイルと URL の両対応
- coding-agent が扱いやすい壊れにくい出力契約

## 位置づけ

imgraft は「完成作品を一発生成する CLI」ではなく、次の用途を狙っています。

- UI 素材
- アイコン
- ロゴたたき台
- ステッカー風アセット
- マスコット
- 他の生成物に合成する部品画像

## インストール

### Homebrew

```bash
brew install youyo/tap/imgraft
```

### go install

```bash
go install github.com/youyo/imgraft/cmd/imgraft@latest
```

## クイックスタート

### 1. 初期設定

対話形式:

```bash
imgraft config init
```

非対話:

```bash
imgraft config init --profile default --api-key YOUR_API_KEY
```

### 2. 認証確認

```bash
imgraft auth whoami
```

### 3. 透過素材を生成

```bash
imgraft "blue robot mascot, simple, modern"
```

### 4. 参照画像付きで生成

```bash
imgraft "convert to icon style" --ref ./input.png
```

### 5. 透過を無効化

```bash
imgraft "scene illustration" --no-transparent
```

## 重要な挙動

### transparent はデフォルト ON

`imgraft` はデフォルトで透過素材生成パイプラインを使います。

- 緑背景を system prompt で固定
- pure Go の背景除去を適用
- trim で余白を削除
- PNG で保存

透過を無効にしたい場合のみ `--no-transparent` を使います。

### 背景はユーザーが制御しない

transparent モードでは、背景は CLI が制御します。ユーザーの prompt に背景指定が含まれていても、アセット生成の成功率を優先して内部制約が適用されます。

### 短い prompt を推奨

よい例:

```bash
imgraft "blue robot mascot, flat, minimal"
imgraft "tech logo, geometric, high contrast"
imgraft "cute cat sticker, pastel"
```

避けたい例:

```bash
imgraft "a beautiful robot standing in a futuristic city at night with neon buildings and reflections"
```

## コマンド

### メイン

```bash
imgraft "<prompt>"
```

### サブコマンド

```bash
imgraft auth login
imgraft auth logout
imgraft auth whoami

imgraft config init
imgraft config use <profile>
imgraft config refresh-models

imgraft version
imgraft completion zsh
```

## 主なフラグ

```text
--model <flash|pro|full_model_name>
--ref <path_or_url>
--output <file>
--dir <directory>
--no-transparent
--profile <name>
--config <path>
--pretty
--verbose
--debug
```

## モデル

`imgraft` は `flash` と `pro` の alias をサポートします。

- `flash`: 標準用途
- `pro`: 高品質寄り

実モデル名は `~/.config/imgraft/config.toml` の `[models]` で管理します。

```toml
[models]
flash = "gemini-3.1-flash-image-preview"
pro = "gemini-3-pro-image-preview"
```

`config init` と `config refresh-models` は、利用可能モデル一覧の取得を試みてこの値を更新します。取得に失敗した場合は内蔵デフォルト値を使います。

## 設定ファイル

### config

```text
~/.config/imgraft/config.toml
```

### credentials

```text
~/.config/imgraft/credentials.json
```

例:

```toml
current_profile = "default"
last_used_profile = "default"
last_used_backend = "google_ai_studio"

default_model = "flash"
default_output_dir = "."

[models]
flash = "gemini-3.1-flash-image-preview"
pro = "gemini-3-pro-image-preview"
```

## JSON 出力

stdout は常に JSON です。

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
      "sha256": "...",
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

## rate_limit の扱い

`rate_limit` オブジェクトは常に存在します。値が取得できない場合は各フィールドが `null` になります。推測値は入れません。

## 参照画像

対応:

- ローカルファイル
- `http://` / `https://` URL

制限:

- 最大 8 枚
- PNG / JPEG / WebP
- 20MB まで
- 4096x4096 まで
- private IP / localhost 系 URL は拒否

## 注意点

- 背景込みの作品生成には向きません
- 文字を正確に入れたい用途には向きません
- 写真の高精度切り抜き用途には向きません
- 透過素材向けに最適化されているため、背景表現の自由度は抑えています

## 開発方針

- Go 単一バイナリ配布
- 外部コマンド依存なし
- coding-agent が扱いやすい JSON 契約を最優先

## 参考ドキュメント

- 詳細仕様: `docs/specs/SPEC.md`
- agent 向け使い方: `skills/imgraft/SKILL.md`

