package reference

import (
	"context"
	"fmt"
	"strings"

	"github.com/youyo/imgraft/internal/errs"
)

// isURL は path が URL かどうかを判定する。
// http:// または https:// で始まる場合は URL とみなす。
func isURL(path string) bool {
	lower := strings.ToLower(path)
	return strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://")
}

// LoadReferences は paths リストを順番に読み込み、ReferenceImage のスライスを返す。
//   - http:// または https:// 始まりは LoadRemoteFile で処理
//   - それ以外は LoadLocalFile で処理
//   - 順序はそのまま保持する
//   - 最大 8 枚チェックはロード前に実施
//   - 1 枚でも失敗したら fail-fast でエラーを返す
//
// SPEC.md セクション 10.4 を参照。
func LoadReferences(ctx context.Context, paths []string) ([]*ReferenceImage, error) {
	// 事前に枚数チェック（ロード前に早期失敗）
	if len(paths) > MaxReferenceCount {
		return nil, errs.New(errs.ErrInvalidArgument,
			fmt.Sprintf("too many reference images: %d (max %d)", len(paths), MaxReferenceCount))
	}

	refs := make([]*ReferenceImage, 0, len(paths))

	for _, p := range paths {
		if isURL(p) {
			ref, err := LoadRemoteFile(ctx, p)
			if err != nil {
				return nil, err
			}
			refs = append(refs, ref)
		} else {
			localRef, err := LoadLocalFile(p)
			if err != nil {
				return nil, err
			}
			refs = append(refs, &localRef)
		}
	}

	return refs, nil
}
