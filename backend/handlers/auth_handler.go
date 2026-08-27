
package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/crypto/bcrypt"

	"live-polling-app/config"
	"live-polling-app/models"
)

type AuthRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

var jwtSecret = []byte("live-polling-secret-key")

// =========================================================
// SIGNUP
// =========================================================

func Signup(c *gin.Context) {
	var req AuthRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data",
		})
		return
	}

	if req.Username == "" || req.Email == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Username, email and password are required",
		})
		return
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	collection := config.DB.Collection("users")

	// Check existing email
	var existingUser models.User

	err := collection.FindOne(
		ctx,
		bson.M{
			"email": req.Email,
		},
	).Decode(&existingUser)

	if err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Email already registered",
		})
		return
	}

	// =====================================================
	// HASH PASSWORD
	// =====================================================

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(req.Password),
		bcrypt.DefaultCost,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to secure password",
		})
		return
	}

	user := models.User{
		ID:       bson.NewObjectID().Hex(),
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	_, err = collection.InsertOne(ctx, user)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create user",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Signup successful",
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
	})
}

// =========================================================
// LOGIN
// =========================================================

func Login(c *gin.Context) {
	var req AuthRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data",
		})
		return
	}

	if req.Email == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Email and password are required",
		})
		return
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	collection := config.DB.Collection("users")

	var user models.User

	// Find user only by email
	err := collection.FindOne(
		ctx,
		bson.M{
			"email": req.Email,
		},
	).Decode(&user)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid email or password",
		})
		return
	}

	// =====================================================
	// COMPARE PASSWORD WITH HASH
	// =====================================================

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(req.Password),
	)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid email or password",
		})
		return
	}

	// =====================================================
	// CREATE JWT TOKEN
	// =====================================================

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"user_id":  user.ID,
			"email":    user.Email,
			"username": user.Username,
			"exp":      time.Now().Add(24 * time.Hour).Unix(),
		},
	)

	tokenString, err := token.SignedString(jwtSecret)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate authentication token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   tokenString,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
	})
}
