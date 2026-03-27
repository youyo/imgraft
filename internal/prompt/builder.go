package prompt

import (
	"context"
	"strings"

	"github.com/youyo/imgraft/internal/errs"
)

// Role 定数。Part.Role に使用する。
const (
	RoleSystem = "system"
	RoleUser   = "user"
)

// Part はプロンプトの1要素（role + text）。
// imgraft 内部型として定義し、M08 のアダプター層で Google API 型に変換する。
type Part struct {
	Role string // RoleSystem | RoleUser
	Text string
}

// Build はユーザープロンプトと transparent フラグからリクエスト用 Parts を組み立てる。
//
// transparent=true の場合、SPEC §11.2 のシステムプロンプトを先頭に追加する。
// transparent=false の場合、システムプロンプトは省略される。
//
// userPrompt が空（trim 後）の場合は INVALID_ARGUMENT エラーを返す。
func Build(_ context.Context, userPrompt string, transparent bool) ([]Part, error) {
	if strings.TrimSpace(userPrompt) == "" {
		return nil, errs.New(errs.ErrInvalidArgument, "user prompt is required")
	}

	var parts []Part

	sys := SystemPrompt(transparent)
	if sys != "" {
		parts = append(parts, Part{Role: RoleSystem, Text: sys})
	}

	parts = append(parts, Part{Role: RoleUser, Text: userPrompt})

	return parts, nil
}
