package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/k0kubun/pp"
)

const (
	DEFAULT_LIMIT int64 = 50
	MIN_LIMIT     int64 = 1
	MAX_LIMIT     int64 = 300
)

type BeansQueryRequest struct {
	URLs         []string     `json:"urls,omitempty"`
	Kind         string       `form:"kind"`
	Authors      []string     `form:"authors"`
	Sources      []string     `form:"sources"`
	CreatedSince time.Time    `form:"created_since" time_format:"2006-01-02T15:04:05Z07:00"`
	UpdatedSince time.Time    `form:"updated_since" time_format:"2006-01-02T15:04:05Z07:00"`
	Categories   []string     `form:"categories"`
	Regions      []string     `form:"regions"`
	Entities     []string     `form:"entities"`
	Embedding    Float32Array `json:"embedding,omitempty"`
	MaxDistance  float64      `json:"max_distance,omitempty" binding:"min=0,max=1"`
	Offset       int64        `form:"offset" json:"offset" binding:"min=0"`
	Limit        int64        `form:"limit" json:"limit" binding:"min=0,max=300" default:"50"`
}

func validateBeansQueryRequest(c *gin.Context) {
	var req BeansQueryRequest
	err := c.ShouldBindQuery(&req)
	pp.Println("err", err)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err = c.ShouldBindJSON(&req)
	pp.Println("err", err)
	// not having a body is not an error, malformed json is
	if err != nil && err.Error() != "EOF" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(req.Embedding) > 0 && req.Limit == 0 && req.MaxDistance == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "For vector search, you must provide either a limit or a max_distance"})
		return
	}
	// adjust the default
	if req.Limit == 0 {
		req.Limit = DEFAULT_LIMIT
	}
	c.Set("req", req)
	c.Next()
}

func createRelatedBeansHandler(ds *Beansack) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := c.MustGet("req").(BeansQueryRequest)
		if len(req.URLs) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "urls is required"})
			return
		}
		c.JSON(http.StatusOK, ds.GetRelated(req.URLs))
	}
}

func createRegionsHandler(ds *Beansack) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := c.MustGet("req").(BeansQueryRequest)
		if len(req.URLs) > 0 {
			c.JSON(http.StatusOK, ds.GetRegions(req.URLs))
		} else {
			c.JSON(http.StatusOK, ds.DistinctRegions())
		}
	}
}

func createEntitiesHandler(ds *Beansack) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := c.MustGet("req").(BeansQueryRequest)
		if len(req.URLs) > 0 {
			c.JSON(http.StatusOK, ds.GetEntities(req.URLs))
		} else {
			c.JSON(http.StatusOK, ds.DistinctEntities())
		}
	}
}

func createCategoriesHandler(ds *Beansack) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := c.MustGet("req").(BeansQueryRequest)
		if len(req.URLs) > 0 {
			c.JSON(http.StatusOK, ds.GetCategories(req.URLs))
		} else {
			c.JSON(http.StatusOK, ds.DistinctCategories())
		}
	}
}

func createSourcesHandler(ds *Beansack) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := c.MustGet("req").(BeansQueryRequest)
		if len(req.Sources) > 0 {
			c.JSON(http.StatusOK, ds.GetSources(req.Sources))
		} else {
			c.JSON(http.StatusOK, ds.DistinctSources())
		}
	}
}

func createLatestBeansHandler(db *Beansack) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := c.MustGet("req").(BeansQueryRequest)
		beans, err := findBeans(db, "latest", req, PUBLIC_COLUMNS)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, beans)
	}
}

func createTrendingBeansHandler(db *Beansack) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := c.MustGet("req").(BeansQueryRequest)
		beans, err := findBeans(db, "trending", req, PUBLIC_COLUMNS)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, beans)
	}
}

func createLatestDigestsHandler(db *Beansack) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := c.MustGet("req").(BeansQueryRequest)
		// pp.Println("req", req)
		beans, err := findBeans(db, "latest", req, DIGEST_COLUMNS)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, beans)
	}
}

func createTrendingDigestsHandler(db *Beansack) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := c.MustGet("req").(BeansQueryRequest)
		// pp.Println("req", req)
		beans, err := findBeans(db, "trending", req, DIGEST_COLUMNS)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, beans)
	}
}

func findBeans(db *Beansack, trending_or_latest string, req BeansQueryRequest, columns string) ([]Bean, error) {
	if trending_or_latest == "trending" {
		return db.QueryTrendingBeans(
			req.URLs,
			[]string{req.Kind}, // only one kind
			req.Authors,
			req.Sources,
			req.CreatedSince,
			time.Time{}, // no collected
			req.UpdatedSince,
			req.Categories,
			req.Regions,
			req.Entities,
			req.Embedding,
			req.MaxDistance,
			nil,
			nil,
			req.Limit,
			req.Offset,
			[]string{columns},
		)
	} else {
		return db.QueryLatestBeans(
			req.URLs,
			[]string{req.Kind}, // only one kind
			req.Authors,
			req.Sources,
			req.CreatedSince,
			time.Time{}, // no collected
			req.Categories,
			req.Regions,
			req.Entities,
			req.Embedding,
			req.MaxDistance,
			nil,
			nil,
			req.Limit,
			req.Offset,
			[]string{columns},
		)
	}
}
