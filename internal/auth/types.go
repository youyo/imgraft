package auth

// Credentials は ~/.config/imgraft/credentials.json の構造体表現。
type Credentials struct {
	Profiles map[string]ProfileCredentials `json:"profiles"`
}

// ProfileCredentials は特定 profile の backend 認証情報セット。
// v1 では google_ai_studio のみだが、将来の拡張性のため backend ごとのネストを残す。
type ProfileCredentials struct {
	GoogleAIStudio *GoogleAIStudioCredentials `json:"google_ai_studio,omitempty"`
}

// GoogleAIStudioCredentials は Google AI Studio の認証情報。
type GoogleAIStudioCredentials struct {
	APIKey string `json:"api_key"`
}

// DefaultCredentials はすべてのフィールドが既定値で埋まった Credentials を返す。
// Profiles は空の map で初期化される（nil ではない）。
func DefaultCredentials() *Credentials {
	return &Credentials{
		Profiles: make(map[string]ProfileCredentials),
	}
}
