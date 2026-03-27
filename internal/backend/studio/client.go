package studio

import (
	"net/http"
	"time"
)

// defaultTimeout is the default HTTP client timeout for API requests.
const defaultTimeout = 60 * time.Second

// HTTPClient is the concrete implementation of StudioClient using HTTP REST calls.
type HTTPClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// New creates an HTTPClient with default settings.
func New(apiKey string) *HTTPClient {
	return &HTTPClient{
		apiKey:  apiKey,
		baseURL: BaseURL,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// NewWithBaseURL creates an HTTPClient with a custom base URL and HTTP client.
// This is primarily for testing with httptest.Server.
func NewWithBaseURL(apiKey, baseURL string, httpClient *http.Client) *HTTPClient {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: defaultTimeout,
		}
	}
	return &HTTPClient{
		apiKey:     apiKey,
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

// Compile-time check that HTTPClient implements StudioClient.
var _ StudioClient = (*HTTPClient)(nil)
