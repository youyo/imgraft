package model_test

import (
	"testing"

	"github.com/youyo/imgraft/internal/errs"
	"github.com/youyo/imgraft/internal/model"
)

// TestFallbackModel_ProFallsToFlash は pro モデルが flash にフォールバックできることを確認する。
func TestFallbackModel_ProFallsToFlash(t *testing.T) {
	fallback, ok := model.FallbackModel(model.BuiltinProModel)
	if !ok {
		t.Fatal("expected fallback to be possible for pro model")
	}
	if fallback != model.BuiltinFlashModel {
		t.Errorf("expected fallback model %q, got %q", model.BuiltinFlashModel, fallback)
	}
}

// TestFallbackModel_FlashNoFallback は flash モデルはフォールバック不可であることを確認する。
func TestFallbackModel_FlashNoFallback(t *testing.T) {
	_, ok := model.FallbackModel(model.BuiltinFlashModel)
	if ok {
		t.Error("expected no fallback for flash model")
	}
}

// TestFallbackModel_CustomProModel はカスタム pro モデル名のフォールバックは不可であることを確認する。
// FallbackModel は exact match で判定する。
func TestFallbackModel_CustomProModel(t *testing.T) {
	_, ok := model.FallbackModel("custom-pro-model")
	if ok {
		t.Error("expected no fallback for unknown/custom model names")
	}
}

// TestFallbackModel_EmptyString は空文字列のフォールバックが不可であることを確認する。
func TestFallbackModel_EmptyString(t *testing.T) {
	_, ok := model.FallbackModel("")
	if ok {
		t.Error("expected no fallback for empty model name")
	}
}

// TestIsFallbackError_RateLimitExceeded は RATE_LIMIT_EXCEEDED がフォールバック対象であることを確認する。
func TestIsFallbackError_RateLimitExceeded(t *testing.T) {
	err := errs.New(errs.ErrRateLimitExceeded, "rate limit hit")
	if !model.IsFallbackError(err) {
		t.Error("expected RATE_LIMIT_EXCEEDED to be a fallback error")
	}
}

// TestIsFallbackError_PermissionDenied は PERMISSION_DENIED (AUTH_INVALID) がフォールバック対象であることを確認する。
func TestIsFallbackError_PermissionDenied(t *testing.T) {
	err := errs.New(errs.ErrAuthInvalid, "permission denied")
	if !model.IsFallbackError(err) {
		t.Error("expected AUTH_INVALID (PERMISSION_DENIED) to be a fallback error")
	}
}

// TestIsFallbackError_BackendUnavailable は BACKEND_UNAVAILABLE がフォールバック対象であることを確認する。
func TestIsFallbackError_BackendUnavailable(t *testing.T) {
	err := errs.New(errs.ErrBackendUnavailable, "backend down")
	if !model.IsFallbackError(err) {
		t.Error("expected BACKEND_UNAVAILABLE to be a fallback error")
	}
}

// TestIsFallbackError_OtherError はフォールバック対象外のエラーが false を返すことを確認する。
func TestIsFallbackError_OtherError(t *testing.T) {
	otherErrors := []error{
		errs.New(errs.ErrInvalidArgument, "bad arg"),
		errs.New(errs.ErrAuthRequired, "auth required"),
		errs.New(errs.ErrInternal, "internal error"),
		errs.New(errs.ErrFileNotFound, "file not found"),
	}
	for _, err := range otherErrors {
		if model.IsFallbackError(err) {
			t.Errorf("expected error %v not to be a fallback error", err)
		}
	}
}

// TestIsFallbackError_Nil は nil エラーが false を返すことを確認する。
func TestIsFallbackError_Nil(t *testing.T) {
	if model.IsFallbackError(nil) {
		t.Error("expected nil error not to be a fallback error")
	}
}

// TestFallbackWarningMessage はフォールバックの警告メッセージを確認する。
func TestFallbackWarningMessage(t *testing.T) {
	msg := model.FallbackWarning(model.BuiltinProModel, model.BuiltinFlashModel)
	if msg == "" {
		t.Error("expected non-empty fallback warning message")
	}
}
