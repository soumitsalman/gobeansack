package nlp

// TODO: convert this into grpc client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const (
	defaultPort         = "8080"
	defaultEmbedModel   = "text-embedding-3-small"
	defaultEmbedBaseURL = "https://api.openai.com"
	embedTimeoutSeconds = 30
)

type openAICompatibleEmbedder struct {
	baseURL string
	apiKey  string
	model   string
	client  *http.Client
}

type embeddingRequest struct {
	Input string `json:"input"`
	Model string `json:"model"`
}

type embeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
}

func (e *openAICompatibleEmbedder) EmbedQuery(ctx context.Context, query string) ([]float32, error) {
	payload, err := json.Marshal(embeddingRequest{Input: query, Model: e.model})
	if err != nil {
		return nil, err
	}
	url := strings.TrimRight(e.baseURL, "/") + "/v1/embeddings"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if e.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+e.apiKey)
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("embedding api returned %s", resp.Status)
	}

	out := embeddingResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if len(out.Data) == 0 || len(out.Data[0].Embedding) == 0 {
		return nil, fmt.Errorf("empty embedding response")
	}
	return out.Data[0].Embedding, nil
}
