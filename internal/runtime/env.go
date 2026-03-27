package runtime

import "os"

// GetWithDefault は環境変数を取得し、未設定なら defaultVal を返す。
// 空文字でセットされている場合は空文字を返す（デフォルト値に戻さない）。
func GetWithDefault(key, defaultVal string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		return defaultVal
	}
	return v
}
