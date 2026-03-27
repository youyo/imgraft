package output

// Output is the root JSON schema for imgraft's stdout.
// Both success and failure cases use the same schema.
// Fields are never omitted; missing values are represented as null.
// See SPEC.md section 14.2.
type Output struct {
	Success   bool        `json:"success"`
	Model     *string     `json:"model"`
	Backend   *string     `json:"backend"`
	Images    []ImageItem `json:"images"`
	RateLimit RateLimit   `json:"rate_limit"`
	Warnings  []string    `json:"warnings"`
	Error     OutputError `json:"error"`
}

// ImageItem holds metadata about a single generated image.
// See SPEC.md section 14.2.
type ImageItem struct {
	Index              int    `json:"index"`
	Path               string `json:"path"`
	Filename           string `json:"filename"`
	Width              int    `json:"width"`
	Height             int    `json:"height"`
	MimeType           string `json:"mime_type"`
	SHA256             string `json:"sha256"`
	TransparentApplied bool   `json:"transparent_applied"`
}

// OutputError holds error information in the JSON output.
// On success, both Code and Message are nil (JSON: null).
// See SPEC.md section 14.2.
type OutputError struct {
	Code    *string `json:"code"`
	Message *string `json:"message"`
}

// RateLimit holds API rate limit information.
// Fields that cannot be determined are nil (JSON: null).
// The object itself is always present in the output.
// See SPEC.md section 15.
type RateLimit struct {
	Provider          *string `json:"provider"`
	LimitType         *string `json:"limit_type"`
	RequestsLimit     *int    `json:"requests_limit"`
	RequestsRemaining *int    `json:"requests_remaining"`
	RequestsUsed      *int    `json:"requests_used"`
	ResetAt           *string `json:"reset_at"`
	RetryAfterSeconds *int    `json:"retry_after_seconds"`
}
