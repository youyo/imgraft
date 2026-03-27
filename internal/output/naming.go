package output

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/youyo/imgraft/internal/errs"
)

const (
	// maxSequence は衝突回避の最大連番。001..999 まで試行する。
	maxSequence = 999
)

// GenerateFilename は time t に基づいてファイル名を生成し、
// dir 配下で未使用のフルパスを返す。
//
// ファイル名フォーマット: imgraft-YYYYMMDD-HHMMSS-XXX.png
// 例: imgraft-20260324-153012-001.png
//
// 同名ファイルが存在する場合は連番をインクリメントして衝突を回避する。
// 001〜999 まで全スロットが使用済みの場合は ErrFileAlreadyExists を返す。
//
// Note: Stat→Create 間に TOCTOU レース条件があるが、v1 はシングルスレッド
// 実行を前提とするため許容範囲とする。
func GenerateFilename(dir string, t time.Time) (string, error) {
	date := t.Format("20060102")
	clock := t.Format("150405")

	for seq := 1; seq <= maxSequence; seq++ {
		name := fmt.Sprintf("imgraft-%s-%s-%03d.png", date, clock, seq)
		path := filepath.Join(dir, name)

		_, err := os.Stat(path)
		if os.IsNotExist(err) {
			// 未使用スロット発見
			return path, nil
		}
		if err != nil {
			// Stat 自体が失敗した場合（権限エラー等）
			return "", errs.Wrap(errs.ErrFileWriteFailed, err)
		}
		// ファイルが存在する場合は次の連番を試す
	}

	return "", errs.New(
		errs.ErrFileAlreadyExists,
		fmt.Sprintf("all filename slots are taken for %s-%s (001..%03d)", date, clock, maxSequence),
	)
}
