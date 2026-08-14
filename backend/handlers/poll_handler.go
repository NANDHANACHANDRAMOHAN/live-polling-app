package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"live-polling-app/config"
	"live-polling-app/models"
)

type VoteRequest struct {
	OptionID string `json:"optionId"`
}

// CREATE POLL
func CreatePoll(c *gin.Context) {
	var poll models.Poll

	if err := c.ShouldBindJSON(&poll); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data",
		})
		return
	}

	if poll.Question == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Question is required",
		})
		return
	}

	if len(poll.Options) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "At least 2 options are required",
		})
		return
	}

	poll.ID = bson.NewObjectID().Hex()
	poll.CreatedAt = time.Now()

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	collection := config.DB.Collection("polls")

	_, err := collection.InsertOne(ctx, poll)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save poll",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Poll created successfully",
		"poll":    poll,
	})
}

// GET LATEST POLL
func GetLatestPoll(c *gin.Context) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	collection := config.DB.Collection("polls")

	var poll models.Poll

	opts := options.FindOne().SetSort(
		bson.M{"createdAt": -1},
	)

	err := collection.FindOne(
		ctx,
		bson.M{},
		opts,
	).Decode(&poll)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "No poll found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"poll": poll,
	})
}

// GET POLL BY ID
func GetPoll(c *gin.Context) {
	pollID := c.Param("id")

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	collection := config.DB.Collection("polls")

	var poll models.Poll

	err := collection.FindOne(
		ctx,
		bson.M{"_id": pollID},
	).Decode(&poll)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Poll not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"poll": poll,
	})
}

// VOTE POLL
func VotePoll(c *gin.Context) {
	var vote VoteRequest

	if err := c.ShouldBindJSON(&vote); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid vote data",
		})
		return
	}

	if vote.OptionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "optionId is required",
		})
		return
	}

	pollID := c.Param("id")

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	collection := config.DB.Collection("polls")

	filter := bson.M{
		"_id":         pollID,
		"options.id": vote.OptionID,
	}

	update := bson.M{
		"$inc": bson.M{
			"options.$.votes": 1,
		},
	}

	result, err := collection.UpdateOne(
		ctx,
		filter,
		update,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to record vote",
		})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Poll or option not found",
		})
		return
	}

	// Get updated poll
	var updatedPoll models.Poll

	err = collection.FindOne(
		ctx,
		bson.M{"_id": pollID},
	).Decode(&updatedPoll)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get updated poll",
		})
		return
	}

	// Send updated poll through Redis
	pollJSON, err := json.Marshal(updatedPoll)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to prepare poll update",
		})
		return
	}

	_, err = config.RedisClient.Publish(
		ctx,
		"poll:"+pollID,
		string(pollJSON),
	).Result()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Vote saved but Redis update failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Vote received successfully",
		"optionId": vote.OptionID,
	})
}

// LIVE POLL STREAM
func PollStream(c *gin.Context) {
	pollID := c.Param("id")

	ctx := c.Request.Context()

	pubsub := config.RedisClient.Subscribe(
		ctx,
		"poll:"+pollID,
	)

	defer pubsub.Close()

	// Make sure Redis subscription is ready
	_, err := pubsub.Receive(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Redis subscription failed",
		})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "http://localhost:5173")

	// Tell browser that live connection is ready
	c.SSEvent("connected", "true")
	c.Writer.Flush()

	messageChannel := pubsub.Channel()

	for {
		select {
		case <-ctx.Done():
			return

		case message, ok := <-messageChannel:
			if !ok {
				return
			}

			c.SSEvent("poll", message.Payload)
			c.Writer.Flush()
		}
	}
}