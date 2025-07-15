package main

import (
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// ✅ Updated CORS configuration to allow all origins
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD", "PATCH",
		},
		AllowHeaders: []string{
			"Origin", "Content-Type", "Accept", "Authorization",
			"X-Requested-With", "Access-Control-Request-Method",
			"Access-Control-Request-Headers",
		},
		ExposeHeaders: []string{
			"Content-Length", "Access-Control-Allow-Origin",
			"Access-Control-Allow-Headers", "Content-Type",
		},
		AllowCredentials: false, // Set to false when using "*" for AllowOrigins
		MaxAge:           86400, // 24 hours
	}))

	// ✅ Additional manual CORS headers as fallback
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// ✅ Handle preflight OPTIONS requests for all paths
	r.OPTIONS("/*path", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	// ✅ Actual POST login route
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		panic("ADMIN_PASSWORD environment variable not set")
	}

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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
