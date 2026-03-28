# imgraft

**LLM coding-agent 向けの透明画像アセット生成 CLI。**

imgraft は汎用の画像生成ツールではありません。Claude Code、Codex、Gemini CLI などの自動化フローから呼び出し、**再利用しやすい透過 PNG 素材**を安定して生成することを主目的とします。

[English README](README.md)

## 特徴

- 透過素材生成が**デフォルト ON**
- pure Go 実装 — 外部コマンド依存なし
- stdout は常に固定スキーマの JSON
- `flash` / `pro` のモデル alias をサポート
- 実モデル名は `config.toml` で差し替え可能
- 参照画像はローカルファイルと URL の両対応
- LLM agent が扱いやすい壊れにくい出力契約

## 用途

imgraft は以下のアセット生成に最適化されています。

- UI 素材
- アイコン
- ロゴたたき台
- ステッカー風アセット
- マスコット
- 他の画像や UI に合成する部品

次の用途には**向いていません**。

- 背景込みの完成作品・風景画
- フォトリアルな出力
- 文字を正確に描く用途
- 高精度な背景切り抜き（remove.bg 相当）

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

非対話形式:

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

## CLI リファレンス

### メインコマンド

```bash
imgraft "<prompt>"
```

透過画像アセットを生成します。サブコマンドなしで実行したときのデフォルトコマンドです。

### サブコマンド

```bash
# 認証
imgraft auth login
imgraft auth logout
imgraft auth whoami

# 設定
imgraft config init
imgraft config use <profile>
imgraft config refresh-models

# ユーティリティ
imgraft version
imgraft completion zsh
```

### フラグ

| フラグ | 説明 | デフォルト |
|--------|------|-----------|
| `--model` | モデル alias またはフルモデル名（`flash`、`pro`、またはフルモデル名） | `flash` |
| `--ref` | 参照画像のパスまたは URL（複数回指定可） | — |
| `--output` | 出力ファイルパス | 自動命名 |
| `--dir` | 出力ディレクトリ | `.` |
| `--no-transparent` | 透過モードを無効化 | `false` |
| `--profile` | 使用するプロファイル名 | current profile |
| `--config` | 設定ファイルパス | `~/.config/imgraft/config.toml` |
| `--pretty` | JSON 出力を整形する | `false` |
| `--verbose` | 詳細ログを stderr に出力 | `false` |
| `--debug` | デバッグログを stderr に出力 | `false` |

## モデル

imgraft は 2 つのモデル alias をサポートします。

- `flash` — 標準用途。速度と安定性を優先。
- `pro` — 高品質寄り。レートリミットや可用性エラー時は `flash` にフォールバック。

実モデル名は `~/.config/imgraft/config.toml` で管理します。

```toml
[models]
flash = "gemini-3.1-flash-image-preview"
pro   = "gemini-3-pro-image-preview"
```

`config init` と `config refresh-models` は最新のモデル一覧を取得してこれらの値を自動更新します。取得に失敗した場合は内蔵デフォルト値を使います。

### フォールバック挙動

`--model pro` を指定したとき、以下のいずれかのエラーが発生すると自動的に `flash` へフォールバックします。

- `RATE_LIMIT_EXCEEDED`
- `PERMISSION_DENIED`
- `BACKEND_UNAVAILABLE`

フォールバックは必ず JSON 出力の `warnings` フィールドに記録されます。

## 設定ファイル

### 設定

```
~/.config/imgraft/config.toml
```

```toml
current_profile    = "default"
last_used_profile  = "default"
last_used_backend  = "google_ai_studio"

default_model      = "flash"
default_output_dir = "."

[models]
flash = "gemini-3.1-flash-image-preview"
pro   = "gemini-3-pro-image-preview"
```

### 認証情報

```
~/.config/imgraft/credentials.json
```

```json
{
  "profiles": {
    "default": {
      "google_ai_studio": {
        "api_key": "YOUR_API_KEY"
      }
    }
  }
}
```

### 設定値の解決優先順位

1. CLI フラグ
2. プロファイルの config 値
3. 環境変数
4. 内蔵デフォルト値

## JSON 出力スキーマ

stdout は成功・失敗どちらも常に JSON です。

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

### スキーマのルール

- `success` が唯一の成否判定フラグ。
- `model` はフォールバックが発生した場合も含め、**実際に使用したモデル名**が入る。
- `backend` は v1 では常に `google_ai_studio`。
- 失敗時は `images` が空配列になる。
- `rate_limit` オブジェクトは常に存在する。取得できない値は各フィールドが `null`。
- `error.code` と `error.message` は失敗時に設定される。成功時は `null`。
- stderr は人間向けログ専用。stdout に混入しない。

### 出力ファイルの自動命名

自動命名は以下のパターンを使います。

```
imgraft-YYYYMMDD-HHMMSS-XXX.png
```

例: `imgraft-20260324-153012-001.png`

## 透明パイプライン

透明モード（デフォルト）では以下のパイプラインを実行します。

1. システムプロンプト適用 — 純粋な緑背景 `#00FF00` を強制
2. API 経由で画像生成
3. PNG デコード
4. 背景除去（四隅サンプリング色距離アルゴリズム）
5. 透明余白のトリミング
6. アルファチャンネル付き PNG としてエンコード
7. ディスクに保存

### 重要: 背景は imgraft が制御する

透明モードでは、ユーザーの prompt に背景指定が含まれていても、**内部のシステムプロンプトが優先**されます。緑背景を固定することで色距離ベースの背景除去が安定します。

背景込みの画像が必要な場合のみ `--no-transparent` を使ってください。

## 参照画像

`--ref` で参照画像を指定します（最大 8 枚まで繰り返し指定可能）。

```bash
imgraft "convert to icon style" --ref ./icon.png --ref ./style.png
```

対応ソース:
- ローカルファイル
- `http://` / `https://` URL

対応フォーマット: PNG / JPEG / WebP

制限:
- 最大 8 枚
- 1 枚につき最大 20 MB
- 最大 4096 × 4096 ピクセル
- private IP、ループバック、localhost URL は拒否

## エラーコード一覧

| コード | 説明 |
|--------|------|
| `INVALID_ARGUMENT` | CLI 引数が不正 |
| `AUTH_REQUIRED` | 認証情報が見つからない |
| `AUTH_INVALID` | API キーが無効 |
| `FILE_NOT_FOUND` | 参照ファイルが見つからない |
| `FILE_READ_FAILED` | ファイル読み込みに失敗 |
| `UNSUPPORTED_IMAGE_FORMAT` | 非対応の画像フォーマット |
| `IMAGE_TOO_LARGE` | サイズまたは解像度の制限超過 |
| `INVALID_IMAGE` | 画像をデコードできなかった |
| `REFERENCE_FETCH_FAILED` | 参照 URL の取得に失敗 |
| `REFERENCE_TIMEOUT` | 参照 URL の取得がタイムアウト |
| `REFERENCE_REDIRECT_LIMIT_EXCEEDED` | リダイレクト回数上限超過 |
| `REFERENCE_URL_FORBIDDEN` | private / localhost URL は使用不可 |
| `OUTPUT_DIR_CREATE_FAILED` | 出力ディレクトリを作成できなかった |
| `FILE_WRITE_FAILED` | 出力ファイルの書き込みに失敗 |
| `FILE_ALREADY_EXISTS` | 出力ファイルが既に存在する |
| `INVALID_OUTPUT_PATH` | 出力パスが不正 |
| `MODEL_RESOLUTION_FAILED` | モデル名を解決できなかった |
| `BACKEND_UNAVAILABLE` | バックエンド API が利用不可 |
| `RATE_LIMIT_EXCEEDED` | API レートリミット超過 |
| `INTERNAL_ERROR` | 予期せぬ内部エラー |

## 開発者向け

### ビルド

```bash
go build ./cmd/imgraft/
```

### テスト

```bash
# ユニット・インテグレーションテスト
go test ./...

# 単一パッケージ
go test ./internal/imageproc/

# E2E テスト（実 API キー必要）
RUN_E2E=1 go test ./...
```

### リント

```bash
go vet ./...
```

### バージョン情報付きビルド

```bash
go build \
  -ldflags "-X github.com/youyo/imgraft/internal/cli.Version=1.0.0 \
            -X github.com/youyo/imgraft/internal/cli.Commit=$(git rev-parse --short HEAD) \
            -X github.com/youyo/imgraft/internal/cli.Date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  ./cmd/imgraft/
```

## 参考ドキュメント

- プロダクト仕様書: `docs/specs/SPEC.md`
- agent 向け使い方ガイド: `skills/imgraft/SKILL.md`

## ライセンス

MIT
