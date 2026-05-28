package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jusu2/hello-world/database"
)

func main() {
	// 初始化数据库链接
	if err := database.Init(); err != nil {
		panic(err)
	}
	defer database.Close()

	r := gin.Default()

	// 然后简单的传一个字符串 然后存到pgsql里面
	r.PUT("/messages", func(c *gin.Context) {
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