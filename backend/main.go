package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"live-polling-app/config"
	"live-polling-app/handlers"
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
			"Content-Type",
		)

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Live Polling API is running!",
		})
	})

	// Added for Render health check
	r.HEAD("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	r.POST("/api/signup", handlers.Signup)

	r.POST("/api/login", handlers.Login)

	r.POST("/api/polls", handlers.CreatePoll)

	r.GET("/api/polls", handlers.GetLatestPoll)

	r.GET("/api/polls/:id", handlers.GetPoll)

	r.POST("/api/polls/:id/vote", handlers.VotePoll)

	r.GET("/api/polls/:id/stream", handlers.PollStream)

	fmt.Println("Server starting on http://localhost:8080")

	r.Run(":8080")
}