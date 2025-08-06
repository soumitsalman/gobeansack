package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

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
