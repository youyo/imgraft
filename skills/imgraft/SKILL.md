# imgraft SKILL

## Purpose

`imgraft` is a CLI for generating **reusable transparent image assets** from LLM coding agents.

This skill defines the rules for agents to use `imgraft` reliably.

---

## What imgraft is for

`imgraft` is not a general-purpose image generator.

Preferred use cases:

- UI components
- Icons
- Logo drafts
- Mascots
- Sticker-style assets
- Composite parts for other images or UIs

The goal is to produce **transparent PNG assets** stably, not complete scenes with backgrounds.

---

## Critical Rules

### 1. Keep prompts short

A good prompt is a short combination of **subject + style + constraint**.

Good examples:

```text
blue robot mascot, simple, modern
tech logo, geometric, high contrast
cute cat sticker, pastel colors
settings icon, line style
```

Avoid:

```text
a beautiful robot standing in a futuristic city at night with neon buildings and dramatic reflections
```

Long prompts increase noise for asset generation use cases.

---

### 2. Do not describe backgrounds

In transparent mode, `imgraft` controls the background internally.

Avoid:

```text
robot in a forest
city background
space scene
night skyline
```

Background descriptions reduce the success rate of transparent asset generation.

---

### 3. Use transparent mode by default

`imgraft` is transparent by default. Do not add `--no-transparent` unless explicitly needed.

Normal usage:

```bash
imgraft "blue robot mascot, simple, modern"
```

Only when a background is explicitly required:

```bash
imgraft "scene illustration" --no-transparent
```

---

### 4. Always parse stdout as JSON

`imgraft`'s stdout is JSON. Treat it as the source of truth, not stderr.

Fields to check:

- `success`
- `model`
- `images[0].path`
- `warnings`
- `error.code`
- `error.message`

---

## Recommended Command Patterns

### Basic

```bash
imgraft "blue robot mascot, simple, modern"
```

### With reference image (local file)

```bash
imgraft "convert to icon style" --ref ./input.png
```

### With reference image (URL)

```bash
imgraft "convert to sticker style" --ref https://example.com/input.png
```

### Multiple reference images

```bash
imgraft "redesign to match style" --ref ./subject.png --ref ./style.png
```

### Higher quality

```bash
imgraft "dragon logo, bold, geometric" --model pro
```

### Specify output directory

```bash
imgraft "cute cat icon, pastel" --dir ./out
```

### Specify output filename

```bash
imgraft "app icon, line style" --output ./asset.png
```

### Pretty-print JSON output

```bash
imgraft "settings icon" --pretty
```

---

## Model Selection Rules

`flash` is suitable for almost all cases.

- `flash`: Default. Speed and stability prioritized.
- `pro`: Higher quality. May fall back to `flash` on failure.

Recommendation:
- Start with `flash`
- Use `pro` only when quality matters

The `model` field in the output contains the **actual** model used. If fallback occurred, it will contain the `flash` model name even when `pro` was requested.

---

## Reference Image Guidelines

Recommended reference images:

- Clear subject
- No UI chrome or unnecessary framing
- Reasonable image quality
- Not overly complex background

Avoid:

- GIF, SVG, PDF formats
- Images larger than 20 MB
- localhost or private IP URLs

---

## Handling Failures

### When `success == false`

Check:

- `error.code`
- `error.message`

Common error codes and actions:

| Code | Action |
|------|--------|
| `RATE_LIMIT_EXCEEDED` | Wait and retry |
| `FILE_NOT_FOUND` | Fix reference image path |
| `REFERENCE_FETCH_FAILED` | Check the URL |
| `REFERENCE_URL_FORBIDDEN` | Do not use private/localhost URLs |
| `UNSUPPORTED_IMAGE_FORMAT` | Convert to PNG/JPEG/WebP and retry |
| `AUTH_REQUIRED` | Run `imgraft auth login` |
| `AUTH_INVALID` | Check API key with `imgraft auth whoami` |

---

## Handling Warnings

Common warnings:

- `model fallback: pro -> flash` — pro was requested but flash was used
- `background removal quality may be low` — transparent quality may be degraded
- `using http reference URL` — consider using https

`warnings` does not mean failure. Use it as a signal for deciding whether to retry.

---

## Agent Decision Rules

### When you need a transparent asset

Use `imgraft` as-is (default behavior).

### When you need a background scene

`imgraft` may not be the best tool. If you must use it, add `--no-transparent`.

### For logos, icons, UI components

`imgraft` is ideal. Use it actively.

### For photo-quality or precise background removal

`imgraft` is not designed for this. Use it as a pure asset generator only.

---

## Post-execution Flow

After running `imgraft`, agents should:

1. Execute `imgraft`
2. Parse stdout as JSON
3. Check `success`
4. Retrieve `images[0].path`
5. Use that path in subsequent operations

---

## JSON Output Schema Reference

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

## Recommended System Prompt Template

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

## Things to Avoid

- Submitting long, scene-describing prompts
- Describing background in the prompt (transparent mode)
- Treating stderr as the source of truth instead of JSON stdout
- Ignoring reference image errors
- Confusing `warnings` with `success: false`
- Using `imgraft` for complex scene generation it is not designed for
