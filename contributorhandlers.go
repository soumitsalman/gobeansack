package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type QueryByMissingTagsRequest struct {
	Missing []string `json:"missing" binding:"required"`
	Where   []string `json:"where"`
	OrderBy []string `json:"order_by"`
	Limit   int64    `json:"limit" binding:"min=0,max=1024"`
}

func createQueryBeansWithMissingTagsHandler(ds *Ducksack) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req QueryByMissingTagsRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if len(req.Missing) == 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "'missing' field is required"})
			return
		}
		if req.Limit == 0 {
			req.Limit = DEFAULT_LIMIT
		}
		where := CreateWhereExprsForMissingTags(req.Missing)
		if len(req.Where) > 0 {
			where = append(where, req.Where...)
		}
		beans, err := ds.QueryBeanCores(where, req.OrderBy, 0, req.Limit)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, beans)
	}
}

// //////// DELETE HANDLERS //////////
type DeleteRequests struct {
	Where []string `json:"where" binding:"required"`
}

func validateDeleteRequest(c *gin.Context) {
	var req DeleteRequests
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(req.Where) == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "'where' field is required"})
		return
	}
	c.Set("req", req)
	c.Next()
}

func createDeleteBeansHandler(ds *Ducksack) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := c.MustGet("req").(DeleteRequests)
		err := ds.DeleteBeans(req.Where...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	}
}

func createDeleteChattersHandler(ds *Ducksack) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := c.MustGet("req").(DeleteRequests)
		err := ds.DeleteChatters(req.Where...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	}
}

func createDeleteSourcesHandler(ds *Ducksack) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := c.MustGet("req").(DeleteRequests)
		err := ds.DeleteSources(req.Where...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	}
}

////////// STORE HANDLERS //////////

func createStoreBeansHandler(ds *Ducksack) gin.HandlerFunc {
	return func(c *gin.Context) {
		var beans []Bean
		if err := c.ShouldBindJSON(&beans); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		go ds.StoreBeans(beans)
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	}
}

func createStoreEmbeddingsHandler(ds *Ducksack) gin.HandlerFunc {
	return func(c *gin.Context) {
		var embeddings []Bean
		if err := c.ShouldBindJSON(&embeddings); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		go ds.StoreEmbeddings(embeddings)
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	}
}

func createStoreTagsHandler(ds *Ducksack) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tags []Bean
		if err := c.ShouldBindJSON(&tags); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		go ds.StoreTags(tags)
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	}
}

func createStoreChatterHandler(ds *Ducksack) gin.HandlerFunc {
	return func(c *gin.Context) {
		var chatters []Chatter
		if err := c.ShouldBindJSON(&chatters); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		go ds.StoreChatters(chatters)
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	}
}

func createStoreSourceHandler(ds *Ducksack) gin.HandlerFunc {
	return func(c *gin.Context) {
		var sources []Source
		if err := c.ShouldBindJSON(&sources); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		go ds.StoreSources(sources)
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	}
}
