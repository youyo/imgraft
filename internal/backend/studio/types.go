// Package studio implements a REST client for the Google AI Studio (Gemini) API.
// It provides image generation, model listing, and API key validation.
//
// Design note: Generate returns http.Header directly rather than a parsed RateLimit struct.
// The upper layer (app/run.go in M11) is responsible for parsing rate-limit information
// via the ratelimit package (M16). This package stays within the HTTP layer and returns raw headers.
package studio

import (
	"context"
	"net/http"
)

// BaseURL is the default Google AI Studio API endpoint.
// Gemini API uses /v1beta/ path — this is unrelated to imgraft's own v1 release versioning.
const BaseURL = "https://generativelanguage.googleapis.com"

// apiVersion is the Gemini API version path segment.
const apiVersion = "v1beta"

// StudioClient is the interface for Google AI Studio API operations.
type StudioClient interface {
	// Generate creates an image using the generateContent endpoint.
	// It returns the generated image data, HTTP response headers (for rate-limit parsing),
	// and any error encountered.
	Generate(ctx context.Context, req GenerateRequest) (GenerateResponse, http.Header, error)

	// ListModels retrieves available models from the API.
	ListModels(ctx context.Context) ([]RemoteModel, error)

	// ValidateAPIKey checks whether the configured API key is valid.
	ValidateAPIKey(ctx context.Context) error
}

// GenerateRequest is the input to Generate.
type GenerateRequest struct {
	Model        string
	Prompt       string
	SystemPrompt string
	References   []ReferenceData
}

// ReferenceData holds a reference image in raw bytes.
// Base64 encoding is performed internally when building the API request.
type ReferenceData struct {
	MimeType string
	Data     []byte
}

// GenerateResponse is the output of Generate.
type GenerateResponse struct {
	ImageData []byte // decoded PNG/image bytes
	MimeType  string
}

// RemoteModel represents a model returned by the models.list endpoint.
type RemoteModel struct {
	Name                string
	DisplayName         string
	SupportedGeneration bool // true if the model supports generateContent or image generation
}

// --- Internal JSON types for API request/response serialization ---

// generateContentRequest is the JSON body sent to the generateContent endpoint.
type generateContentRequest struct {
	Contents         []contentPart     `json:"contents"`
	SystemInstruction *systemInstruction `json:"systemInstruction,omitempty"`
	GenerationConfig  generationConfig  `json:"generationConfig"`
}

type contentPart struct {
	Role  string `json:"role"`
	Parts []part `json:"parts"`
}

type part struct {
	Text       string      `json:"text,omitempty"`
	InlineData *inlineData `json:"inlineData,omitempty"`
}

type inlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"` // base64 encoded
}

type systemInstruction struct {
	Parts []part `json:"parts"`
}

type generationConfig struct {
	ResponseModalities []string `json:"responseModalities"`
}

// generateContentResponse is the JSON body returned by the generateContent endpoint.
type generateContentResponse struct {
	Candidates []candidate `json:"candidates"`
}

type candidate struct {
	Content      candidateContent `json:"content"`
	FinishReason string           `json:"finishReason"`
}

type candidateContent struct {
	Parts []part `json:"parts"`
}

// listModelsResponse is the JSON body returned by the models.list endpoint.
type listModelsResponse struct {
	Models        []modelInfo `json:"models"`
	NextPageToken string      `json:"nextPageToken,omitempty"`
}

type modelInfo struct {
	Name                       string   `json:"name"`
	DisplayName                string   `json:"displayName"`
	SupportedGenerationMethods []string `json:"supportedGenerationMethods"`
}

// apiErrorResponse is the JSON body returned on API errors.
type apiErrorResponse struct {
	Error apiErrorDetail `json:"error"`
}

type apiErrorDetail struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

// RateLimitInfo holds parsed rate-limit information from HTTP response headers.
// This is a temporary type within the studio package; M16 will implement
// internal/ratelimit as the canonical package and this will be refactored.
type RateLimitInfo struct {
	Provider          *string
	LimitType         *string
	RequestsLimit     *int
	RequestsRemaining *int
	RequestsUsed      *int
	ResetAt           *string
	RetryAfterSeconds *int
}
