package groq

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const baseURL = "https://api.groq.com/openai/v1"

// Client is an authenticated HTTP client for the Groq chat API.
type Client struct {
	apiKey  string
	model   string
	http    *http.Client
	baseURL string
}

func NewClient(apiKey, model string) *Client {
	return &Client{
		apiKey:  apiKey,
		model:   model,
		http:    &http.Client{Timeout: 60 * time.Second},
		baseURL: baseURL,
	}
}

// Chat sends a conversation history to Groq and returns the assistant reply.
func (c *Client) Chat(ctx context.Context, messages []Message) (string, error) {
	if c.apiKey == "" {
		return "", fmt.Errorf("GROQ_API_KEY is not set – add it to your .env or export it")
	}

	body, err := json.Marshal(ChatRequest{
		Model:     c.model,
		Messages:  messages,
		MaxTokens: 2048,
	})
	if err != nil {
		return "", fmt.Errorf("marshalling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("decoding response: %w", err)
	}
	if chatResp.Error != nil {
		return "", fmt.Errorf("Groq API error: %s", chatResp.Error.Message)
	}
	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("Groq returned no choices")
	}
	return chatResp.Choices[0].Message.Content, nil
}
