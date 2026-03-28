package model

import (
	"errors"
	"fmt"

	"github.com/youyo/imgraft/internal/errs"
)

// fallbackMap は pro モデルから flash モデルへのフォールバックマッピング。
// SPEC.md §9.4: pro 指定時のみ fallback を許可する。
var fallbackMap = map[string]string{
	BuiltinProModel: BuiltinFlashModel,
}

// fallbackErrorCodes は FallbackModel のトリガーとなるエラーコード一覧。
// SPEC.md §9.4: RATE_LIMIT_EXCEEDED, PERMISSION_DENIED, BACKEND_UNAVAILABLE
var fallbackErrorCodes = map[errs.ErrorCode]bool{
	errs.ErrRateLimitExceeded: true,
	errs.ErrAuthInvalid:       true, // PERMISSION_DENIED に対応
	errs.ErrBackendUnavailable: true,
}

// FallbackModel は現在のモデル名に対するフォールバック先モデル名と可否を返す。
// フォールバック可能な場合は (fallbackModel, true) を返す。
// フォールバック不可能な場合は ("", false) を返す。
//
// SPEC.md §9.4: pro → flash へのフォールバックのみ許可する。
func FallbackModel(current string) (string, bool) {
	if fallback, ok := fallbackMap[current]; ok {
		return fallback, true
	}
	return "", false
}

// IsFallbackError はエラーがフォールバックのトリガーとなるかを判定する。
// SPEC.md §9.4: RATE_LIMIT_EXCEEDED, PERMISSION_DENIED (AUTH_INVALID), BACKEND_UNAVAILABLE の場合に true を返す。
func IsFallbackError(err error) bool {
	if err == nil {
		return false
	}
	var coded *errs.CodedError
	if errors.As(err, &coded) {
		return fallbackErrorCodes[coded.Code]
	}
	return false
}

// FallbackWarning はフォールバック発生時の警告メッセージを返す。
// warnings フィールドに記録するために使用する。
func FallbackWarning(fromModel, toModel string) string {
	return fmt.Sprintf("model fallback: %s → %s (retrying with fallback model)", fromModel, toModel)
}
