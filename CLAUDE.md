# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## プロジェクト概要

imgraft は **LLM coding-agent 向けの透明画像アセット生成 CLI**。Claude Code / Codex / Gemini CLI などのエージェントから呼び出され、背景透過 PNG を自動生成する。汎用画像生成ツールではない。

> **imgraft = transparent asset generator for automation pipelines**

## 技術スタック

- **言語:** Go (pure Go, 外部コマンド依存なし)
- **ツールバージョン管理:** mise (`mise.toml`)
- **リリース:** GoReleaser
- **バックエンド:** Google AI Studio API のみ (v1)

## ビルド・テスト・リント

```bash
# ビルド
go build ./cmd/imgraft/

# テスト
go test ./...

# 単一パッケージのテスト
go test ./internal/imageproc/

# E2Eテスト（実APIを呼ぶため通常はスキップ）
RUN_E2E=1 go test ./...

# リント
go vet ./...
```

## アーキテクチャ

### ディレクトリ構成

```
cmd/imgraft/main.go    # エントリポイント
internal/
  app/                 # run ロジック（メインパイプライン）
  cli/                 # CLI ハンドラ（cobra等）
  config/              # TOML設定ローダー
  auth/                # 認証（API キー管理）
  backend/studio/      # Google AI Studio APIクライアント
  model/               # モデルalias解決（flash/pro→実モデル名）
  reference/           # 参照画像ローダー（ローカル/URL）
  prompt/              # プロンプト構築（透明モード用システムプロンプト含む）
  imageproc/           # 画像処理（背景除去・trim）
  output/              # ファイル出力
  ratelimit/           # レート制限ヘッダー解析
  errs/                # エラーコード定義
  runtime/             # 環境・パス・clock
```

### メイン処理パイプライン

CLI parse → Config load → Auth解決 → Model解決 → 参照画像読込・検証 → Prompt構築 → API呼出 → (Fallback) → 透明パイプライン(背景除去・trim) → ファイル保存 → メタデータ・SHA256 → JSON出力

### 設計上の重要ルール

- **stdout は常に固定スキーマの JSON**。スキーマのフィールドは省略せず、欠損は `null` で表現
- **透明モードがデフォルト ON**。システムプロンプトで純緑背景(#00FF00)を強制し、pure Go で背景除去・trim
- **モデルfallback:** `pro` → `flash` に自動フォールバック（RATE_LIMIT_EXCEEDED, PERMISSION_DENIED, BACKEND_UNAVAILABLE時）
- **参照画像:** 最大8枚、各20MB以下、4096×4096以下、PNG/JPEG/WebPのみ、プライベートIP/localhost URL拒否
- **エラーコード:** 32種類定義済み（`internal/errs/` に集約）

## 設定ファイル

- **設定:** `~/.config/imgraft/config.toml`
- **認証情報:** `~/.config/imgraft/credentials.json`

## 仕様書

詳細仕様は `docs/specs/SPEC.md` に集約されている。実装時はこのファイルを参照すること。
