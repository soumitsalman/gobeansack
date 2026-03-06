package router

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gin-gonic/gin"
	bs "github.com/soumitsalman/gobeansack/beansack"
	"github.com/soumitsalman/gobeansack/nlp"
)

const (
	NAME             = "Beans API & MCP"
	VERSION          = "0.1"
	DESCRIPTION      = "Beans is an intelligent news & blogs aggregation and search service that curates fresh content from RSS feeds using AI-powered natural language queries and filters."
	DEFAULT_ACCURACY = 0.75
	FAVICON_PATH     = "https://cafecito-assets.t3.storage.dev/images/beans.png"
)

type HealthOutput struct {
	Body []map[string]any
}

type TagsInput struct {
	Offset int `query:"offset" default:"0" minimum:"0"`
	Limit  int `query:"limit" default:"16" minimum:"1" maximum:"100"`
}

type StringListOutput struct {
	Body []string
}

type ArticlesInput struct {
	Q              string    `query:"q" minLength:"3" maxLength:"512"`
	Acc            float64   `query:"acc" default:"0.75" minimum:"0" maximum:"1"`
	Kind           string    `query:"kind" enum:"news,blog"`
	Tags           []string  `query:"tags,explode"`
	Sources        []string  `query:"sources,explode"`
	PublishedSince time.Time `query:"published_since" format:"date-time"`
	TrendingSince  time.Time `query:"trending_since" format:"date-time"`
	WithContent    bool      `query:"with_content" default:"false"`
	Limit          int       `query:"limit" default:"16" minimum:"1" maximum:"100"`
	Offset         int       `query:"offset" default:"0" minimum:"0"`
}

type LatestArticlesOutput struct {
	Body []bs.Bean
}

type TrendingArticlesOutput struct {
	Body []bs.BeanAggregate
}

type PublishersInput struct {
	Sources []string `query:"sources,explode"`
	Limit   int      `query:"limit" default:"16" minimum:"1" maximum:"100"`
	Offset  int      `query:"offset" default:"0" minimum:"0"`
}

type PublishersOutput struct {
	Body []bs.Publisher
}

type Configuration struct {
	DB       bs.Beansack
	Embedder nlp.Embedder
	APIKeys  map[string]string
	// Queue controls the number of concurrent requests handled by the
	// service.  The channel is created in `main` based on the
	// MAX_CONCURRENT_REQUESTS environment variable; it is used by a
	// middleware that blocks when the channel is full, effectively
	// queueing excess callers until capacity frees up.
	Queue chan int
}

// health is a no-input handler used by huma.Get. the empty struct
// parameter is required by the generic signature.
func (r *Configuration) health(ctx context.Context, _ *struct{}) (*HealthOutput, error) {
	out := &HealthOutput{Body: []map[string]any{{"status": "alive"}}}
	return out, nil
}

// faviconHuma is a Huma handler (no input) that performs the same redirect.
func (r *Configuration) favicon(ctx context.Context, _ *struct{}) (*struct{}, error) {
	ginCtx := humagin.Unwrap(ctx.(huma.Context))
	ginCtx.Redirect(http.StatusFound, FAVICON_PATH)
	return nil, nil
}

func (r *Configuration) getCategories(ctx context.Context, input *TagsInput) (*StringListOutput, error) {
	data, err := r.DB.DistinctCategories(ctx, bs.Pagination{Limit: input.Limit, Offset: input.Offset})
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to query categories", err)
	}
	return &StringListOutput{Body: data}, nil
}
func (r *Configuration) getEntities(ctx context.Context, input *TagsInput) (*StringListOutput, error) {
	data, err := r.DB.DistinctEntities(ctx, bs.Pagination{Limit: input.Limit, Offset: input.Offset})
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to query entities", err)
	}
	return &StringListOutput{Body: data}, nil
}
func (r *Configuration) getRegions(ctx context.Context, input *TagsInput) (*StringListOutput, error) {
	data, err := r.DB.DistinctRegions(ctx, bs.Pagination{Limit: input.Limit, Offset: input.Offset})
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to query regions", err)
	}
	return &StringListOutput{Body: data}, nil
}

func (r *Configuration) getPublishers(ctx context.Context, input *PublishersInput) (*PublishersOutput, error) {
	if len(input.Sources) == 0 {
		return nil, huma.Error400BadRequest("sources is required")
	}
	items, err := r.DB.QueryPublishers(ctx, bs.Condition{Sources: input.Sources}, bs.Pagination{Limit: input.Limit, Offset: input.Offset}, []string{bs.CORE_PUBLISHER_FIELDS})
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to query publishers", err)
	}
	return &PublishersOutput{Body: items}, nil
}

func (r *Configuration) getSources(ctx context.Context, input *TagsInput) (*StringListOutput, error) {
	items, err := r.DB.DistinctSources(ctx, bs.Pagination{Limit: input.Limit, Offset: input.Offset})
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to query publisher IDs", err)
	}
	return &StringListOutput{Body: items}, nil
}

func (r *Configuration) getLatestArticles(ctx context.Context, input *ArticlesInput) (*LatestArticlesOutput, error) {
	conditions := r.prepareBeanConditions(ctx, input)
	columns := []string{bs.CORE_BEAN_FIELDS}
	if input.WithContent {
		columns = []string{bs.K_CONTENT}
	}
	items, err := r.DB.QueryLatestBeans(ctx, *conditions, bs.Pagination{Limit: input.Limit, Offset: input.Offset}, columns)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to query latest articles", err)
	}
	return &LatestArticlesOutput{Body: items}, nil
}

func (r *Configuration) getTrendingArticles(ctx context.Context, input *ArticlesInput) (*TrendingArticlesOutput, error) {
	conditions := r.prepareBeanConditions(ctx, input)
	columns := []string{bs.CORE_BEAN_FIELDS, bs.K_TRENDSCORE}
	if input.WithContent {
		columns = []string{bs.K_CONTENT}
	}
	items, err := r.DB.QueryTrendingBeans(ctx, *conditions, bs.Pagination{Limit: input.Limit, Offset: input.Offset}, columns)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to query trending articles", err)
	}
	return &TrendingArticlesOutput{Body: items}, nil
}

type humaMiddleware func(ctx huma.Context, next func(huma.Context))

func (r *Configuration) verifyAPIKey(ctx huma.Context) bool {
	if len(r.APIKeys) == 0 {
		return true
	}
	for header, expected := range r.APIKeys {
		if strings.TrimSpace(ctx.Header(header)) == expected {
			return true
		}
	}
	return false
}

func createAPIKeyMiddleware(r *Configuration, api huma.API) humaMiddleware {
	return func(ctx huma.Context, next func(huma.Context)) {
		if !r.verifyAPIKey(ctx) {
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "Missing API Key")
			return
		}
		next(ctx)
	}
}

func createConcurrencyMiddleware(r *Configuration) humaMiddleware {
	if r.Queue == nil {
		r.Queue = make(chan int, 1) // default: 1 item at a time
	}
	return func(ctx huma.Context, next func(huma.Context)) {
		if r.Queue != nil {
			r.Queue <- 1
			defer func() { <-r.Queue }()
		}
		next(ctx)
	}
}

// concurrencyMiddleware enforces a maximum number of concurrent
// requests.  It is registered on the top‑level Huma router so that every
// operation (including `/health`) goes through it.  When the buffer is
// full the middleware will block on send; callers are naturally queued by
// the Go scheduler until a slot becomes available.

func NewRouter(config *Configuration) *gin.Engine {
	router := gin.Default()

	// huma doesnt handle redirect well so directing from gin
	router.GET("/favicon.ico", func(c *gin.Context) {
		c.Redirect(http.StatusFound, FAVICON_PATH)
	})

	api := humagin.New(router, huma.DefaultConfig(NAME, VERSION))
	api.OpenAPI().Info.Description = DESCRIPTION

	huma.Get(api, "/health", config.health)

	protected := huma.NewGroup(api)
	protected.UseMiddleware(
		createAPIKeyMiddleware(config, api),
		createConcurrencyMiddleware(config),
	)

	huma.Register(
		protected,
		huma.Operation{
			Method:      http.MethodGet,
			Path:        "/tags/categories",
			OperationID: "get-tags-get-categories",
			Summary:     "List categories",
			Description: "Retrieves a list of unique values of articles categories/topics, such as Artificial Intelligence, Cybersecurity, Politics, Software Engineering etc.",
		},
		config.getCategories,
	)
	huma.Register(
		protected,
		huma.Operation{
			Method:      http.MethodGet,
			Path:        "/tags/entities",
			OperationID: "get-tags-get-entities",
			Summary:     "List entities",
			Description: "Retrieves a list of unique values of named entities (people, organizations, products) mentioned in the articles.",
		},
		config.getEntities,
	)
	huma.Register(
		protected,
		huma.Operation{
			Method:      http.MethodGet,
			Path:        "/tags/regions",
			OperationID: "get-tags-get-regions",
			Summary:     "List regions",
			Description: "Retrieves a list of unique values of geographic regions mentioned in the articles such as UK, US, Europe etc.",
		},
		config.getRegions,
	)

	huma.Register(
		protected,
		huma.Operation{
			Method:      http.MethodGet,
			Path:        "/publishers",
			OperationID: "get-publishers",
			Summary:     "Get publishers' metadata",
			Description: "Retrieves publisher metadata filtered by one or more publisher IDs.",
		},
		config.getPublishers,
	)
	huma.Register(
		protected,
		huma.Operation{
			Method:      http.MethodGet,
			Path:        "/publishers/sources",
			OperationID: "get-publishers-ids",
			Summary:     "List publisher IDs",
			Description: "Retrieves a list of unique values of publisher IDs/sources from which the articles are sourced.",
		},
		config.getSources,
	)

	huma.Register(
		protected,
		huma.Operation{
			Method:      http.MethodGet,
			Path:        "/articles/latest",
			OperationID: "get-latest-articles",
			Summary:     "Search latest articles",
			Description: "Searches for the latest articles/news/blogs. For vector search (when `q` is provided), the results are sorted by relevance. Otherwise, they are sorted by publication date in descending order (newest first).",
		},
		config.getLatestArticles,
	)
	huma.Register(
		protected,
		huma.Operation{
			Method:      http.MethodGet,
			Path:        "/articles/trending",
			OperationID: "get-articles-trending",
			Summary:     "Search trending articles",
			Description: "Searches for the trending articles/news/blogs. For vector search (when `q` is provided), the results are sorted by relevance. Otherwise, they are sorted by internal trend score (calculated from social media engagement: comments, likes, shares, last engagement etc.).",
		},
		config.getTrendingArticles,
	)

	return router
}

func (config *Configuration) prepareBeanConditions(ctx context.Context, input *ArticlesInput) *bs.Condition {
	conditions := bs.Condition{
		Kind:    input.Kind,
		Created: input.PublishedSince,
		Updated: input.TrendingSince,
		Tags:    input.Tags,
		Sources: input.Sources,
		// TODO: add vector
		Extra: []string{bs.PROCESSED_BEANS_CONDITIONS},
	}
	if input.WithContent {
		conditions.Extra = append(conditions.Extra, bs.UNRESTRICTED_CONTENT_CONDITIONS)
	}
	if input.Q != "" {
		conditions.Embedding = config.Embedder.EmbedQuery(ctx, input.Q)
		conditions.Distance = 1 - input.Acc
	}
	return &conditions
}
