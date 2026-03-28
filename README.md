# imgraft

**Transparent image asset generator for LLM coding agents.**

imgraft is not a general-purpose image generation CLI. Its primary purpose is to be called from automation pipelines such as Claude Code, Codex, and Gemini CLI to stably generate **reusable transparent PNG assets**.

[日本語版 README はこちら](README.ja.md)

## Features

- Transparent asset generation is **on by default**
- Pure Go implementation — no external command dependencies
- stdout is always a fixed-schema JSON
- `flash` / `pro` model aliases supported
- Real model names are configurable via `config.toml`
- Reference images support both local files and URLs
- Stable output contract designed for LLM agents

## Use Cases

imgraft is optimized for these asset types:

- UI components
- Icons
- Logo drafts
- Sticker-style assets
- Mascots
- Composite parts to be used in other images

It is **not** suited for:
- Full scene / background art generation
- Photo-realistic outputs
- Precise text rendering
- High-accuracy background removal (remove.bg equivalent)

## Installation

### Homebrew

```bash
brew install youyo/tap/imgraft
```

### go install

```bash
go install github.com/youyo/imgraft/cmd/imgraft@latest
```

## Quick Start

### 1. Initialize configuration

Interactive:

```bash
imgraft config init
```

Non-interactive:

```bash
imgraft config init --profile default --api-key YOUR_API_KEY
```

### 2. Verify authentication

```bash
imgraft auth whoami
```

### 3. Generate a transparent asset

```bash
imgraft "blue robot mascot, simple, modern"
```

### 4. Generate with a reference image

```bash
imgraft "convert to icon style" --ref ./input.png
```

### 5. Disable transparent mode

```bash
imgraft "scene illustration" --no-transparent
```

## CLI Reference

### Main command

```bash
imgraft "<prompt>"
```

Generate a transparent image asset. When no subcommand is specified, this is the default.

### Subcommands

```bash
# Authentication
imgraft auth login
imgraft auth logout
imgraft auth whoami

# Configuration
imgraft config init
imgraft config use <profile>
imgraft config refresh-models

# Utilities
imgraft version
imgraft completion zsh
```

### Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--model` | Model alias or full name (`flash`, `pro`, or full model name) | `flash` |
| `--ref` | Reference image path or URL (repeatable) | — |
| `--output` | Output file path | auto-named |
| `--dir` | Output directory | `.` |
| `--no-transparent` | Disable transparent mode | `false` |
| `--profile` | Profile name to use | current profile |
| `--config` | Config file path | `~/.config/imgraft/config.toml` |
| `--pretty` | Pretty-print JSON output | `false` |
| `--verbose` | Enable verbose logging to stderr | `false` |
| `--debug` | Enable debug logging to stderr | `false` |

## Models

imgraft supports two model aliases:

- `flash` — Standard use. Speed and stability prioritized.
- `pro` — Higher quality. Falls back to `flash` on rate limit or availability errors.

Real model names are managed in `~/.config/imgraft/config.toml`:

```toml
[models]
flash = "gemini-3.1-flash-image-preview"
pro   = "gemini-3-pro-image-preview"
```

`config init` and `config refresh-models` attempt to fetch the latest model list and update these values automatically. If the fetch fails, built-in defaults are used.

### Fallback behavior

When `--model pro` is specified and one of the following errors occurs, imgraft automatically falls back to `flash`:

- `RATE_LIMIT_EXCEEDED`
- `PERMISSION_DENIED`
- `BACKEND_UNAVAILABLE`

The fallback is always recorded in the `warnings` field of the JSON output.

## Configuration Files

### Config

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

### Credentials

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

### Config resolution order

1. CLI flags
2. Profile config values
3. Environment variables
4. Built-in defaults

## JSON Output Schema

stdout is always JSON, both on success and failure.

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

### Schema rules

- `success` is the single source of truth for success/failure.
- `model` is the **actual** model name used (may differ from requested if fallback occurred).
- `backend` is always `google_ai_studio` in v1.
- `images` is an empty array on failure.
- `rate_limit` object is always present; individual fields are `null` when unavailable.
- `error.code` and `error.message` are set on failure; `null` on success.
- stderr is for human-readable logs only — never mixed into stdout.

### Output file naming

Auto-named files follow this pattern:

```
imgraft-YYYYMMDD-HHMMSS-XXX.png
```

Example: `imgraft-20260324-153012-001.png`

## Transparent Pipeline

When transparent mode is active (default), imgraft runs the following pipeline:

1. Apply system prompt — forces solid pure green `#00FF00` background
2. Generate image via API
3. Decode PNG
4. Background removal (corner-sampling color-distance algorithm)
5. Trim transparent margins
6. Encode as PNG with alpha channel
7. Save to disk

### Important: background is controlled by imgraft

In transparent mode, any background description in the user's prompt is **overridden** by the internal system prompt. This is intentional — the green background makes color-distance-based removal more reliable.

Use `--no-transparent` only when you explicitly need a background.

## Reference Images

Reference images are passed via `--ref` (repeatable up to 8 times):

```bash
imgraft "convert to icon style" --ref ./icon.png --ref ./style.png
```

Supported sources:
- Local files
- `http://` and `https://` URLs

Supported formats: PNG, JPEG, WebP

Limits:
- Max 8 images
- Max 20 MB per image
- Max 4096 × 4096 resolution
- Private IP, loopback, and localhost URLs are rejected

## Error Codes

| Code | Description |
|------|-------------|
| `INVALID_ARGUMENT` | Invalid CLI argument |
| `AUTH_REQUIRED` | No authentication found |
| `AUTH_INVALID` | Invalid API key |
| `FILE_NOT_FOUND` | Reference file not found |
| `FILE_READ_FAILED` | Failed to read file |
| `UNSUPPORTED_IMAGE_FORMAT` | Unsupported image format |
| `IMAGE_TOO_LARGE` | Image exceeds size or resolution limit |
| `INVALID_IMAGE` | Image could not be decoded |
| `REFERENCE_FETCH_FAILED` | Failed to fetch reference URL |
| `REFERENCE_TIMEOUT` | Reference URL fetch timed out |
| `REFERENCE_REDIRECT_LIMIT_EXCEEDED` | Too many redirects |
| `REFERENCE_URL_FORBIDDEN` | Private/localhost URL rejected |
| `OUTPUT_DIR_CREATE_FAILED` | Could not create output directory |
| `FILE_WRITE_FAILED` | Failed to write output file |
| `FILE_ALREADY_EXISTS` | Output file already exists |
| `INVALID_OUTPUT_PATH` | Invalid output path specified |
| `MODEL_RESOLUTION_FAILED` | Could not resolve model name |
| `BACKEND_UNAVAILABLE` | Backend API unavailable |
| `RATE_LIMIT_EXCEEDED` | API rate limit exceeded |
| `INTERNAL_ERROR` | Unexpected internal error |

## Development

### Build

```bash
go build ./cmd/imgraft/
```

### Test

```bash
# Unit and integration tests
go test ./...

# Single package
go test ./internal/imageproc/

# E2E tests (requires real API key)
RUN_E2E=1 go test ./...
```

### Lint

```bash
go vet ./...
```

### Build with version info

```bash
go build \
  -ldflags "-X github.com/youyo/imgraft/internal/cli.Version=1.0.0 \
            -X github.com/youyo/imgraft/internal/cli.Commit=$(git rev-parse --short HEAD) \
            -X github.com/youyo/imgraft/internal/cli.Date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  ./cmd/imgraft/
```

## References

- Product specification: `docs/specs/SPEC.md`
- Agent usage guide: `skills/imgraft/SKILL.md`

## License

MIT
