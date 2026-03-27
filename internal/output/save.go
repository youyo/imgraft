package output

import (
	"os"
	"path/filepath"

	"github.com/youyo/imgraft/internal/errs"
	"github.com/youyo/imgraft/internal/imageproc"
	"github.com/youyo/imgraft/internal/runtime"
)

// SaveOptions はファイル保存の設定を保持する。
type SaveOptions struct {
	// OutputPath は --output フラグで指定された場合のパス。
	// 空の場合は Dir + 自動命名を使用する。
	OutputPath string

	// Dir は出力ディレクトリ。OutputPath が空の場合に使用する。
	// 空の場合は "." (カレントディレクトリ) を使用する。
	Dir string

	// Clock は自動命名用の時刻プロバイダー。
	// nil の場合は runtime.SystemClock を使用する。
	Clock runtime.Clock

	// Index は images 配列内のインデックス。
	Index int

	// TransparentApplied は透明処理が適用されたかを示す。
	TransparentApplied bool
}

// SavePNG はバイト列をPNGファイルとして保存し、ImageItem を返す。
//
// パイプライン: write → close → inspect → sha256
//
// エラーが発生した場合、書き込み中のファイルを削除して中途半端な状態を残さない。
// エラーコード:
//   - ErrFileWriteFailed: 空データ・書き込み失敗
//   - ErrOutputDirCreateFailed: ディレクトリ作成失敗
//   - ErrFileAlreadyExists: 既存ファイルとの衝突
//   - ErrInvalidOutputPath: 無効な出力パス
//   - ErrInvalidImage: InspectFile での画像解析失敗
func SavePNG(data []byte, opts SaveOptions) (ImageItem, error) {
	// 事前検証: 空データは早期失敗
	if len(data) == 0 {
		return ImageItem{}, errs.New(errs.ErrFileWriteFailed, "PNG data is empty")
	}

	// Clock が nil の場合は SystemClock にフォールバック
	clk := opts.Clock
	if clk == nil {
		clk = runtime.SystemClock{}
	}

	// 最終ファイルパスを決定
	finalPath, err := resolveFinalPath(opts, clk)
	if err != nil {
		return ImageItem{}, err
	}

	// ファイル書き込み → close
	if err := writeFile(finalPath, data); err != nil {
		return ImageItem{}, err
	}

	// inspect (write/close 成功後)
	meta, err := imageproc.InspectFile(finalPath)
	if err != nil {
		// 部分ファイルを削除してクリーンアップ
		_ = os.Remove(finalPath)
		return ImageItem{}, err
	}

	// sha256
	hash, err := imageproc.SHA256OfFile(finalPath)
	if err != nil {
		_ = os.Remove(finalPath)
		return ImageItem{}, err
	}

	// 絶対パスに変換
	absPath, err := filepath.Abs(finalPath)
	if err != nil {
		_ = os.Remove(finalPath)
		return ImageItem{}, errs.Wrap(errs.ErrFileWriteFailed, err)
	}

	return ImageItem{
		Index:              opts.Index,
		Path:               absPath,
		Filename:           filepath.Base(absPath),
		Width:              meta.Width,
		Height:             meta.Height,
		MimeType:           meta.MimeType,
		SHA256:             hash,
		TransparentApplied: opts.TransparentApplied,
	}, nil
}

// resolveFinalPath は保存先のファイルパスを決定する。
// OutputPath が指定されている場合はそれを使用し、既存ファイルの場合はエラー。
// OutputPath が空の場合は Dir + GenerateFilename で自動命名する。
func resolveFinalPath(opts SaveOptions, clk runtime.Clock) (string, error) {
	if opts.OutputPath != "" {
		// --output 指定: ディレクトリ作成 + 既存確認
		dir := filepath.Dir(opts.OutputPath)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return "", errs.Wrap(errs.ErrOutputDirCreateFailed, err)
		}
		if _, err := os.Stat(opts.OutputPath); err == nil {
			return "", errs.New(
				errs.ErrFileAlreadyExists,
				"output file already exists: "+opts.OutputPath,
			)
		}
		return opts.OutputPath, nil
	}

	// 自動命名: Dir が空なら "." を使用
	dir := opts.Dir
	if dir == "" {
		dir = "."
	}

	// ディレクトリを作成
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", errs.Wrap(errs.ErrOutputDirCreateFailed, err)
	}

	return GenerateFilename(dir, clk.Now())
}

// writeFile はファイルを作成してデータを書き込み、クローズする。
// エラー時は部分ファイルを削除する。
func writeFile(path string, data []byte) error {
	f, err := os.Create(path)
	if err != nil {
		return errs.Wrap(errs.ErrFileWriteFailed, err)
	}

	if _, writeErr := f.Write(data); writeErr != nil {
		_ = f.Close()
		_ = os.Remove(path)
		return errs.Wrap(errs.ErrFileWriteFailed, writeErr)
	}

	if closeErr := f.Close(); closeErr != nil {
		_ = os.Remove(path)
		return errs.Wrap(errs.ErrFileWriteFailed, closeErr)
	}

	return nil
}
