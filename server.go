package main

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

type StoreBeansRequest struct {
	Beans []Bean `json:"beans"`
}

type StoreEmbeddingsRequest struct {
	Embeddings []Bean `json:"embeddings"`
}

type StoreTagsRequest struct {
	Tags []TagData `json:"tags"`
	Type string    `json:"type"`
}

type StoreChatterRequest struct {
	Chatters []Chatter `json:"chatters"`
}

type StoreSourceRequest struct {
	Sources []Source `json:"sources"`
}

type QueryBeansRequest struct {
	Kind       string    `json:"kind"`
	Since      time.Time `json:"since"`
	Categories []string  `json:"categories"`
	Regions    []string  `json:"regions"`
	Entities   []string  `json:"entities"`
	Sources    []string  `json:"sources"`
	Offset     int64     `json:"offset"`
	Limit      int64     `json:"limit"`
}

type VectorSearchRequest struct {
	Embedding  Float32Array `json:"embedding"`
	Threshold  float64      `json:"threshold"`
	Kind       string       `json:"kind"`
	Since      time.Time    `json:"since"`
	Categories []string     `json:"categories"`
	Regions    []string     `json:"regions"`
	Entities   []string     `json:"entities"`
	Sources    []string     `json:"sources"`
	Offset     int64        `json:"offset"`
	Limit      int64        `json:"limit"`
}

func healthCheckHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

func requireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-KEY")
		expectedKey := os.Getenv("COFFEEMAKER_KEY")

		if apiKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Missing X-API-KEY header",
			})
			return
		}

		if apiKey != expectedKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid API key",
			})
			return
		}

		c.Next()
	}
}

func storeBeansHandler(ds *Ducksack) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req StoreBeansRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ds.StoreBeans(req.Beans)
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	}
}

func storeEmbeddingsHandler(ds *Ducksack) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req StoreEmbeddingsRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ds.StoreEmbeddings(req.Embeddings)
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	}
}

func storeTagsHandler(ds *Ducksack) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req StoreTagsRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ds.StoreTags(req.Tags, req.Type)
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	}
}

func storeChatterHandler(ds *Ducksack) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req StoreChatterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ds.StoreChatters(req.Chatters)
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	}
}

func storeSourceHandler(ds *Ducksack) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req StoreSourceRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ds.StoreSources(req.Sources)
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	}
}

func queryBeansHandler(ds *Ducksack) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req QueryBeansRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		beans := ds.QueryBeans(
			req.Kind,
			req.Since,
			req.Categories,
			req.Regions,
			req.Entities,
			req.Sources,
			req.Offset,
			req.Limit,
		)

		c.JSON(http.StatusOK, gin.H{
			"beans": beans,
		})
	}
}

func vectorSearchHandler(ds *Ducksack) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req VectorSearchRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		beans := ds.VectorSearchBeans(
			req.Embedding,
			req.Threshold,
			req.Kind,
			req.Since,
			req.Categories,
			req.Regions,
			req.Entities,
			req.Sources,
			req.Offset,
			req.Limit,
		)

		c.JSON(http.StatusOK, gin.H{
			"beans": beans,
		})
	}
}

func setupRouter(ds *Ducksack) *gin.Engine {
	r := gin.Default()

	// Health check endpoint - no auth required
	r.GET("/health", healthCheckHandler)

	// Protected routes - require authentication
	protected := r.Group("/", requireAuth())
	{
		// Store beans
		protected.POST("/beans", storeBeansHandler(ds))

		// Store embeddings
		protected.POST("/embeddings", storeEmbeddingsHandler(ds))

		// Store tags (gists, categories, sentiments)
		protected.POST("/tags", storeTagsHandler(ds))

		// Store chatters
		protected.POST("/chatters", storeChatterHandler(ds))

		// Store sources
		protected.POST("/sources", storeSourceHandler(ds))

		// Query beans with scalar filters
		protected.POST("/beans/query", queryBeansHandler(ds))

		// Vector search
		protected.POST("/beans/vector-search", vectorSearchHandler(ds))
	}

	return r
}

func StartServer(ds *Ducksack, port string) error {
	r := setupRouter(ds)
	return r.Run(":" + port)
}
