package runtime

import "time"

// Clock は現在時刻を取得する抽象。テスト時に FixedClock を注入できる。
type Clock interface {
	Now() time.Time
}

// SystemClock は time.Now() をそのまま返す本番実装。
type SystemClock struct{}

// Now は現在時刻を返す。
func (SystemClock) Now() time.Time { return time.Now() }

// FixedClock はテスト用の固定時刻実装。
type FixedClock struct {
	T time.Time
}

// Now は固定時刻を返す。
func (c FixedClock) Now() time.Time { return c.T }

// NewFixedClock は指定時刻で固定された Clock を生成する。
func NewFixedClock(t time.Time) FixedClock { return FixedClock{T: t} }
