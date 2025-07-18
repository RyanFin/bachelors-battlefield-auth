package main

import (
	"bachelors-battlefield-auth/pkg/handlers"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB Atlas
var client *mongo.Client
var database *mongo.Database

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

func initMongoDB() {

	dbUserName := os.Getenv("DB_USERNAME")
	if dbUserName == "" {
		panic("DB_USERNAME environment variable not set")
	}

	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		panic("DB_USERNAME environment variable not set")
	}

	dbCluster := os.Getenv("DB_CLUSTER")
	if dbCluster == "" {
		panic("DB_USERNAME environment variable not set")
	}

	// MongoDB connection string
	connectionString := fmt.Sprintf("mongodb+srv://%s:%s@%s/?retryWrites=true&w=majority&appName=BattlefieldCluster", dbUserName, dbPassword, dbCluster)

	// log.Println("connection str: ", connectionString)

	// Set client options
	// Use SCRAM-SHA-256 for authentication
	clientOptions := options.Client().ApplyURI(connectionString).SetAuth(options.Credential{
		AuthMechanism: "SCRAM-SHA-256",
		Username:      dbUserName,
		Password:      dbPassword,
	})

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	// Check the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("Failed to ping MongoDB:", err)
	}

	// Set up Mongo DB database
	dbName := os.Getenv("DATABASE_NAME")
	if dbName == "" {
		panic("DATABASE_NAME environment variable not set")
	}

	database = client.Database(dbName)
	log.Println("Connected to MongoDB Atlas!")
}

func main() {
	r := gin.Default()

	// Initialize MongoDB connection
	initMongoDB()

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

	// Login endpoint
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		panic("ADMIN_PASSWORD environment variable not set")
	}

	// ----------------------
	// Routes
	// ----------------------

	// Root endpoint
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "API is running",
			"status":  "ok",
		})
	})

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

	// retrieve handlers using an instance of a handlers object
	h := &handlers.Handler{DB: database}

	// Routes
	api := r.Group("/api")
	{
		// Location routes
		api.GET("/locations", h.GetLocations)
		api.POST("/locations", h.CreateLocation)
		api.PUT("/locations/:id", h.UpdateLocation)
		api.DELETE("/locations/:id", h.DeleteLocation)
		api.GET("/locations/:id", h.GetLocation)

		// Voting routes
		api.POST("/locations/:id/vote", h.AddVote)
		api.GET("/locations/:id/votes", h.GetVotes)

		// Notes routes
		api.POST("/locations/:id/notes", h.AddNote)
		api.GET("/locations/:id/notes", h.GetNotes)
		api.DELETE("/locations/:id/notes/:noteId", h.DeleteNote)
	}

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

	// Test endpoint
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "CORS test endpoint",
			"status":  "working",
		})
	})

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// Heroku port
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r.Run(":" + port)
}
