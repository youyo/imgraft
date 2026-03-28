package reference

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/youyo/imgraft/internal/errs"
	"github.com/youyo/imgraft/internal/imageproc"
)

const (
	// maxRedirects はリダイレクトの最大回数。
	maxRedirects = 3

	// totalTimeout は全体タイムアウト（SPEC.md セクション 10.8 では 20 秒）。
	totalTimeout = 20 * time.Second
)

// testTransport はテスト時に注入できる HTTP トランスポート。nil の場合は http.DefaultTransport を使用する。
var testTransport http.RoundTripper

// SetHTTPTransport はテスト用に HTTP トランスポートを設定する。
// 本番コードからは呼ばないこと。nil を渡すとデフォルトに戻る。
func SetHTTPTransport(t http.RoundTripper) {
	testTransport = t
}

// newRemoteHTTPClient はリモート取得用の HTTP クライアントを生成する。
// リダイレクト上限とタイムアウトを設定する。
func newRemoteHTTPClient() *http.Client {
	redirectCount := 0
	client := &http.Client{
		Timeout: totalTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			redirectCount++
			if redirectCount > maxRedirects {
				return errs.New(errs.ErrReferenceRedirectLimitExceeded,
					fmt.Sprintf("exceeded maximum redirect limit of %d", maxRedirects))
			}
			return nil
		},
	}
	if testTransport != nil {
		client.Transport = testTransport
	}
	return client
}

// LoadRemoteFile は URL から ReferenceImage を取得して返す。
// SPEC.md セクション 10.8 に準拠:
//   - リダイレクト最大 3 回
//   - 全体タイムアウト 20 秒
//   - Content-Length が 20MB を超える場合は拒否
//   - プライベート IP / localhost は ValidateURL で事前拒否
func LoadRemoteFile(ctx context.Context, rawURL string) (*ReferenceImage, error) {
	// URL のセキュリティ検証
	if err := ValidateURL(rawURL); err != nil {
		return nil, err
	}

	// 親コンテキストに加えてタイムアウトも設定（先に終わった方が有効）
	ctx, cancel := context.WithTimeout(ctx, totalTimeout)
	defer cancel()

	client := newRemoteHTTPClient()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, errs.Wrap(errs.ErrReferenceFetchFailed, err)
	}

	resp, err := client.Do(req)
	if err != nil {
		// コンテキストタイムアウト / キャンセル判定
		if ctx.Err() != nil {
			return nil, errs.Wrap(errs.ErrReferenceTimeout, ctx.Err())
		}
		// リダイレクト上限エラーを判定
		if isRedirectLimitError(err) {
			return nil, errs.New(errs.ErrReferenceRedirectLimitExceeded,
				fmt.Sprintf("exceeded maximum redirect limit of %d", maxRedirects))
		}
		return nil, errs.Wrap(errs.ErrReferenceFetchFailed, err)
	}
	defer resp.Body.Close()

	// ステータスコード確認
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errs.New(errs.ErrReferenceFetchFailed,
			fmt.Sprintf("HTTP %d: %s", resp.StatusCode, resp.Status))
	}

	// Content-Length による事前チェック（20MB 制限）
	if resp.ContentLength > MaxFileSizeBytes {
		return nil, errs.New(errs.ErrImageTooLarge,
			fmt.Sprintf("Content-Length %d bytes exceeds limit of %d bytes (20MB)", resp.ContentLength, MaxFileSizeBytes))
	}

	// ボディ読み込み（20MB + 1バイトで制限を超えたら拒否）
	limitedReader := io.LimitReader(resp.Body, MaxFileSizeBytes+1)
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		if ctx.Err() != nil {
			return nil, errs.Wrap(errs.ErrReferenceTimeout, ctx.Err())
		}
		return nil, errs.Wrap(errs.ErrReferenceFetchFailed, err)
	}

	// 読み込んだサイズが制限を超えていないか確認
	if int64(len(data)) > MaxFileSizeBytes {
		return nil, errs.New(errs.ErrImageTooLarge,
			fmt.Sprintf("response body exceeds limit of %d bytes (20MB)", MaxFileSizeBytes))
	}

	// 画像メタデータを取得（デコード検証含む）
	meta, err := imageproc.InspectBytes(data)
	if err != nil {
		return nil, err
	}

	// ファイル名を URL から取得
	filename := filenameFromURL(rawURL)

	return &ReferenceImage{
		SourceType:      "url",
		OriginalInput:   rawURL,
		LocalCachedPath: "",
		Filename:        filename,
		MimeType:        meta.MimeType,
		Width:           meta.Width,
		Height:          meta.Height,
		SizeBytes:       int64(len(data)),
		Data:            data,
	}, nil
}

// filenameFromURL は URL のパス部分からファイル名を取得する。
func filenameFromURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "remote"
	}
	base := path.Base(u.Path)
	if base == "" || base == "." || base == "/" {
		return "remote"
	}
	return base
}

// isRedirectLimitError はエラーがリダイレクト上限エラーかを判定する。
func isRedirectLimitError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "exceeded maximum redirect limit")
}
