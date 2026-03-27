# Roadmap: imgraft

## Meta
| 項目 | 値 |
|------|---|
| ゴール | LLM coding-agent 向け透明画像アセット生成 CLI の v1 完成 |
| 成功基準 | SPEC.md セクション30 の完成定義を全て満たす |
| 制約 | pure Go, 外部コマンド依存なし, stdout は固定スキーマ JSON のみ |
| 対象リポジトリ | /Users/youyo/src/github.com/youyo/imgraft |
| 作成日 | 2026-03-27 |
| 最終更新 | 2026-03-27 16:20 |
| ステータス | 未着手 |

## 技術選定
| 項目 | 選定 |
|------|------|
| CLI フレームワーク | kong (github.com/alecthomas/kong) |
| TOML パーサー | BurntSushi/toml (github.com/BurntSushi/toml) |
| WebP デコード | golang.org/x/image/webp |
| Google AI SDK | google.golang.org/api/generativelanguage/... or REST client |

## Current Focus
- **マイルストーン**: M01
- **直近の完了**: ロードマップ作成
- **次のアクション**: M01 の詳細計画に基づき実装開始

## Progress

### M01: プロジェクト基盤・ランタイム
- [ ] go.mod 初期化・依存追加
- [ ] internal/runtime (paths, env, clock)
- [ ] unit test
- 📄 詳細: plans/imgraft-m01-project-foundation.md

### M02: 設定管理
- [ ] internal/config/types.go (Config 型)
- [ ] internal/config/loader.go (TOML 読込・defaults 補完)
- [ ] internal/config/saver.go (TOML 書込)
- [ ] unit test
- 📄 詳細: plans/imgraft-m02-config.md

### M03: 認証情報管理
- [ ] internal/auth/types.go (Credentials, ProfileCredentials 型)
- [ ] internal/auth/loader.go (JSON 読込)
- [ ] internal/auth/saver.go (JSON 書込, 0600 permission)
- [ ] internal/auth/validate.go (API key マスキング)
- [ ] unit test
- 📄 詳細: plans/imgraft-m03-credentials.md

### M04: エラーコード定義
- [ ] internal/errs/codes.go (32種 enum)
- [ ] internal/errs/map.go (error → code 変換)
- [ ] unit test
- 📄 詳細: plans/imgraft-m04-error-codes.md

### M05: JSON 出力契約
- [ ] internal/output/types.go (Output, ImageItem, OutputError, RateLimit 型)
- [ ] internal/output/json.go (JSON encoder, null 保証, --pretty 対応)
- [ ] unit test (スキーマ固定テスト)
- 📄 詳細: plans/imgraft-m05-json-contract.md

### M06: モデル解決
- [ ] internal/model/defaults.go (built-in flash/pro)
- [ ] internal/model/resolver.go (alias → full model name)
- [ ] unit test
- 📄 詳細: plans/imgraft-m06-model-resolver.md

### M07: プロンプト構築
- [ ] internal/prompt/system.go (transparent ON/OFF system prompt)
- [ ] internal/prompt/builder.go (user prompt + system prompt 合成)
- [ ] unit test
- 📄 詳細: plans/imgraft-m07-prompt-builder.md

### M08: Studio API クライアント
- [ ] internal/backend/studio/client.go (StudioClient interface)
- [ ] internal/backend/studio/generate.go (Generate メソッド)
- [ ] internal/backend/studio/types.go (Request/Response 型)
- [ ] internal/backend/studio/errors.go (API エラーマッピング)
- [ ] unit test (mock)
- 📄 詳細: plans/imgraft-m08-studio-client.md

### M09: 画像デコード・エンコード
- [ ] internal/imageproc/decode.go (PNG/JPEG/WebP デコード)
- [ ] internal/imageproc/encode.go (PNG エンコード with alpha)
- [ ] internal/imageproc/hash.go (SHA256)
- [ ] internal/imageproc/inspect.go (width, height, mime_type)
- [ ] unit test
- 📄 詳細: plans/imgraft-m09-image-codec.md

### M10: ファイル出力
- [ ] internal/output/naming.go (imgraft-YYYYMMDD-HHMMSS-XXX.png, 衝突回避)
- [ ] internal/output/save.go (write → close → inspect → hash)
- [ ] unit test
- 📄 詳細: plans/imgraft-m10-file-output.md

### M11: メイン生成パス統合
- [ ] cmd/imgraft/main.go (エントリポイント)
- [ ] internal/cli/root.go (kong CLI 定義)
- [ ] internal/app/run.go (13ステップ パイプライン)
- [ ] integration test
- [ ] E2E: `imgraft "blue robot mascot"` で PNG 生成確認
- 📄 詳細: plans/imgraft-m11-main-generate-path.md

### M12: 参照画像 - ローカル
- [ ] internal/reference/types.go (ReferenceImage 型)
- [ ] internal/reference/local.go (ファイル読込)
- [ ] internal/reference/validate.go (decode 検証, サイズ制限)
- [ ] app/run に参照画像 inject
- [ ] unit test
- 📄 詳細: plans/imgraft-m12-reference-local.md

### M13: 参照画像 - リモート
- [ ] internal/reference/remote.go (HTTP GET, リダイレクト≤3, タイムアウト)
- [ ] internal/reference/forbidden.go (private IP / localhost 拒否)
- [ ] mixed references (ローカル + URL 混在, 順序保持)
- [ ] unit test (mock HTTP server)
- 📄 詳細: plans/imgraft-m13-reference-remote.md

### M14: 透明パイプライン - 背景除去
- [ ] internal/imageproc/background.go (corner_sampling_color_distance)
- [ ] 四隅サンプル → 背景色推定
- [ ] RGB 距離計算 (threshold=40)
- [ ] 段階的 alpha 生成
- [ ] 3x3 エッジ平滑化
- [ ] unit test (アルゴリズム正確性)
- 📄 詳細: plans/imgraft-m14-background-removal.md

### M15: 透明パイプライン - Trim・統合
- [ ] internal/imageproc/trim.go (alpha > 0 bounding box)
- [ ] 全透明検出 → INTERNAL_ERROR
- [ ] transparent pipeline をメイン生成パスに統合
- [ ] transparent_applied フラグ出力
- [ ] unit test + integration test
- 📄 詳細: plans/imgraft-m15-trim-integration.md

### M16: レート制限解析
- [ ] internal/ratelimit/types.go (RateLimit 型)
- [ ] internal/ratelimit/headers.go (HTTP ヘッダー解析)
- [ ] Retry-After 抽出
- [ ] null 初期化保証
- [ ] unit test
- 📄 詳細: plans/imgraft-m16-ratelimit.md

### M17: モデル Fallback
- [ ] pro → flash fallback ロジック
- [ ] 対象エラー判定 (RATE_LIMIT_EXCEEDED, PERMISSION_DENIED, BACKEND_UNAVAILABLE)
- [ ] warnings 記録
- [ ] unit test
- 📄 詳細: plans/imgraft-m17-fallback.md

### M18: 認証コマンド
- [ ] internal/cli/auth.go (auth サブコマンド)
- [ ] internal/auth/login.go (対話フロー + 疎通確認)
- [ ] internal/auth/logout.go (profile 内 backend 削除)
- [ ] internal/auth/whoami.go (表示, API key 末尾4文字のみ)
- [ ] integration test
- 📄 詳細: plans/imgraft-m18-auth-commands.md

### M19: config init・refresh-models
- [ ] internal/cli/config.go (config サブコマンド)
- [ ] config init フロー (profile → API key → 疎通 → models.list → save)
- [ ] --profile, --api-key non-interactive オプション
- [ ] config refresh-models (モデル一覧取得 → [models] 上書き)
- [ ] internal/backend/studio/models.go (ListModels)
- [ ] internal/model/refresh.go (flash/pro ヒューリスティック解決)
- [ ] config use <profile>
- [ ] unit test + integration test
- 📄 詳細: plans/imgraft-m19-config-commands.md

### M20: CLI 補助コマンド
- [ ] imgraft version (ldflags 埋込)
- [ ] imgraft completion zsh
- [ ] unit test
- 📄 詳細: plans/imgraft-m20-cli-helpers.md

### M21: ドキュメント
- [ ] README.md (英語)
- [ ] README.ja.md (日本語)
- [ ] skills/imgraft/SKILL.md (LLM agent 向けガイド)
- 📄 詳細: plans/imgraft-m21-documentation.md

### M22: リリース設定
- [ ] .goreleaser.yaml (macOS/Linux バイナリ, Homebrew tap)
- [ ] GitHub Releases 設定
- 📄 詳細: plans/imgraft-m22-release.md

## 依存関係
```
M01 ──→ M02 ──→ M03
  │       │
  │       └──→ M06 ──→ M07
  │
  └──→ M04 ──→ M05
         │
         └──→ M08 ──→ M09 ──→ M10
                │
                └──→ M11 (M01-M10 全て必要)
                       │
                       ├──→ M12 ──→ M13
                       │
                       ├──→ M14 ──→ M15
                       │
                       ├──→ M16 ──→ M17
                       │
                       └──→ M18 ──→ M19
                              │
                              └──→ M20

M15 + M17 + M19 + M20 ──→ M21 ──→ M22
```

## Blockers
なし

## Architecture Decisions
| # | 決定 | 理由 | 日付 |
|---|------|------|------|
| 1 | CLI フレームワークに kong を採用 | struct ベースで型安全、cobra より簡潔 | 2026-03-27 |
| 2 | TOML パーサーに BurntSushi/toml を採用 | Go の TOML 定番、TOML v1.0 準拠 | 2026-03-27 |
| 3 | マイルストーンを22分割 | 各MSが独立テスト・デモ可能な粒度 | 2026-03-27 |

## Changelog
| 日時 | 種別 | 内容 |
|------|------|------|
| 2026-03-27 16:20 | 作成 | ロードマップ初版作成。SPEC.md セクション28-29に準拠し22マイルストーンに細分化 |
