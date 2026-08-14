package routes

import (
	"github.com/gin-gonic/gin"
	"live-polling-app/handlers"
)

func PollRoutes(r *gin.Engine) {
	r.POST("/api/polls", handlers.CreatePoll)
	r.POST("/api/polls/:id/vote", handlers.VotePoll)
}