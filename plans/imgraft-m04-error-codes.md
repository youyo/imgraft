# M04: エラーコード定義

## Meta
| 項目 | 値 |
|------|---|
| マイルストーン | M04 |
| 依存 | M01（go.mod, internal/runtime） |
| 対象ファイル | internal/errs/codes.go, internal/errs/map.go |
| 作成日 | 2026-03-27 |
| ステータス | 実装中 |

## 目的

SPEC.md セクション16に定義されたエラーコードを Go の型安全な定数として実装し、
`error` 値からエラーコードへの変換マッピングを提供する。

stdout JSON の `error.code` フィールドに使われる文字列を一元管理する。

## エラーコード一覧（SPEC.md セクション16.1より）

```
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

合計: **20種**

> 注: ロードマップでは「32種」と記載されていたが、SPEC.md セクション16.1 が正本であり20種が正しい。

## 設計方針

### codes.go

- `ErrorCode` を `string` の型エイリアスとして定義する
- 各コードを `const` ブロックで定義する
- `String()` メソッドを持たせる（`ErrorCode` は既に `string` なので不要）
- `IsValid()` メソッドで有効なコードかチェックできるようにする

```go
type ErrorCode string

const (
    ErrInvalidArgument              ErrorCode = "INVALID_ARGUMENT"
    ErrAuthRequired                 ErrorCode = "AUTH_REQUIRED"
    ErrAuthInvalid                  ErrorCode = "AUTH_INVALID"
    ErrFileNotFound                 ErrorCode = "FILE_NOT_FOUND"
    ErrFileReadFailed               ErrorCode = "FILE_READ_FAILED"
    ErrUnsupportedImageFormat       ErrorCode = "UNSUPPORTED_IMAGE_FORMAT"
    ErrImageTooLarge                ErrorCode = "IMAGE_TOO_LARGE"
    ErrInvalidImage                 ErrorCode = "INVALID_IMAGE"
    ErrReferenceFetchFailed         ErrorCode = "REFERENCE_FETCH_FAILED"
    ErrReferenceTimeout             ErrorCode = "REFERENCE_TIMEOUT"
    ErrReferenceRedirectLimitExceeded ErrorCode = "REFERENCE_REDIRECT_LIMIT_EXCEEDED"
    ErrReferenceURLForbidden        ErrorCode = "REFERENCE_URL_FORBIDDEN"
    ErrOutputDirCreateFailed        ErrorCode = "OUTPUT_DIR_CREATE_FAILED"
    ErrFileWriteFailed              ErrorCode = "FILE_WRITE_FAILED"
    ErrFileAlreadyExists            ErrorCode = "FILE_ALREADY_EXISTS"
    ErrInvalidOutputPath            ErrorCode = "INVALID_OUTPUT_PATH"
    ErrModelResolutionFailed        ErrorCode = "MODEL_RESOLUTION_FAILED"
    ErrBackendUnavailable           ErrorCode = "BACKEND_UNAVAILABLE"
    ErrRateLimitExceeded            ErrorCode = "RATE_LIMIT_EXCEEDED"
    ErrInternal                     ErrorCode = "INTERNAL_ERROR"
)
```

### map.go

`CodedError` 型: `ErrorCode` + `error` を組み合わせた構造体

```go
type CodedError struct {
    Code ErrorCode
    Err  error
}

func (e *CodedError) Error() string
func (e *CodedError) Unwrap() error

func New(code ErrorCode, msg string) *CodedError
func Wrap(code ErrorCode, err error) *CodedError
func CodeOf(err error) ErrorCode  // error から ErrorCode を取得（非 CodedError は ErrInternal）
```

## TDD 設計

### Red フェーズ（テスト先行）

**codes_test.go:**
1. `ErrorCode` の各定数が期待する文字列値を持つことをテスト
2. `AllCodes()` が全20種を返すことをテスト
3. 有効なコードと無効なコードの区別テスト

**map_test.go:**
1. `New()` で `CodedError` が生成され、`Error()` が正しい文字列を返すテスト
2. `Wrap()` で既存 `error` をラップして `Unwrap()` で取得できるテスト
3. `CodeOf()` で `CodedError` から正しいコードが取得できるテスト
4. `CodeOf()` で非 `CodedError` に対して `ErrInternal` が返るテスト
5. `errors.As()` で `CodedError` を取得できるテスト

### Green フェーズ

テストを通す最小限の実装を行う。

### Refactor フェーズ

- `AllCodes()` のリテラル列挙をチェック
- エラーメッセージのフォーマットを統一する

## 実装ステップ

1. `internal/errs/` ディレクトリ作成
2. `codes_test.go` 作成（Red）
3. `codes.go` 実装（Green）
4. `map_test.go` 作成（Red）
5. `map.go` 実装（Green）
6. リファクタリング（Refactor）
7. `go test ./internal/errs/` で green 確認
8. `go vet ./...` でリント確認

## 検証

```bash
go test ./internal/errs/ -v
go vet ./...
```

## 完了条件

- [ ] `internal/errs/codes.go` が実装されている
- [ ] `internal/errs/map.go` が実装されている
- [ ] `go test ./internal/errs/ -v` で全テスト green
- [ ] `go vet ./...` でエラーなし
- [ ] 全20種のエラーコードが定義されている
