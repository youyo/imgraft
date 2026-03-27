package studio

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/youyo/imgraft/internal/errs"
)

// --- roundTripFunc enables mocking http.RoundTripper with a function. ---

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// mockClient creates an HTTPClient backed by a round-trip function mock.
func mockClient(fn roundTripFunc) *HTTPClient {
	return NewWithBaseURL("test-key", "https://mock.api", &http.Client{Transport: fn})
}

// --- Test helpers ---

func fakeImageBase64() string {
	return base64.StdEncoding.EncodeToString([]byte("fake-png-data"))
}

func makeGenerateResponse(mimeType, b64data string) []byte {
	resp := generateContentResponse{
		Candidates: []candidate{
			{
				Content: candidateContent{
					Parts: []part{
						{
							InlineData: &inlineData{
								MimeType: mimeType,
								Data:     b64data,
							},
						},
					},
				},
				FinishReason: "STOP",
			},
		},
	}
	data, _ := json.Marshal(resp)
	return data
}

func makeErrorResponse(code int, message, status string) []byte {
	resp := apiErrorResponse{
		Error: apiErrorDetail{
			Code:    code,
			Message: message,
			Status:  status,
		},
	}
	data, _ := json.Marshal(resp)
	return data
}

func makeModelsResponse(models []modelInfo) []byte {
	resp := listModelsResponse{Models: models}
	data, _ := json.Marshal(resp)
	return data
}

func jsonResponse(statusCode int, body []byte, extraHeaders map[string]string) *http.Response {
	h := http.Header{}
	h.Set("Content-Type", "application/json")
	for k, v := range extraHeaders {
		h.Set(k, v)
	}
	return &http.Response{
		StatusCode: statusCode,
		Header:     h,
		Body:       io.NopCloser(bytes.NewReader(body)),
	}
}

// --- T01: Generate success ---

func TestGenerate_Success(t *testing.T) {
	b64 := fakeImageBase64()
	client := mockClient(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", req.Method)
		}
		return jsonResponse(200, makeGenerateResponse("image/png", b64), map[string]string{
			"x-ratelimit-limit-requests": "60",
		}), nil
	})

	resp, header, err := client.Generate(context.Background(), GenerateRequest{
		Model:  "flash",
		Prompt: "test prompt",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp.ImageData) != "fake-png-data" {
		t.Errorf("ImageData = %q, want %q", resp.ImageData, "fake-png-data")
	}
	if resp.MimeType != "image/png" {
		t.Errorf("MimeType = %q, want %q", resp.MimeType, "image/png")
	}
	if header.Get("x-ratelimit-limit-requests") != "60" {
		t.Errorf("expected rate limit header in response")
	}
}

// --- T02: Generate with references ---

func TestGenerate_WithReferences(t *testing.T) {
	b64 := fakeImageBase64()
	var receivedBody generateContentRequest

	client := mockClient(func(req *http.Request) (*http.Response, error) {
		bodyBytes, _ := io.ReadAll(req.Body)
		json.Unmarshal(bodyBytes, &receivedBody)
		return jsonResponse(200, makeGenerateResponse("image/png", b64), nil), nil
	})

	_, _, err := client.Generate(context.Background(), GenerateRequest{
		Model:  "flash",
		Prompt: "convert to icon",
		References: []ReferenceData{
			{MimeType: "image/png", Data: []byte("ref-image-1")},
			{MimeType: "image/jpeg", Data: []byte("ref-image-2")},
		},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(receivedBody.Contents) == 0 {
		t.Fatal("expected contents in request")
	}
	parts := receivedBody.Contents[0].Parts
	if len(parts) != 3 {
		t.Fatalf("expected 3 parts (text + 2 refs), got %d", len(parts))
	}
	if parts[1].InlineData == nil || parts[1].InlineData.MimeType != "image/png" {
		t.Error("expected first reference as image/png inlineData")
	}
	if parts[2].InlineData == nil || parts[2].InlineData.MimeType != "image/jpeg" {
		t.Error("expected second reference as image/jpeg inlineData")
	}
}

// --- T03: Generate with system prompt ---

func TestGenerate_WithSystemPrompt(t *testing.T) {
	b64 := fakeImageBase64()
	var receivedBody generateContentRequest

	client := mockClient(func(req *http.Request) (*http.Response, error) {
		bodyBytes, _ := io.ReadAll(req.Body)
		json.Unmarshal(bodyBytes, &receivedBody)
		return jsonResponse(200, makeGenerateResponse("image/png", b64), nil), nil
	})

	_, _, err := client.Generate(context.Background(), GenerateRequest{
		Model:        "flash",
		Prompt:       "blue robot",
		SystemPrompt: "Generate a single isolated subject asset",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedBody.SystemInstruction == nil {
		t.Fatal("expected systemInstruction in request body")
	}
	if len(receivedBody.SystemInstruction.Parts) == 0 {
		t.Fatal("expected parts in systemInstruction")
	}
	if receivedBody.SystemInstruction.Parts[0].Text != "Generate a single isolated subject asset" {
		t.Errorf("systemInstruction text = %q, want expected value", receivedBody.SystemInstruction.Parts[0].Text)
	}
}

// --- T03b: Generate without system prompt omits field ---

func TestGenerate_NoSystemPrompt(t *testing.T) {
	b64 := fakeImageBase64()
	var receivedBody generateContentRequest

	client := mockClient(func(req *http.Request) (*http.Response, error) {
		bodyBytes, _ := io.ReadAll(req.Body)
		json.Unmarshal(bodyBytes, &receivedBody)
		return jsonResponse(200, makeGenerateResponse("image/png", b64), nil), nil
	})

	_, _, err := client.Generate(context.Background(), GenerateRequest{
		Model:  "flash",
		Prompt: "test",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receivedBody.SystemInstruction != nil {
		t.Error("expected systemInstruction to be nil when not provided")
	}
}

// --- T04: ListModels success ---

func TestListModels_Success(t *testing.T) {
	models := []modelInfo{
		{
			Name:                       "models/gemini-flash",
			DisplayName:                "Gemini Flash",
			SupportedGenerationMethods: []string{"generateContent", "countTokens"},
		},
		{
			Name:                       "models/gemini-pro",
			DisplayName:                "Gemini Pro",
			SupportedGenerationMethods: []string{"generateContent"},
		},
		{
			Name:                       "models/text-only",
			DisplayName:                "Text Only",
			SupportedGenerationMethods: []string{"countTokens"},
		},
	}

	client := mockClient(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", req.Method)
		}
		return jsonResponse(200, makeModelsResponse(models), nil), nil
	})

	result, err := client.ListModels(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 models, got %d", len(result))
	}
	if !result[0].SupportedGeneration {
		t.Error("gemini-flash should support generation")
	}
	if !result[1].SupportedGeneration {
		t.Error("gemini-pro should support generation")
	}
	if result[2].SupportedGeneration {
		t.Error("text-only should not support generation")
	}
}

// --- T05: ValidateAPIKey success ---

func TestValidateAPIKey_Success(t *testing.T) {
	client := mockClient(func(req *http.Request) (*http.Response, error) {
		if req.URL.Query().Get("pageSize") != "1" {
			t.Error("expected pageSize=1 in query")
		}
		return jsonResponse(200, makeModelsResponse([]modelInfo{{Name: "models/test"}}), nil), nil
	})

	err := client.ValidateAPIKey(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- T09: Generate 429 ---

func TestGenerate_429(t *testing.T) {
	client := mockClient(func(req *http.Request) (*http.Response, error) {
		return jsonResponse(429, makeErrorResponse(429, "Resource exhausted", "RESOURCE_EXHAUSTED"), map[string]string{
			"Retry-After": "10",
		}), nil
	})

	_, header, err := client.Generate(context.Background(), GenerateRequest{Model: "flash", Prompt: "test"})

	assertErrorCode(t, err, errs.ErrRateLimitExceeded)
	if header.Get("Retry-After") != "10" {
		t.Error("expected Retry-After header in response")
	}
}

// --- T10: Generate 403 ---

func TestGenerate_403(t *testing.T) {
	client := mockClient(func(req *http.Request) (*http.Response, error) {
		return jsonResponse(403, makeErrorResponse(403, "Permission denied", "PERMISSION_DENIED"), nil), nil
	})

	_, _, err := client.Generate(context.Background(), GenerateRequest{Model: "pro", Prompt: "test"})
	assertErrorCode(t, err, errs.ErrAuthInvalid)
}

// --- T11: Generate 400 ---

func TestGenerate_400(t *testing.T) {
	client := mockClient(func(req *http.Request) (*http.Response, error) {
		return jsonResponse(400, makeErrorResponse(400, "Invalid argument", "INVALID_ARGUMENT"), nil), nil
	})

	_, _, err := client.Generate(context.Background(), GenerateRequest{Model: "flash", Prompt: "test"})
	assertErrorCode(t, err, errs.ErrInvalidArgument)
}

// --- T12: Generate 503 ---

func TestGenerate_503(t *testing.T) {
	client := mockClient(func(req *http.Request) (*http.Response, error) {
		return jsonResponse(503, makeErrorResponse(503, "Service unavailable", "UNAVAILABLE"), nil), nil
	})

	_, _, err := client.Generate(context.Background(), GenerateRequest{Model: "flash", Prompt: "test"})
	assertErrorCode(t, err, errs.ErrBackendUnavailable)
}

// --- T13: Generate 500 ---

func TestGenerate_500(t *testing.T) {
	client := mockClient(func(req *http.Request) (*http.Response, error) {
		return jsonResponse(500, makeErrorResponse(500, "Internal error", "INTERNAL"), nil), nil
	})

	_, _, err := client.Generate(context.Background(), GenerateRequest{Model: "flash", Prompt: "test"})
	assertErrorCode(t, err, errs.ErrInternal)
}

// --- T14: Generate context.Canceled ---

func TestGenerate_ContextCanceled(t *testing.T) {
	client := mockClient(func(req *http.Request) (*http.Response, error) {
		return nil, context.Canceled
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, err := client.Generate(ctx, GenerateRequest{Model: "flash", Prompt: "test"})
	assertErrorCode(t, err, errs.ErrBackendUnavailable)
}

// --- T15: Generate context.DeadlineExceeded ---

func TestGenerate_DeadlineExceeded(t *testing.T) {
	client := mockClient(func(req *http.Request) (*http.Response, error) {
		return nil, context.DeadlineExceeded
	})

	_, _, err := client.Generate(context.Background(), GenerateRequest{Model: "flash", Prompt: "test"})
	assertErrorCode(t, err, errs.ErrBackendUnavailable)
}

// --- T16: Generate no image in response ---

func TestGenerate_NoImageInResponse(t *testing.T) {
	resp := generateContentResponse{
		Candidates: []candidate{
			{
				Content: candidateContent{
					Parts: []part{
						{Text: "Here is a description"},
					},
				},
				FinishReason: "STOP",
			},
		},
	}
	data, _ := json.Marshal(resp)

	client := mockClient(func(req *http.Request) (*http.Response, error) {
		return jsonResponse(200, data, nil), nil
	})

	_, _, err := client.Generate(context.Background(), GenerateRequest{Model: "flash", Prompt: "test"})
	assertErrorCode(t, err, errs.ErrInternal)
}

// --- T17: ValidateAPIKey 401 ---

func TestValidateAPIKey_401(t *testing.T) {
	client := mockClient(func(req *http.Request) (*http.Response, error) {
		return jsonResponse(401, makeErrorResponse(401, "Invalid API key", "UNAUTHENTICATED"), nil), nil
	})

	err := client.ValidateAPIKey(context.Background())
	assertErrorCode(t, err, errs.ErrAuthInvalid)
}

// --- T19: Generate with empty API key ---

func TestGenerate_EmptyAPIKey(t *testing.T) {
	c := NewWithBaseURL("", "https://mock.api", &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return jsonResponse(401, makeErrorResponse(401, "API key not valid", "UNAUTHENTICATED"), nil), nil
		}),
	})

	_, _, err := c.Generate(context.Background(), GenerateRequest{Model: "flash", Prompt: "test"})
	assertErrorCode(t, err, errs.ErrAuthInvalid)
}

// --- T22: ListModels empty list ---

func TestListModels_EmptyList(t *testing.T) {
	client := mockClient(func(req *http.Request) (*http.Response, error) {
		return jsonResponse(200, makeModelsResponse([]modelInfo{}), nil), nil
	})

	result, err := client.ListModels(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("result should not be nil, expected empty slice")
	}
	if len(result) != 0 {
		t.Errorf("expected 0 models, got %d", len(result))
	}
}

// --- T23: Generate empty candidates ---

func TestGenerate_EmptyCandidates(t *testing.T) {
	resp := generateContentResponse{
		Candidates: []candidate{},
	}
	data, _ := json.Marshal(resp)

	client := mockClient(func(req *http.Request) (*http.Response, error) {
		return jsonResponse(200, data, nil), nil
	})

	_, _, err := client.Generate(context.Background(), GenerateRequest{Model: "flash", Prompt: "test"})
	assertErrorCode(t, err, errs.ErrInternal)
}

// --- T_network: Generate network error ---

func TestGenerate_NetworkError(t *testing.T) {
	client := mockClient(func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("connection refused")
	})

	_, _, err := client.Generate(context.Background(), GenerateRequest{Model: "flash", Prompt: "test"})
	assertErrorCode(t, err, errs.ErrBackendUnavailable)
}

// --- Test helper ---

func assertErrorCode(t *testing.T, err error, want errs.ErrorCode) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error with code %q, got nil", want)
	}
	var coded *errs.CodedError
	if !errors.As(err, &coded) {
		t.Fatalf("expected CodedError, got %T: %v", err, err)
	}
	if coded.Code != want {
		t.Errorf("error code = %q, want %q", coded.Code, want)
	}
}
