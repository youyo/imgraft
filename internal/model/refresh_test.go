package model_test

import (
	"testing"

	"github.com/youyo/imgraft/internal/model"
)

// TestResolveAliasesFromModels_FlashAndPro は flash と pro を含むモデル名を正しく解決するテスト。
func TestResolveAliasesFromModels_FlashAndPro(t *testing.T) {
	models := []string{
		"models/gemini-2.0-flash-exp-image-generation",
		"models/gemini-2.5-pro-exp-03-25",
		"models/gemini-1.5-flash",
		"models/gemma-2-27b-it",
	}

	aliases := model.ResolveAliasesFromModels(models)

	// flash を含むモデルが flash alias に解決されるべき
	if aliases["flash"] == "" {
		t.Error("flash alias should be resolved")
	}

	// pro を含むモデルが pro alias に解決されるべき
	if aliases["pro"] == "" {
		t.Error("pro alias should be resolved")
	}
}

// TestResolveAliasesFromModels_NoFlashNoProModels は flash/pro を含まないモデルのみの場合のテスト。
func TestResolveAliasesFromModels_NoFlashNoProModels(t *testing.T) {
	models := []string{
		"models/gemma-2-27b-it",
		"models/text-embedding-004",
	}

	aliases := model.ResolveAliasesFromModels(models)

	// flash/pro がないので空マップか空エントリのみ
	if aliases["flash"] != "" {
		t.Errorf("flash alias should be empty, got %q", aliases["flash"])
	}
	if aliases["pro"] != "" {
		t.Errorf("pro alias should be empty, got %q", aliases["pro"])
	}
}

// TestResolveAliasesFromModels_EmptyModels は空のモデルリストのテスト。
func TestResolveAliasesFromModels_EmptyModels(t *testing.T) {
	aliases := model.ResolveAliasesFromModels([]string{})

	if len(aliases) != 0 {
		t.Errorf("expected empty map, got %v", aliases)
	}
}

// TestResolveAliasesFromModels_PreferLatestFlash は複数の flash モデルがある場合、最後のものを選択するテスト。
// ヒューリスティックとして、リストの末尾にあるモデルが最新版として採用されることを確認する。
func TestResolveAliasesFromModels_MultipleFlashModels(t *testing.T) {
	models := []string{
		"models/gemini-1.5-flash",
		"models/gemini-2.0-flash-exp",
	}

	aliases := model.ResolveAliasesFromModels(models)

	if aliases["flash"] == "" {
		t.Error("flash alias should be resolved")
	}
	// 最後に見つかったflashモデルが選択される
	if aliases["flash"] != "models/gemini-2.0-flash-exp" {
		t.Errorf("flash alias = %q, want %q", aliases["flash"], "models/gemini-2.0-flash-exp")
	}
}

// TestResolveAliasesFromModels_StripPrefix は models/ プレフィックスがある場合も正しく処理するテスト。
func TestResolveAliasesFromModels_StripPrefix(t *testing.T) {
	models := []string{
		"models/gemini-2.0-flash-image-generation",
	}

	aliases := model.ResolveAliasesFromModels(models)

	// models/ プレフィックス付きのモデル名がそのまま格納される
	if aliases["flash"] != "models/gemini-2.0-flash-image-generation" {
		t.Errorf("flash alias = %q, want %q", aliases["flash"], "models/gemini-2.0-flash-image-generation")
	}
}
