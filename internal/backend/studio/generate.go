package studio

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/youyo/imgraft/internal/errs"
)

// Generate sends a generateContent request to the Google AI Studio API
// and returns the generated image data along with the raw HTTP response headers.
func (c *HTTPClient) Generate(ctx context.Context, req GenerateRequest) (GenerateResponse, http.Header, error) {
	apiReq := buildGenerateRequest(req)

	body, err := json.Marshal(apiReq)
	if err != nil {
		return GenerateResponse{}, nil, errs.Wrap(errs.ErrInternal, fmt.Errorf("failed to marshal request: %w", err))
	}

	url := fmt.Sprintf("%s/%s/models/%s:generateContent?key=%s", c.baseURL, apiVersion, req.Model, c.apiKey)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return GenerateResponse{}, nil, errs.Wrap(errs.ErrInternal, fmt.Errorf("failed to create request: %w", err))
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return GenerateResponse{}, nil, mapNetworkError(err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return GenerateResponse{}, resp.Header, errs.Wrap(errs.ErrInternal, fmt.Errorf("failed to read response body: %w", err))
	}

	if resp.StatusCode != http.StatusOK {
		return GenerateResponse{}, resp.Header, mapHTTPError(resp.StatusCode, respBody)
	}

	genResp, err := parseGenerateResponse(respBody)
	if err != nil {
		return GenerateResponse{}, resp.Header, err
	}

	return genResp, resp.Header, nil
}

// buildGenerateRequest converts a GenerateRequest into the API JSON structure.
func buildGenerateRequest(req GenerateRequest) generateContentRequest {
	// Build user content parts.
	parts := []part{}

	// Add text prompt.
	if req.Prompt != "" {
		parts = append(parts, part{Text: req.Prompt})
	}

	// Add reference images as inlineData.
	for _, ref := range req.References {
		encoded := base64.StdEncoding.EncodeToString(ref.Data)
		parts = append(parts, part{
			InlineData: &inlineData{
				MimeType: ref.MimeType,
				Data:     encoded,
			},
		})
	}

	apiReq := generateContentRequest{
		Contents: []contentPart{
			{
				Role:  "user",
				Parts: parts,
			},
		},
		GenerationConfig: generationConfig{
			ResponseModalities: []string{"IMAGE", "TEXT"},
		},
	}

	// Add system instruction if provided.
	if req.SystemPrompt != "" {
		apiReq.SystemInstruction = &systemInstruction{
			Parts: []part{
				{Text: req.SystemPrompt},
			},
		}
	}

	return apiReq
}

// parseGenerateResponse parses the API response body and extracts image data.
func parseGenerateResponse(body []byte) (GenerateResponse, error) {
	var resp generateContentResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return GenerateResponse{}, errs.Wrap(errs.ErrInternal, fmt.Errorf("failed to parse response: %w", err))
	}

	if len(resp.Candidates) == 0 {
		return GenerateResponse{}, errs.New(errs.ErrInternal, "no candidates in response")
	}

	candidate := resp.Candidates[0]

	// Find the first inlineData part containing image data.
	for _, p := range candidate.Content.Parts {
		if p.InlineData != nil && p.InlineData.Data != "" {
			decoded, err := base64.StdEncoding.DecodeString(p.InlineData.Data)
			if err != nil {
				return GenerateResponse{}, errs.Wrap(errs.ErrInternal, fmt.Errorf("failed to decode base64 image data: %w", err))
			}
			return GenerateResponse{
				ImageData: decoded,
				MimeType:  p.InlineData.MimeType,
			}, nil
		}
	}

	return GenerateResponse{}, errs.New(errs.ErrInternal, "no image data in response candidates")
}

// mapNetworkError converts network-level errors to errs.CodedError.
func mapNetworkError(err error) *errs.CodedError {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return errs.Wrap(errs.ErrBackendUnavailable, fmt.Errorf("request failed: %w", err))
	}
	return errs.Wrap(errs.ErrBackendUnavailable, fmt.Errorf("network error: %w", err))
}
