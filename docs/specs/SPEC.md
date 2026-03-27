# imgraft 超詳細仕様書

## 1. 文書の目的

本書は、`imgraft` を **LLM coding-agent 向けの透明画像アセット生成 CLI** として実装するための超詳細仕様書である。実装者は本書のみを参照して、Go による v1 の実装、README 整備、スキル作成、リリース設定まで進められることを目的とする。

`imgraft` は汎用画像生成ツールではない。位置づけは次のとおり。

> **imgraft = transparent asset generator for automation pipelines**

主用途は以下。

- Claude Code / Codex / Gemini CLI などの coding-agent からの利用
- UI 素材、アイコン、マスコット、ロゴ、ステッカー風アセットの生成
- 背景透過 PNG の継続的な自動生成
- 参照画像を使った軽量なスタイル変換・素材化

非目的は以下。

- 風景画、写真、複雑な構図、完成作品の一発生成
- 動画処理
- Photoshop / remove.bg 相当の高精度切り抜き
- Vertex AI 対応

---

## 2. 設計原則

### 2.1 最重要原則

1. **透明アセット生成をデフォルトにする**
2. **外部コマンド依存を完全に排除する**
3. **stdout は常に JSON**
4. **スキーマは固定し、欠損ではなく null で表現する**
5. **LLM agent が壊れにくい CLI 契約を優先する**
6. **シンプルさを優先し、v1 は Google AI Studio のみ対応する**
7. **将来拡張性は保つが、今使わない複雑さは持ち込まない**

### 2.2 実装言語

- Go 1.23 以上を想定
- pure Go 実装
- FFmpeg, ImageMagick, Python などの外部依存なし

### 2.3 参考にするリポジトリ構成方針

`logvalet` は `cmd/logvalet`, `docs/specs`, `internal`, `plans`, `skills/logvalet`, `.goreleaser.yaml`, `CLAUDE.md`, `README.md`, `README.ja.md` を持つ構成になっている。`bundr` も `cmd`, `docs`, `internal`, `plans`, `scripts`, `.goreleaser.yaml`, `README.md`, `README.ja.md`, `action.yml` を持つ。`imgraft` もこの系統に寄せる。 citeturn892709view0turn892709view2

---

## 3. スコープ

### 3.1 v1 に含めるもの

- 単一コマンドの画像生成 CLI
- Google AI Studio API キー認証
- モデル alias 解決 (`flash`, `pro`)
- `config init` 時のモデル自動解決
- `config refresh-models`
- ローカルファイルと URL の参照画像対応
- transparent デフォルト ON
- pure Go 背景除去
- pure Go trim
- JSON 契約固定
- README
- SKILL.md
- GoReleaser 設定

### 3.2 v1 に含めないもの

- Vertex AI
- 動画
- PDF / SVG / GIF 入力
- 背景色変更
- 背景除去パラメータのユーザー公開
- `--overwrite`
- 画像複数枚生成
- GUI
- GitHub Action

---

## 4. CLI 名称と配置

- リポジトリ名: `imgraft`
- バイナリ名: `imgraft`
- 読み: イムグラフト
- 設定ディレクトリ: `~/.config/imgraft/`
- スキル配置: `skills/imgraft/SKILL.md`

---

## 5. 想定リポジトリ構成

```text
imgraft/
  cmd/
    imgraft/
      main.go
  docs/
    specs/
      SPEC.md
  internal/
    app/
      run.go
    cli/
      root.go
      auth.go
      config.go
      version.go
      completion.go
    config/
      loader.go
      saver.go
      types.go
      profiles.go
    auth/
      login.go
      whoami.go
      logout.go
      validate.go
    backend/
      studio/
        client.go
        models.go
        generate.go
        ratelimit.go
        errors.go
    model/
      resolver.go
      refresh.go
      defaults.go
    reference/
      loader.go
      local.go
      remote.go
      validate.go
      types.go
    prompt/
      builder.go
      system.go
    imageproc/
      background.go
      trim.go
      hash.go
      encode.go
      decode.go
      inspect.go
    output/
      naming.go
      save.go
      json.go
      errors.go
    ratelimit/
      types.go
      headers.go
    errs/
      codes.go
      map.go
    runtime/
      env.go
      paths.go
      clock.go
  plans/
    roadmap.md
  skills/
    imgraft/
      SKILL.md
  .goreleaser.yaml
  README.md
  README.ja.md
  CLAUDE.md
  go.mod
  go.sum
```

### 5.1 構成方針

- `cmd/` はエントリポイントのみ
- 業務ロジックは `internal/` に閉じ込める
- 仕様書は `docs/specs/` に置く
- skills は単独ディレクトリ配下に置く
- `README.md` と `README.ja.md` を併置する

---

## 6. コマンド体系

### 6.1 メインコマンド

```bash
imgraft "<prompt>"
```

画像生成本体にはサブコマンドを設けない。`generate` や `edit` のようなサブコマンドは不要。参照画像の有無で generate/edit 相当を兼ねる。

### 6.2 サブコマンド

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

### 6.3 採用しないサブコマンド

- `imgraft generate`
- `imgraft edit`
- `imgraft models`
- `imgraft config list`

### 6.4 グローバルフラグ

```text
--model <value>
--ref <path_or_url>    # 複数回指定可
--output <file>
--dir <directory>
--no-transparent
--profile <name>
--config <path>
--pretty
--verbose
--debug
```

### 6.5 代表的な利用例

```bash
imgraft "blue robot mascot"
imgraft "convert to icon style" --ref ./input.png
imgraft "tech logo, minimal" --model pro
imgraft "scene illustration" --no-transparent
imgraft "cute cat sticker" --dir ./out
imgraft "avatar asset" --output ./asset.png
```

---

## 7. 設定ファイルと認証情報

### 7.1 保存パス

- 設定: `~/.config/imgraft/config.toml`
- 認証情報: `~/.config/imgraft/credentials.json`

### 7.2 config.toml スキーマ

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

### 7.3 credentials.json スキーマ

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

### 7.4 方針

- 複数 profile を許可する
- 初期 profile 名は `default`
- 最後に成功した profile を保持する
- backend は v1 では `google_ai_studio` 固定だが、構造上は backend ごとのネストを残す
- API key は初版では平文保存でよい

### 7.5 設定値の解決優先順位

実行時の設定解決順:

1. CLI フラグ
2. 指定 profile の config 値
3. 環境変数
4. 既定値

認証解決順:

1. `--profile`
2. `current_profile`
3. `last_used_profile`

---

## 8. 認証仕様

### 8.1 サポート backend

- v1: `google_ai_studio` のみ
- 将来 reserved: `vertex_ai`

### 8.2 `auth login`

対話フロー:

1. profile 名入力（default を既定値にする）
2. API key 入力
3. 軽い疎通確認
4. 成功時のみ `credentials.json` に保存
5. `current_profile`, `last_used_profile`, `last_used_backend` を更新

### 8.3 `auth logout`

- current profile の `google_ai_studio` 認証を削除
- profile 自体は削除しない

### 8.4 `auth whoami`

表示内容:

- current profile
- last used profile
- last used backend
- current profile で利用可能な backend
- API key は末尾 4 文字のみ表示

表示例:

```text
Profile: default
Last used backend: google_ai_studio

Available backends:
- google_ai_studio (api_key: ****abcd)
```

---

## 9. モデル解決仕様

### 9.1 基本方針

- CLI の正式 alias は `flash`, `pro`
- 実モデル名は `config.toml` の `[models]` で保持する
- ユーザーは config を編集することで CLI 更新なしにモデル差し替えできる
- フルモデル名直接指定も許可する

### 9.2 `--model` の受理値

- 未指定
- `flash`
- `pro`
- フルモデル名

### 9.3 解決順

1. `--model` 未指定なら `config.default_model`
2. 値が `flash` なら `config.models.flash`
3. 値が `pro` なら `config.models.pro`
4. それ以外はフルモデル名としてそのまま使う

### 9.4 fallback

- `pro` 指定時のみ fallback を許可する
- fallback 先は `config.models.flash`
- 対象エラー:
  - `RATE_LIMIT_EXCEEDED`
  - `PERMISSION_DENIED`
  - `BACKEND_UNAVAILABLE`
- fallback 発生時は `warnings` に必ず記録する
- JSON の `model` には最終的に使った実モデル名を出す

### 9.5 built-in defaults

`config init` や `config refresh-models` が失敗したときの保険として、CLI は以下を内蔵してよい。

```text
flash -> gemini-3.1-flash-image-preview
pro   -> gemini-3-pro-image-preview
```

ただし、これは **正本ではなく fallback 用シード** である。

### 9.6 `config init` 時のモデル初期化

Google の Gemini API には `models` エンドポイントがあり、利用可能モデルを列挙できる。`imgraft config init` は API key が使える場合、この一覧取得を試行し、`flash` / `pro` の実体モデルを config に書き込む。失敗時は built-in defaults で埋める。 citeturn167325view0

### 9.7 `config refresh-models`

- current profile の API key でモデル一覧を取得する
- `flash` / `pro` を再解決して `config.toml` の `[models]` を上書きする
- `default_model` は変更しない
- 失敗した場合は既存 config を保持し、コマンドはエラー終了する

### 9.8 モデル一覧からの解決ヒューリスティック

`flash` 解決優先順位:

1. 画像生成または画像編集に使える
2. 名前に `flash` を含む
3. 名前に `image` を含む
4. より安定した候補を優先する

`pro` 解決優先順位:

1. 画像生成または画像編集に使える
2. 名前に `pro` を含む
3. 名前に `image` を含む
4. より安定した候補を優先する

### 9.9 AI Studio のレートリミットに関する前提

Gemini API の現在有効な rate limits は AI Studio 側で確認する案内になっており、モデルや tier によって変動する。したがって `imgraft` は上限値をハードコードしない。実行時 JSON にはレスポンスから確実に取得できた値のみを入れ、取れないものは `null` にする。 citeturn167325view0

---

## 10. 参照画像仕様

### 10.1 入力ソース

- ローカルファイル
- `http://` URL
- `https://` URL

### 10.2 非対応

- `data:` URL
- `file:` URL
- `ftp:`
- `s3:`
- `gs:`
- stdin

### 10.3 フラグ

```bash
--ref <path_or_url>
```

複数回指定可。

### 10.4 ルール

- 最大 8 枚
- 指定順をそのまま保持する
- ローカルと URL が混在しても順序を維持する
- 1 枚でも不正なら fail-fast で API 呼び出し前に終了する

### 10.5 対応フォーマット

- PNG
- JPEG/JPG
- WebP

### 10.6 検証方法

- 拡張子だけでは信用しない
- 実デコードできるかで判定する
- URL でも Content-Type のみでは決めず、最終的にデコード成功で確定する

### 10.7 サイズ制限

- 最大ファイルサイズ: 20 MB
- 最大解像度: 4096 x 4096
- 自動リサイズは v1 ではしない

### 10.8 URL 取得ルール

- GET してからローカル統一形式へ落とし込む
- リダイレクトは最大 3 回
- 接続タイムアウト: 5 秒
- 全体タイムアウト: 20 秒
- `localhost`, loopback, private IP, link-local, metadata 系アドレスは拒否する
- `http://` は許可するが warning を出してよい

### 10.9 内部統一型

各参照画像は内部で以下を持つ。

- `source_type` (`file` or `url`)
- `original_input`
- `local_cached_path`
- `filename`
- `mime_type`
- `width`
- `height`
- `size_bytes`

---

## 11. プロンプト仕様

### 11.1 基本方針

`imgraft` は user prompt をそのまま API に投げない。`transparent` デフォルト設計により、system prompt でアセット生成用途へ強く誘導する。

### 11.2 transparent ON 時の system prompt

```text
Generate a single isolated subject asset for compositing.

Use a solid pure green background.
Do not use gradients.
Do not use shadows on the background.

Do not include background objects, scenery, environment, text, borders, or frames.

Center the subject.
Keep the full silhouette visible and cleanly separated from the background.

Ensure strong color contrast between subject and background.
```

### 11.3 transparent OFF 時

`--no-transparent` が指定された場合、上記制約は弱める。少なくとも以下は外す。

- solid pure green background
- no background objects
- silhouette separation

### 11.4 背景指示の扱い

transparent ON 時、ユーザーの prompt に背景指定が含まれていても、CLI は背景固定を優先する。README と SKILL.md にその旨を明記する。

---

## 12. transparent 仕様

### 12.1 基本方針

- transparent は **デフォルト ON**
- 無効化は `--no-transparent`
- 背景色は固定でユーザー変更不可
- 外部コマンドは使わない
- pure Go の簡易背景除去を適用する

### 12.2 パイプライン

1. system prompt 適用
2. 画像生成
3. PNG デコード
4. 背景除去
5. trim
6. PNG エンコード
7. 保存

### 12.3 背景色

- 固定色: `#00FF00`
- 理由: RGB 距離で分離しやすく、pure Go 処理と相性が良い

### 12.4 背景除去アルゴリズム

アルゴリズム名:

- `corner_sampling_color_distance`

手順:

1. 画像四隅のサンプルから背景色を推定する
2. 各ピクセルとの RGB 距離を計算する
3. 距離に応じて alpha を生成する
4. 境界は hard cut ではなく段階的 alpha を適用する
5. 軽いエッジ補正を行う

### 12.5 しきい値

- 既定 threshold: `40`
- v1 では CLI フラグとして公開しない

### 12.6 trim

transparent ON 時は trim を自動適用する。

- alpha > 0 の bounding box を計算
- 周囲の完全透明領域を削除
- 出力素材として扱いやすくする

### 12.7 失敗時の扱い

原則:

- transparent の品質不足は fatal にしない
- warning で通知して保存してよい

例外:

- デコード失敗
- 背景除去により画像がほぼ消失した

この場合はエラーにする。

### 12.8 JSON 反映

出力画像ごとに `transparent_applied` を持たせる。

---

## 13. 出力ファイル仕様

### 13.1 形式

- PNG 固定

### 13.2 出力フラグ

```text
--output <file>
--dir <directory>
```

### 13.3 優先順位

1. `--output`
2. `--dir`
3. `config.default_output_dir`
4. 現在ディレクトリ `.`

### 13.4 `--output`

- 単一ファイル出力用
- v1 は画像 1 枚のみなので常に利用可能
- ファイル名・拡張子は尊重するが、最終形式は PNG 固定

### 13.5 `--dir`

- ディレクトリを指定する
- 存在しなければ作成する

### 13.6 自動命名規則

```text
imgraft-YYYYMMDD-HHMMSS-XXX.png
```

例:

```text
imgraft-20260324-153012-001.png
```

### 13.7 衝突回避

- 上書きはデフォルトで禁止
- 同名が存在する場合は連番をインクリメントして回避する
- `--overwrite` は v1 では実装しない

### 13.8 保存後に取得するメタ情報

- `width`
- `height`
- `mime_type`
- `sha256`

### 13.9 JSON のパス表現

- `path`: 絶対パス
- `filename`: ファイル名のみ

### 13.10 保存順序

1. ファイル書き込み
2. close
3. inspect
4. sha256 計算
5. JSON 出力

---

## 14. JSON 出力仕様

### 14.1 方針

- stdout は JSON のみ
- 成功/失敗どちらも同一スキーマ
- スキーマは常に固定
- 欠損ではなく `null` または空配列で表現する

### 14.2 トップレベルスキーマ

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

### 14.3 ルール

- `success` は唯一の成否判定フラグ
- `model` は実際に使った最終モデル名
- `backend` は v1 では常に `google_ai_studio`
- 失敗時 `images` は空配列
- 失敗時 `error.code`, `error.message` を設定する
- 失敗時も `rate_limit` オブジェクトは必ず出す

### 14.4 stderr

- 人間向けログ
- warning 詳細
- debug 情報

stdout には絶対にログを混ぜない。

---

## 15. rate_limit 仕様

### 15.1 基本方針

- `rate_limit` はルートごと `null` にしない
- オブジェクトは必ず存在させる
- 値が取得できないときは各フィールドを `null` にする
- 推測値を入れない
- ハードコードしない

### 15.2 スキーマ

```json
{
  "provider": "string|null",
  "limit_type": "string|null",
  "requests_limit": "number|null",
  "requests_remaining": "number|null",
  "requests_used": "number|null",
  "reset_at": "string|null",
  "retry_after_seconds": "number|null"
}
```

### 15.3 初期値

```json
{
  "provider": null,
  "limit_type": null,
  "requests_limit": null,
  "requests_remaining": null,
  "requests_used": null,
  "reset_at": null,
  "retry_after_seconds": null
}
```

### 15.4 取得元

- HTTP response headers
- SDK metadata

### 15.5 429 時

- `Retry-After` ヘッダがあれば `retry_after_seconds` に反映
- 無ければ `null`

### 15.6 実装上の注意

Gemini API の上限値はプロジェクトや tier により変動し、AI Studio で確認する案内が中心である。`imgraft` は「現在の限度値」を自前で管理しない。レスポンス由来のみ採用する。 citeturn167325view0

---

## 16. エラー設計

### 16.1 エラーコード一覧

```text
INVALID_ARGUMENT
AUTH_REQUIRED
AUTH_INVALID

FILE_NOT_FOUND
FILE_READ_FAILED
UNSUPPORTED_IMAGE_FORMAT
IMAGE_TOO_LARGE
INVALID_IMAGE

REFERENCE_FETCH_FAILED
REFERENCE_TIMEOUT
REFERENCE_REDIRECT_LIMIT_EXCEEDED
REFERENCE_URL_FORBIDDEN

OUTPUT_DIR_CREATE_FAILED
FILE_WRITE_FAILED
FILE_ALREADY_EXISTS
INVALID_OUTPUT_PATH

MODEL_RESOLUTION_FAILED
BACKEND_UNAVAILABLE
RATE_LIMIT_EXCEEDED

INTERNAL_ERROR
```

### 16.2 エラー方針

- panic でも可能な限り JSON で返す
- partial success は許可しない
- 中途半端な保存ファイルは削除する

### 16.3 終了コード

- `0`: success
- `1`: error

v1 では retryable/non-retryable を exit code で分けない。

---

## 17. レートリミット・エラーと fallback の関係

### 17.1 `pro -> flash` fallback 条件

- `RATE_LIMIT_EXCEEDED`
- `PERMISSION_DENIED`
- `BACKEND_UNAVAILABLE`

### 17.2 やらないこと

- 全エラーで fallback しない
- silent fallback しない
- 他 backend への自動切替はしない

---

## 18. `config init` 仕様

### 18.1 目的

- config/credentials の骨格を作る
- profile を初期化する
- API key を保存する
- 可能なら remote model discovery を実行する

### 18.2 フロー

1. profile 名決定
2. API key 入力または引数受取
3. 軽い疎通確認
4. `config.toml` 作成
5. `credentials.json` 作成
6. `models.list` によるモデル解決を試行
7. 成功したら `[models]` を保存
8. 失敗したら built-in defaults を保存し warning を出す

### 18.3 non-interactive オプション

v1 README では以下を案内してよい。

```bash
imgraft config init --profile default --api-key YOUR_API_KEY
```

実装では `--init-profile`, `--init-api-key` などに分けてもよいが、README と CLI の整合を必ずとること。

---

## 19. 実行フロー

### 19.1 メイン実行パス

1. CLI parse
2. config load
3. profile resolve
4. auth resolve
5. model resolve
6. reference load/validate
7. prompt build
8. API generate
9. optional fallback
10. transparent pipeline
11. save
12. inspect/hash
13. JSON emit

### 19.2 参照画像あり

- 参照画像を content parts に変換して API 呼び出し
- 順番は保持

---

## 20. Go インターフェース設計

### 20.1 Config 型

```go
type Config struct {
    CurrentProfile   string            `toml:"current_profile"`
    LastUsedProfile  string            `toml:"last_used_profile"`
    LastUsedBackend  string            `toml:"last_used_backend"`
    DefaultModel     string            `toml:"default_model"`
    DefaultOutputDir string            `toml:"default_output_dir"`
    Models           map[string]string `toml:"models"`
}
```

### 20.2 Credentials 型

```go
type Credentials struct {
    Profiles map[string]ProfileCredentials `json:"profiles"`
}

type ProfileCredentials struct {
    GoogleAIStudio *GoogleAIStudioCredentials `json:"google_ai_studio,omitempty"`
}

type GoogleAIStudioCredentials struct {
    APIKey string `json:"api_key"`
}
```

### 20.3 RateLimit 型

```go
type RateLimit struct {
    Provider          *string `json:"provider"`
    LimitType         *string `json:"limit_type"`
    RequestsLimit     *int    `json:"requests_limit"`
    RequestsRemaining *int    `json:"requests_remaining"`
    RequestsUsed      *int    `json:"requests_used"`
    ResetAt           *string `json:"reset_at"`
    RetryAfterSeconds *int    `json:"retry_after_seconds"`
}
```

### 20.4 JSON 出力型

```go
type Output struct {
    Success   bool         `json:"success"`
    Model     *string      `json:"model"`
    Backend   *string      `json:"backend"`
    Images    []ImageItem  `json:"images"`
    RateLimit RateLimit    `json:"rate_limit"`
    Warnings  []string     `json:"warnings"`
    Error     OutputError  `json:"error"`
}

type ImageItem struct {
    Index              int    `json:"index"`
    Path               string `json:"path"`
    Filename           string `json:"filename"`
    Width              int    `json:"width"`
    Height             int    `json:"height"`
    MimeType           string `json:"mime_type"`
    SHA256             string `json:"sha256"`
    TransparentApplied bool   `json:"transparent_applied"`
}

type OutputError struct {
    Code    *string `json:"code"`
    Message *string `json:"message"`
}
```

### 20.5 Studio client interface

```go
type StudioClient interface {
    Generate(ctx context.Context, req GenerateRequest) (GenerateResponse, http.Header, error)
    ListModels(ctx context.Context) ([]RemoteModel, error)
    ValidateAPIKey(ctx context.Context) error
}
```

### 20.6 Reference 型

```go
type ReferenceImage struct {
    SourceType      string
    OriginalInput   string
    LocalCachedPath string
    Filename        string
    MimeType        string
    Width           int
    Height          int
    SizeBytes       int64
}
```

---

## 21. 背景除去アルゴリズム詳細

### 21.1 背景色推定

- 四隅をサンプルする
- 四隅の平均を背景色とする
- transparent ON では system prompt で背景を緑固定しているため、推定は主にノイズ吸収のために使う

### 21.2 色距離

```text
sqrt((r1-r2)^2 + (g1-g2)^2 + (b1-b2)^2)
```

### 21.3 alpha 生成

- `distance < threshold` は alpha 0
- しきい値付近は線形補間で 0..255 を割り当てる
- しきい値の上は alpha 255

### 21.4 エッジ補正

- alpha を 3x3 程度で軽く平滑化してよい
- heavy blur は不要

### 21.5 trim

- alpha > 0 の最小矩形を求める
- 画像全体が透明になった場合は `INTERNAL_ERROR` 相当で失敗とする

---

## 22. レートリミット解析詳細

### 22.1 ポリシー

- ヘッダが無ければ全部 `null`
- provider は backend が確定していれば `google_ai_studio` を入れてよい
- `requests_used` は `limit - remaining` のような推測計算をしない

### 22.2 実装観点

- SDK がヘッダを露出しない場合は HTTP client をラップしてレスポンスヘッダを取得する
- `Retry-After` は整数秒のみ対応でよい
- HTTP-date 形式まで解釈するなら unit test を必ず付ける

---

## 23. テスト戦略

### 23.1 unit test

必須:

- model resolver
- fallback 条件判定
- config loader/saver
- credentials loader/saver
- reference local validator
- reference remote validator
- URL forbidden host 判定
- filename generator
- trim
- background removal
- rate limit parser
- JSON encoder

### 23.2 integration test

- `config init` のファイル生成
- `auth login` -> `auth whoami`
- reference の混在入力
- `--output` と `--dir` 優先順位
- transparent デフォルト ON の実行フロー

### 23.3 e2e test

- `RUN_E2E=1` のときのみ実 API を叩く
- API key は環境変数から読む

---

## 24. README 方針

### 24.1 README.md

- 日本語 README へのリンクを入れてよい
- コンセプトを明記する
- `imgraft` は asset generator であると最初に書く
- transparent がデフォルトであることを先頭で説明する

### 24.2 README.ja.md

- 日本語版を提供する
- 使い方、設計思想、制約、スキル向け注意点を書く

---

## 25. SKILL.md 方針

`skills/imgraft/SKILL.md` では以下を明記する。

- 短い prompt を使う
- 背景を指示しない
- 素材用途で使う
- JSON を必ずパースする
- `pro` が落ちても `flash` fallback がある
- `--no-transparent` は例外的用途でのみ使う

---

## 26. GoReleaser 方針

### 26.1 目的

- macOS / Linux 向け単一バイナリ配布
- GitHub Releases
- Homebrew tap 連携

### 26.2 方針

- `cmd/imgraft` をビルド対象にする
- archive 名は OS/Arch を含める
- changelog は簡略でよい
- Homebrew formula を `youyo/tap` 前提で作る

---

## 27. 実装上の禁止事項

- stdout に JSON 以外を出す
- rate limit を推測値で埋める
- transparent デフォルトをオプトインに戻す
- 背景色をユーザー可変にする
- 参照画像を黙って無視する
- 失敗時にスキーマを変える
- 外部コマンドを追加する

---

## 28. v1 実装優先順位

1. config/auth
2. JSON 契約
3. main generate path
4. model resolution / refresh
5. reference local
6. reference remote
7. transparent pipeline
8. README / SKILL / GoReleaser

---

## 29. コーディングエージェントへの実装指示

- まず `internal/config`, `internal/auth`, `internal/model`, `internal/output` を作る
- 次に `internal/backend/studio` を作る
- その後 `internal/reference` と `internal/imageproc` を作る
- panic を避けるよりも、panic 時にも JSON を返すラッパーを整える
- README と SKILL.md はコードと同じタイミングで更新する
- 実装中に仕様追加が必要になっても、まず本書を更新してからコードに入る

---

## 30. 完成定義

以下を満たしたら v1 完成とする。

- `imgraft "blue robot mascot"` で透過 PNG が生成できる
- `imgraft auth login` が機能する
- `imgraft config init` が config を作れる
- `imgraft config refresh-models` が動く
- `--ref` にローカルファイルと URL を指定できる
- stdout JSON が固定スキーマを守る
- `rate_limit` がオブジェクト固定で出る
- `README.md`, `README.ja.md`, `skills/imgraft/SKILL.md`, `.goreleaser.yaml` が揃う

