package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func healthCheckHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

func createAuthVerificationHandler(expectedKeyName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-KEY")
		expectedKey := os.Getenv(expectedKeyName)

		if expectedKey == "" || apiKey == expectedKey {
			c.Next()
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
		}
	}
}

func parseRequestParams(c *gin.Context) {
	var req BeanSearchRequest
	err := c.ShouldBindQuery(&req)
	fmt.Println("query", err)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err = c.ShouldBindJSON(&req)
	fmt.Println("json", err != nil && err.Error() != "EOF")
	if err != nil && err.Error() != "EOF" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(req.Embedding) > 0 && req.Limit == 0 && req.MaxDistance == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "For vector search, you must provide either a limit or a max_distance"})
		return
	}
	if req.Limit == 0 {
		req.Limit = DEFAULT_LIMIT
	}
	c.Set("req", req)
	c.Next()
}

func trendingBeanVectorSearchValidation(c *gin.Context) {
	req := c.MustGet("req").(BeanSearchRequest)
	if len(req.Embedding) > 0 && req.MaxDistance == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "For vector search, you must provide a max_distance"})
		return
	}
	c.Next()
}

func setupRoutes(ds *Ducksack) *gin.Engine {
	r := gin.Default()

	// Health check endpoint - no auth required
	r.GET("/health", healthCheckHandler)

	// NEWS API ENDPOINTS
	// everything sorted by created_at DESC
	articles := r.Group("/articles", parseRequestParams)
	{
		articles.GET("/latest", createBeansSearchHandler(ds))
		articles.GET("/related", createRelatedBeansHandler(ds))
		articles.GET("/regions", createRegionsHandler(ds))
		articles.GET("/entities", createEntitiesHandler(ds))
		articles.GET("/categories", createCategoriesHandler(ds))
		articles.GET("/sources", createSourcesHandler(ds))
	}

	// CONTRIBUTOR ENDPOINTS
	contibutor := r.Group("/", createAuthVerificationHandler("CONTRIBUTOR_KEY"))
	{
		contibutor.POST("/beans", createStoreBeansHandler(ds))
		contibutor.POST("/beans/embeddings", createStoreEmbeddingsHandler(ds))
		contibutor.POST("/beans/tags", createStoreTagsHandler(ds))
		contibutor.POST("/chatters", createStoreChatterHandler(ds))
		contibutor.POST("/sources", createStoreSourceHandler(ds))
	}

	// REGISTERED APPLICATION ENDPOINTS
	// everything sorted by trending DESC
	regapp := r.Group("/", createAuthVerificationHandler("REG_APP_KEY"), parseRequestParams, trendingBeanVectorSearchValidation)
	{
		regapp.GET("/beans/exists", createExistsHandler(ds))
		regapp.GET("/beans/trending", createTrendingBeansHandler(ds))
		regapp.GET("/beans/trending/gists", createTrendingBeanGistsHandler(ds))
		regapp.GET("/beans/trending/embeddings", createTrendingBeanEmbeddingsHandler(ds))
	}

	// ADMIN ENDPOINTS
	admin := r.Group("/admin", createAuthVerificationHandler("ADMIN_KEY"))
	{
		admin.POST("/execute", createAdminCommandHandler(ds))
	}

	return r
}
