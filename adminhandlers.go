package main

// import (
// 	"net/http"

// 	"github.com/gin-gonic/gin"
// )

// type AdminCommandRequest struct {
// 	Commands []string `json:"commands"`
// }

// func createAdminCommandHandler(ds *Ducksack) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		var req AdminCommandRequest
// 		if err := c.ShouldBindJSON(&req); err != nil {
// 			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 			return
// 		}
// 		err := ds.Execute(req.Commands...)
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 			return
// 		}
// 		c.JSON(http.StatusOK, gin.H{"status": "success"})
// 	}
// }
