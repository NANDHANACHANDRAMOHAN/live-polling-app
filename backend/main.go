
package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"live-polling-app/config"
	"live-polling-app/handlers"
	"live-polling-app/middleware"
)

func main() {

	// Load .env file
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Warning: .env file not found")
	}

	// Connect to MongoDB
	err = config.ConnectDatabase()
	if err != nil {
		panic(err)
	}

	// Connect to Redis
	err = config.ConnectRedis()
	if err != nil {
		panic(err)
	}

	r := gin.Default()

	// CORS
	r.Use(func(c *gin.Context) {

		c.Writer.Header().Set(
			"Access-Control-Allow-Origin",
			"http://localhost:5173",
		)

		c.Writer.Header().Set(
			"Access-Control-Allow-Methods",
			"GET, POST, OPTIONS",
		)

		c.Writer.Header().Set(
			"Access-Control-Allow-Headers",
			"Content-Type, Authorization",
		)

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// Root
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Live Polling API is running!",
		})
	})

	// =========================
	// PUBLIC AUTH ROUTES
	// =========================

	r.POST("/api/signup", handlers.Signup)
	r.POST("/api/login", handlers.Login)

	// =========================
	// PUBLIC POLL ROUTES
	// =========================

	r.GET("/api/polls", handlers.GetLatestPoll)
	r.GET("/api/polls/:id", handlers.GetPoll)

	r.GET("/api/polls/:id/stream", handlers.PollStream)

	// =========================
	// PROTECTED POLL ROUTES
	// =========================

	protected := r.Group("/api")
	protected.Use(middleware.JWTAuth())

	protected.POST("/polls", handlers.CreatePoll)
	protected.POST("/polls/:id/vote", handlers.VotePoll)
	

	fmt.Println("Server starting on http://localhost:8080")

	r.Run(":8080")
}

