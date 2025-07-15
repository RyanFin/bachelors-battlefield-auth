package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Get the password from environment variable
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		panic("ADMIN_PASSWORD environment variable not set")
	}

	// POST /admin/login
	r.POST("/admin/login", func(c *gin.Context) {
		var req struct {
			Password string `json:"password" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Password is required",
			})
			return
		}

		if req.Password == adminPassword {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Authenticated",
			})
			return
		}

		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Invalid password",
		})
	})

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
