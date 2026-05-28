package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jusu2/hello-world/database"
)

func main() {
	if err := database.Init(); err != nil {
		panic(err)
	}
	defer database.Close()

	r := gin.Default()

	// 静态文件：serve frontend 目录
	r.Static("/", "../frontend")

	// API
	r.PUT("/api/messages", func(c *gin.Context) {
		var req struct {
			Content string `json:"content" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "pls provide content field"})
			return
		}

		if err := database.SaveMessage(req.Content); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "write failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "save successful", "content": req.Content})
	})

	_ = r.Run(":8080")
}