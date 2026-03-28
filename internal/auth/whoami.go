package auth

import (
	"fmt"

	"github.com/youyo/imgraft/internal/config"
)

// WhoamiOptions は Whoami 関数の入力パラメータ。
type WhoamiOptions struct {
	// CredPath は credentials.json のパス。空なら既定パスを使用。
	CredPath string

	// ConfigPath は config.toml のパス。空なら既定パスを使用。
	ConfigPath string
}

// BackendInfo は whoami の backend 情報。
type BackendInfo struct {
	Name      string
	MaskedKey string
}

// WhoamiResult は Whoami の結果。
type WhoamiResult struct {
	Profile         string
	LastUsedProfile string
	LastUsedBackend string
	Backends        []BackendInfo
}

// String は whoami の出力フォーマットを返す。
func (r WhoamiResult) String() string {
	s := fmt.Sprintf("Profile: %s\n", r.Profile)
	s += fmt.Sprintf("Last used backend: %s\n", r.LastUsedBackend)
	s += "\nAvailable backends:\n"
	if len(r.Backends) == 0 {
		s += "  (none)\n"
	} else {
		for _, b := range r.Backends {
			s += fmt.Sprintf("- %s (api_key: %s)\n", b.Name, b.MaskedKey)
		}
	}
	return s
}

// Whoami は現在の認証状態を返す。
// 表示内容:
//   - current profile
//   - last used profile
//   - last used backend
//   - current profile で利用可能な backend（API key は末尾4文字のみ）
func Whoami(opts WhoamiOptions) (WhoamiResult, error) {
	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		return WhoamiResult{}, fmt.Errorf("load config: %w", err)
	}

	creds, err := Load(opts.CredPath)
	if err != nil {
		return WhoamiResult{}, fmt.Errorf("load credentials: %w", err)
	}

	result := WhoamiResult{
		Profile:         cfg.CurrentProfile,
		LastUsedProfile: cfg.LastUsedProfile,
		LastUsedBackend: cfg.LastUsedBackend,
	}

	// current profile の backend 情報を収集
	if pc, ok := creds.Profiles[cfg.CurrentProfile]; ok {
		result.Backends = collectBackendInfo(pc)
	}

	return result, nil
}

// collectBackendInfo は ProfileCredentials から利用可能な BackendInfo を収集する。
func collectBackendInfo(pc ProfileCredentials) []BackendInfo {
	var backends []BackendInfo

	if pc.GoogleAIStudio != nil && pc.GoogleAIStudio.APIKey != "" {
		backends = append(backends, BackendInfo{
			Name:      "google_ai_studio",
			MaskedKey: MaskAPIKey(pc.GoogleAIStudio.APIKey),
		})
	}

	return backends
}
