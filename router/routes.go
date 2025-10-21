package router

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	bs "github.com/soumitsalman/gobeansack/beansack"
)

var query_throttler chan int

func checkHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

func createAuthVerificationHandler(expectedKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-KEY")
		expectedKey := os.Getenv(expectedKey)

		if expectedKey == "" || apiKey == expectedKey {
			c.Next()
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
		}
	}
}

func enterThrottler(c *gin.Context) {
	query_throttler <- 1
	c.Next()
}

func exitThrottler(c *gin.Context) {
	<-query_throttler
	c.Next()
}

func InitializeRoutes(db *bs.Beansack, max_concurrent_requests int) *gin.Engine {
	r := gin.Default()

	query_throttler = make(chan int, max_concurrent_requests)

	// Health check endpoint - no auth required
	r.GET("/health", checkHealth)

	// NEWS API ENDPOINTS
	articles := r.Group("/articles", createAuthVerificationHandler("API_KEY"), validateBeansQueryRequest)
	{
		articles.GET("/latest", enterThrottler, createLatestBeansHandler(db), exitThrottler)
		articles.GET("/trending", enterThrottler, createTrendingBeansHandler(db), exitThrottler)
		articles.GET("/related", createRelatedBeansHandler(db))
		articles.GET("/regions", createRegionsHandler(db))
		articles.GET("/entities", createEntitiesHandler(db))
		articles.GET("/categories", createCategoriesHandler(db))
		articles.GET("/publishers", createSourcesHandler(db))
	}

	// REGISTERED APPLICATION ENDPOINTS
	regapp := r.Group("/digests", createAuthVerificationHandler("API_KEY"), validateBeansQueryRequest)
	{
		regapp.GET("/latest", enterThrottler, createLatestDigestsHandler(db), exitThrottler)
		regapp.GET("/trending", enterThrottler, createTrendingDigestsHandler(db), exitThrottler)
	}

	return r
}
