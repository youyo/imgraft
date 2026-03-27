package studio

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/youyo/imgraft/internal/errs"
)

// ListModels retrieves available models from the Google AI Studio API.
// It returns all models with SupportedGeneration set to true for models
// that include "generateContent" or "generateImages" in their supported methods.
func (c *HTTPClient) ListModels(ctx context.Context) ([]RemoteModel, error) {
	url := fmt.Sprintf("%s/%s/models?key=%s", c.baseURL, apiVersion, c.apiKey)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, errs.Wrap(errs.ErrInternal, fmt.Errorf("failed to create request: %w", err))
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, mapNetworkError(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errs.Wrap(errs.ErrInternal, fmt.Errorf("failed to read response body: %w", err))
	}

	if resp.StatusCode != http.StatusOK {
		return nil, mapHTTPError(resp.StatusCode, body)
	}

	var listResp listModelsResponse
	if err := json.Unmarshal(body, &listResp); err != nil {
		return nil, errs.Wrap(errs.ErrInternal, fmt.Errorf("failed to parse models response: %w", err))
	}

	// Convert to RemoteModel slice — always return non-nil slice.
	result := make([]RemoteModel, 0, len(listResp.Models))
	for _, m := range listResp.Models {
		result = append(result, RemoteModel{
			Name:                m.Name,
			DisplayName:         m.DisplayName,
			SupportedGeneration: supportsGeneration(m.SupportedGenerationMethods),
		})
	}

	return result, nil
}

// ValidateAPIKey checks if the configured API key is valid by performing
// a lightweight models.list request with pageSize=1.
func (c *HTTPClient) ValidateAPIKey(ctx context.Context) error {
	url := fmt.Sprintf("%s/%s/models?key=%s&pageSize=1", c.baseURL, apiVersion, c.apiKey)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return errs.Wrap(errs.ErrInternal, fmt.Errorf("failed to create request: %w", err))
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return mapNetworkError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errs.Wrap(errs.ErrInternal, fmt.Errorf("failed to read response body: %w", err))
	}

	return mapHTTPError(resp.StatusCode, body)
}

// supportsGeneration checks if a model supports image generation.
func supportsGeneration(methods []string) bool {
	for _, m := range methods {
		if m == "generateContent" || m == "generateImages" {
			return true
		}
	}
	return false
}
