package auth

// MaskAPIKey は api_key の末尾4文字のみ表示し、残りを "*" でマスクする。
// key が4文字以下の場合は "****" を返す。
//
// 前提: Google AI Studio の API key は ASCII 文字のみで構成される。
// したがって len(key) はバイト長と文字数が一致し、バイトスライスで安全に操作できる。
func MaskAPIKey(key string) string {
	if len(key) <= 4 {
		return "****"
	}
	return "****" + key[len(key)-4:]
}

// HasBackend は ProfileCredentials が指定 backend の有効な認証情報を持つか判定する。
// API key が空文字の場合は false を返す。
func HasBackend(pc ProfileCredentials, backend string) bool {
	switch backend {
	case "google_ai_studio":
		return pc.GoogleAIStudio != nil && pc.GoogleAIStudio.APIKey != ""
	default:
		return false
	}
}

// GetAPIKey は ProfileCredentials から指定 backend の API key を取得する。
// 認証情報が存在し API key が非空の場合は (key, true) を返す。
// 存在しない場合は ("", false) を返す。
func GetAPIKey(pc ProfileCredentials, backend string) (string, bool) {
	switch backend {
	case "google_ai_studio":
		if pc.GoogleAIStudio != nil && pc.GoogleAIStudio.APIKey != "" {
			return pc.GoogleAIStudio.APIKey, true
		}
		return "", false
	default:
		return "", false
	}
}

// AvailableBackends は ProfileCredentials で利用可能な backend 名のリストを返す。
func AvailableBackends(pc ProfileCredentials) []string {
	var backends []string
	if HasBackend(pc, "google_ai_studio") {
		backends = append(backends, "google_ai_studio")
	}
	return backends
}
