package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	DEFAULT_LIMIT int64 = 50
	MIN_LIMIT     int64 = 1
	MAX_LIMIT     int64 = 200
)

var SELECT_PUBLIC_FIELDS = []string{"url", "kind", "title", "summary", "author", "source", "created", "categories", "sentiments", "regions", "entities", "updated", "likes", "comments", "shares"}
var SELECT_GISTS = []string{"url", "created", "updated", "gist"}
var SELECT_EMBEDDINGS = []string{"url", "created", "updated", "embedding"}
var ORDER_BY_CREATED = []string{"created DESC"}
var ORDER_BY_DISTANCE = []string{"distance ASC"}
var ORDER_BY_CHATTERS = []string{"updated DESC", "comments DESC", "likes DESC", "shares DESC"}

type BeanSearchRequest struct {
	// these are query params
	Kind       string    `form:"kind"`
	Since      time.Time `form:"created_since" time_format:"2006-01-02T15:04:05Z07:00"`
	Categories []string  `form:"categories"`
	Regions    []string  `form:"regions"`
	Entities   []string  `form:"entities"`
	Domains    []string  `form:"domains"`
	Offset     int64     `form:"offset" binding:"min=0"`
	Limit      int64     `form:"limit" binding:"min=0,max=200" default:"50"`
	// these are body params
	Embedding   Float32Array `json:"embedding,omitempty"`
	MaxDistance float64      `json:"max_distance,omitempty" binding:"min=0,max=1" default:"0.5"`
	URLs        []string     `json:"urls,omitempty"`
}

func createBeansSearchHandler(ds *Ducksack) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := c.MustGet("req").(BeanSearchRequest)
		order_by := ORDER_BY_CREATED
		if len(req.Embedding) > 0 {
			order_by = ORDER_BY_DISTANCE
		}
		c.JSON(http.StatusOK, findBeans(ds, req, order_by, SELECT_PUBLIC_FIELDS))
	}
}

func createRelatedBeansHandler(ds *Ducksack) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := c.MustGet("req").(BeanSearchRequest)
		if len(req.URLs) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "urls is required"})
			return
		}
		c.JSON(http.StatusOK, ds.GetRelated(req.URLs))
	}
}

func createRegionsHandler(ds *Ducksack) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := c.MustGet("req").(BeanSearchRequest)
		if len(req.URLs) > 0 {
			c.JSON(http.StatusOK, ds.GetRegions(req.URLs))
		} else {
			c.JSON(http.StatusOK, ds.DistinctRegions())
		}
	}
}

func createEntitiesHandler(ds *Ducksack) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := c.MustGet("req").(BeanSearchRequest)
		if len(req.URLs) > 0 {
			c.JSON(http.StatusOK, ds.GetEntities(req.URLs))
		} else {
			c.JSON(http.StatusOK, ds.DistinctEntities())
		}
	}
}

func createCategoriesHandler(ds *Ducksack) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := c.MustGet("req").(BeanSearchRequest)
		if len(req.URLs) > 0 {
			c.JSON(http.StatusOK, ds.GetCategories(req.URLs))
		} else {
			c.JSON(http.StatusOK, ds.DistinctCategories())
		}
	}
}

func createSourcesHandler(ds *Ducksack) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := c.MustGet("req").(BeanSearchRequest)
		if len(req.Domains) > 0 {
			c.JSON(http.StatusOK, ds.GetSources(req.Domains))
		} else {
			c.JSON(http.StatusOK, ds.DistinctSources())
		}
	}
}

func createExistsHandler(ds *Ducksack) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := c.MustGet("req").(BeanSearchRequest)
		c.JSON(http.StatusOK, ds.Exists(req.URLs))
	}
}

func createTrendingBeansHandler(ds *Ducksack) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := c.MustGet("req").(BeanSearchRequest)
		if len(req.Embedding) > 0 && req.MaxDistance == 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "For vector search, you must provide a max_distance"})
			return
		}
		c.JSON(http.StatusOK, findBeans(ds, req, ORDER_BY_CHATTERS, nil))
	}
}

func createTrendingBeanGistsHandler(ds *Ducksack) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := c.MustGet("req").(BeanSearchRequest)
		if len(req.Embedding) > 0 && req.MaxDistance == 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "For vector search, you must provide a max_distance"})
			return
		}
		c.JSON(http.StatusOK, findBeans(ds, req, ORDER_BY_CHATTERS, SELECT_GISTS))
	}
}

func createTrendingBeanEmbeddingsHandler(ds *Ducksack) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := c.MustGet("req").(BeanSearchRequest)
		c.JSON(http.StatusOK, findBeans(ds, req, ORDER_BY_CHATTERS, SELECT_EMBEDDINGS))
	}
}

func findBeans(ds *Ducksack, req BeanSearchRequest, order_by []string, fields []string) []Bean {
	var beans []Bean
	if len(req.Embedding) > 0 {
		beans = ds.VectorSearchBeans(
			req.Embedding,
			req.MaxDistance,
			req.Kind,
			req.Since,
			req.Categories,
			req.Regions,
			req.Entities,
			req.Domains,
			order_by,
			req.Offset,
			req.Limit,
			fields,
		)
	} else {
		beans = ds.QueryBeans(
			req.Kind,
			req.Since,
			req.Categories,
			req.Regions,
			req.Entities,
			req.Domains,
			order_by,
			req.Offset,
			req.Limit,
			fields,
		)
	}
	fmt.Println("beans", len(beans))
	return beans
}
