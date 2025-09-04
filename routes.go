package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

var query_throttler chan int

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

func enterQueryThrottler(c *gin.Context) {
	query_throttler <- 1
	c.Next()
}

func exitQueryThrottler(c *gin.Context) {
	<-query_throttler
	c.Next()
}

func initRoutes(ds *Ducksack, throttle_max int) *gin.Engine {
	query_throttler = make(chan int, throttle_max)
	r := gin.Default()
	// Health check endpoint - no auth required
	r.GET("/health", healthCheckHandler)

	// NEWS API ENDPOINTS
	// everything sorted by created_at DESC
	public := r.Group("/public/beans", validateQueryRequest)
	{
		public.GET("/latest", enterQueryThrottler, createLatestBeansHandler(ds), exitQueryThrottler)
		public.GET("/exists", createExistsHandler(ds))
		public.GET("/related", createRelatedBeansHandler(ds))
		public.GET("/regions", createRegionsHandler(ds))
		public.GET("/entities", createEntitiesHandler(ds))
		public.GET("/categories", createCategoriesHandler(ds))
		public.GET("/sources", createSourcesHandler(ds))
	}

	// REGISTERED APPLICATION ENDPOINTS
	// everything sorted by trending DESC
	privileged := r.Group("/privileged/beans", createAuthVerificationHandler("PRIVILEGED_KEY"), validateQueryRequest, enterQueryThrottler)
	{
		privileged.GET("/latest/contents", createLatestContentsHandler(ds), exitQueryThrottler)
		privileged.GET("/trending", createTrendingBeansHandler(ds), exitQueryThrottler)
		privileged.GET("/trending/tags", createTrendingTagsHandler(ds), exitQueryThrottler)
		privileged.GET("/trending/embeddings", createTrendingEmbeddingsHandler(ds), exitQueryThrottler)
	}
	// CONTRIBUTOR ENDPOINTS
	publisher := r.Group("/publisher", createAuthVerificationHandler("PUBLISHER_KEY"))
	{
		publisher.POST("/beans", createStoreBeansHandler(ds))
		publisher.POST("/beans/tags", createStoreTagsHandler(ds))
		publisher.POST("/beans/embeddings", createStoreEmbeddingsHandler(ds))
		publisher.POST("/chatters", createStoreChatterHandler(ds))
		publisher.POST("/sources", createStoreSourceHandler(ds))
		publisher.DELETE("/beans", validateDeleteRequest, createDeleteBeansHandler(ds))
		publisher.DELETE("/chatters", validateDeleteRequest, createDeleteChattersHandler(ds))
		publisher.DELETE("/sources", validateDeleteRequest, createDeleteSourcesHandler(ds))
	}

	// ADMIN ENDPOINTS
	admin := r.Group("/admin", createAuthVerificationHandler("ADMIN_KEY"))
	{
		admin.POST("/execute", createAdminCommandHandler(ds))
	}

	return r
}
