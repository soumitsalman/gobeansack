package nlp

// TODO: convert this into grpc client

import (
	"context"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	log "github.com/sirupsen/logrus"
)

const (
	_TIMEOUT = 10
)

type Embedder interface {
	EmbedQuery(ctx context.Context, query string) []float32
	EmbedDocuments(ctx context.Context, docs []string) [][]float32
}

type RemoteEmbedder struct {
	client openai.Client
	model  string
}

func NewRemoteEmbedder(base_url, api_key, model string) *RemoteEmbedder {
	options := make([]option.RequestOption, 0, 2)
	if base_url != "" {
		options = append(options, option.WithBaseURL(base_url))
	}
	if api_key != "" {
		options = append(options, option.WithAPIKey(api_key))
	}
	embedder := &RemoteEmbedder{
		client: openai.NewClient(options...),
		model:  model,
	}
	return embedder
}

func (e *RemoteEmbedder) EmbedQuery(ctx context.Context, query string) []float32 {
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, time.Duration(_TIMEOUT)*time.Minute)
	defer cancel()

	resp, err := e.client.Embeddings.New(ctx, openai.EmbeddingNewParams{
		Input: openai.EmbeddingNewParamsInputUnion{OfString: openai.String(query)},
		Model: e.model,
	})
	if err != nil {
		log.Error(err)
		return nil
	}
	if len(resp.Data) == 0 || len(resp.Data[0].Embedding) == 0 {
		log.Error("empty embedding response")
		return nil
	}
	return toFloat32Slice(resp.Data[0].Embedding)
}

func (e *RemoteEmbedder) EmbedDocuments(ctx context.Context, docs []string) [][]float32 {
	panic("Not Implemented")
}

func toFloat32Slice(in []float64) []float32 {
	embedding := make([]float32, len(in))
	for i, v := range in {
		embedding[i] = float32(v)
	}
	return embedding
}

// type openAICompatibleEmbedder struct {
// 	baseURL string
// 	apiKey  string
// 	model   string
// 	client  *http.Client
// }

// type embeddingRequest struct {
// 	Input string `json:"input"`
// 	Model string `json:"model"`
// }

// type embeddingResponse struct {
// 	Data []struct {
// 		Embedding []float32 `json:"embedding"`
// 	} `json:"data"`
// }

// func (e *openAICompatibleEmbedder) EmbedQuery(ctx context.Context, query string) ([]float32, error) {
// 	payload, err := json.Marshal(embeddingRequest{Input: query, Model: e.model})
// 	if err != nil {
// 		return nil, err
// 	}
// 	url := strings.TrimRight(e.baseURL, "/") + "/v1/embeddings"
// 	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
// 	if err != nil {
// 		return nil, err
// 	}
// 	req.Header.Set("Content-Type", "application/json")
// 	if e.apiKey != "" {
// 		req.Header.Set("Authorization", "Bearer "+e.apiKey)
// 	}

// 	resp, err := e.client.Do(req)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()
// 	if resp.StatusCode >= 300 {
// 		return nil, fmt.Errorf("embedding api returned %s", resp.Status)
// 	}

// 	out := embeddingResponse{}
// 	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
// 		return nil, err
// 	}
// 	if len(out.Data) == 0 || len(out.Data[0].Embedding) == 0 {
// 		return nil, fmt.Errorf("empty embedding response")
// 	}
// 	return out.Data[0].Embedding, nil
// }
