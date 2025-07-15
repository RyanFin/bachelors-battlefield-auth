package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// Ultra-permissive CORS middleware
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Always set these headers for every request
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, HEAD, PATCH")
		c.Header("Access-Control-Allow-Headers", "*")
		c.Header("Access-Control-Expose-Headers", "*")
		c.Header("Access-Control-Max-Age", "86400")
		c.Header("Access-Control-Allow-Credentials", "false")

		// Handle preflight OPTIONS requests immediately
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}

func main() {
	r := gin.Default()

	// Apply CORS middleware to ALL routes
	r.Use(corsMiddleware())

	// Additional middleware to ensure headers are set
	r.Use(func(c *gin.Context) {
		// Force set headers again as backup
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, HEAD, PATCH")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "*")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	})

	// Root endpoint
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "API is running",
			"status":  "ok",
		})
	})

	// Test endpoint
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "CORS test endpoint",
			"status":  "working",
		})
	})

	// Login endpoint
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		panic("ADMIN_PASSWORD environment variable not set")
	}

	r.OPTIONS("/admin/login", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

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

	// Catch-all OPTIONS handler for any missed preflight requests
	r.NoRoute(func(c *gin.Context) {
		if c.Request.Method == "OPTIONS" {
			c.Status(http.StatusOK)
			return
		}
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Route not found",
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r.Run(":" + port)
}
