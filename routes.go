package main

import (
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

func setupRoutes(ds *Ducksack) *gin.Engine {
	r := gin.Default()

	// Health check endpoint - no auth required
	r.GET("/health", healthCheckHandler)

	// NEWS API ENDPOINTS
	// everything sorted by created_at DESC
	articles := r.Group("/articles", validateBeansQueryRequest)
	{
		articles.GET("/latest", createLatestBeansHandler(ds))
		articles.GET("/related", createRelatedBeansHandler(ds))
		articles.GET("/regions", createRegionsHandler(ds))
		articles.GET("/entities", createEntitiesHandler(ds))
		articles.GET("/categories", createCategoriesHandler(ds))
		articles.GET("/sources", createSourcesHandler(ds))
	}

	// CONTRIBUTOR ENDPOINTS
	contibutor := r.Group("/", createAuthVerificationHandler("CONTRIBUTOR_KEY"))
	{
		contibutor.GET("/beans/missing", createQueryBeansWithMissingTagsHandler(ds))
		contibutor.POST("/beans", createStoreBeansHandler(ds))
		contibutor.POST("/beans/embeddings", createStoreEmbeddingsHandler(ds))
		contibutor.POST("/beans/tags", createStoreTagsHandler(ds))
		contibutor.POST("/chatters", createStoreChatterHandler(ds))
		contibutor.POST("/sources", createStoreSourceHandler(ds))
		contibutor.DELETE("/beans", validateDeleteRequest, createDeleteBeansHandler(ds))
		contibutor.DELETE("/chatters", validateDeleteRequest, createDeleteChattersHandler(ds))
		contibutor.DELETE("/sources", validateDeleteRequest, createDeleteSourcesHandler(ds))
	}

	// REGISTERED APPLICATION ENDPOINTS
	// everything sorted by trending DESC
	regapp := r.Group("/", createAuthVerificationHandler("REG_APP_KEY"), validateBeansQueryRequest, validateVectorSearchRequest)
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
