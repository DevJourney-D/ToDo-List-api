package handlers

import (
	"database/sql"
	"net/http"
	"todo-backend/config"
	"todo-backend/middleware"
	"todo-backend/models"
	"todo-backend/utils"

	"github.com/gin-gonic/gin"
)

// Register handles user registration
func Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if username already exists
	var existingUser models.User
	err := config.DB.QueryRow("SELECT id FROM users WHERE username = $1", req.Username).Scan(&existingUser.ID)
	if err != sql.ErrNoRows {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Insert user into database
	var user models.User
	err = config.DB.QueryRow(
		"INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING id, username, created_at",
		req.Username, hashedPassword,
	).Scan(&user.ID, &user.Username, &user.CreatedAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(user.ID, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Log the registration event
	_, err = config.DB.Exec(
		"INSERT INTO logs (user_id, event_type, description) VALUES ($1, $2, $3)",
		user.ID, "user_registration", "User registered successfully",
	)
	if err != nil {
		// Log error but don't fail the request
		// log.Printf("Failed to log registration event: %v", err)
	}

	c.JSON(http.StatusCreated, models.AuthResponse{
		Token: token,
		User:  user,
	})
}

// Login handles user authentication
func Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user from database
	var user models.User
	err := config.DB.QueryRow(
		"SELECT id, username, password_hash, created_at FROM users WHERE username = $1",
		req.Username,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Check password
	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(user.ID, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Log the login event
	_, err = config.DB.Exec(
		"INSERT INTO logs (user_id, event_type, description) VALUES ($1, $2, $3)",
		user.ID, "user_login", "User logged in successfully",
	)
	if err != nil {
		// Log error but don't fail the request
		// log.Printf("Failed to log login event: %v", err)
	}

	c.JSON(http.StatusOK, models.AuthResponse{
		Token: token,
		User:  user,
	})
}

// GetUserInfo retrieves user information and stats
func GetUserInfo(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get user information
	var user models.User
	err := config.DB.QueryRow(
		"SELECT id, username, created_at FROM users WHERE id = $1",
		userID,
	).Scan(&user.ID, &user.Username, &user.CreatedAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user information"})
		return
	}

	// Get user statistics
	var stats models.UserStats

	// Get task statistics
	err = config.DB.QueryRow(
		"SELECT COUNT(*) FROM tasks WHERE user_id = $1",
		userID,
	).Scan(&stats.TotalTasks)
	if err != nil {
		stats.TotalTasks = 0
	}

	err = config.DB.QueryRow(
		"SELECT COUNT(*) FROM tasks WHERE user_id = $1 AND is_completed = true",
		userID,
	).Scan(&stats.CompletedTasks)
	if err != nil {
		stats.CompletedTasks = 0
	}

	stats.PendingTasks = stats.TotalTasks - stats.CompletedTasks

	// Get habit statistics
	err = config.DB.QueryRow(
		"SELECT COUNT(*) FROM habits WHERE user_id = $1",
		userID,
	).Scan(&stats.TotalHabits)
	if err != nil {
		stats.TotalHabits = 0
	}

	err = config.DB.QueryRow(
		"SELECT COUNT(*) FROM habits WHERE user_id = $1 AND is_achieved = true",
		userID,
	).Scan(&stats.AchievedHabits)
	if err != nil {
		stats.AchievedHabits = 0
	}

	userInfo := models.UserInfo{
		User:  user,
		Stats: stats,
	}

	c.JSON(http.StatusOK, userInfo)
}
