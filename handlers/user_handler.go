package handlers

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/manishknema/inventory_management/auth"
	"github.com/manishknema/inventory_management/database"
	"github.com/manishknema/inventory_management/models"
	"golang.org/x/crypto/bcrypt"
)

// Secret key for JWT
var jwtSecretKey = []byte("your_secret_key") // 🔹 Change this in production

// Signup registers a new user
func Signup(c *gin.Context) {
	var user models.User

	if err := c.ShouldBindJSON(&user); err != nil {
		log.Println("❌ ERROR: Invalid signup request:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Hash the password before storing
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("❌ ERROR: Password hashing failed:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not process request"})
		return
	}

	// Insert user into database
	_, err = database.DB.Exec("INSERT INTO users (username, password) VALUES (?, ?)", user.Username, string(hashedPassword))
	if err != nil {
		log.Println("❌ ERROR: Database insertion failed:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Signup failed"})
		return
	}

	// ✅ Call GenerateJWT() after successful signup
	token, err := auth.GenerateJWT(user.Username)
	if err != nil {
		log.Println("❌ ERROR: JWT generation failed:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Signup successful but token generation failed"})
		return
	}

	log.Println("✅ User registered successfully:", user.Username)
	c.JSON(http.StatusCreated, gin.H{"message": "Signup successful", "token": token})
}

// Login authenticates a user and returns a JWT token
func Login(c *gin.Context) {
	var loginRequest models.User
	var storedUser models.User

	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		log.Println("❌ ERROR: Invalid login request:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Fetch user from database
	err := database.DB.QueryRow("SELECT username, password FROM users WHERE username = ?", loginRequest.Username).
		Scan(&storedUser.Username, &storedUser.Password)

	if err == sql.ErrNoRows {
		log.Println("❌ ERROR: User not found:", loginRequest.Username)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	} else if err != nil {
		log.Println("❌ ERROR: Database query failed:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not process request"})
		return
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(loginRequest.Password))
	if err != nil {
		log.Println("❌ ERROR: Password mismatch for user:", loginRequest.Username)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// ✅ Call GenerateJWT() after successful login
	token, err := auth.GenerateJWT(storedUser.Username)
	if err != nil {
		log.Println("❌ ERROR: JWT generation failed:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Login successful but token generation failed"})
		return
	}

	log.Println("✅ User logged in successfully:", storedUser.Username)
	c.JSON(http.StatusOK, gin.H{"message": "Login successful", "token": token})
}

// CheckUsers returns the number of registered users
func CheckUsers(c *gin.Context) {
	log.Println("📥 Checking if users exist")

	var userCount int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	if err != nil {
		log.Println("❌ SQL Query Error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	log.Println("✅ Total users found:", userCount)
	c.JSON(http.StatusOK, gin.H{"users": userCount})
}

// Delete a single item by ID
func DeleteItem(c *gin.Context) {
	id := c.Param("id")
	log.Println("📥 Received request to delete item ID:", id)

	_, err := database.DB.Exec("DELETE FROM inventory WHERE id = ?", id)
	if err != nil {
		log.Println("❌ SQL Delete Error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not delete item"})
		return
	}

	log.Println("✅ Item deleted successfully:", id)
	c.JSON(http.StatusOK, gin.H{"message": "Item deleted successfully"})
}
