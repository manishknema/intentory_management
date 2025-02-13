package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/manishknema/inventory_management/database"
	"github.com/manishknema/inventory_management/models"
	"golang.org/x/crypto/bcrypt"
)

// Secret key for JWT
var jwtSecretKey = []byte("your_secret_key") // 🔹 Change this in production

// Signup registers a new user
func Signup(c *gin.Context) {
	log.Println("📥 Received request to Signup")

	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		log.Println("❌ Error parsing request JSON:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Hash the password before storing it
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("❌ Error hashing password:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not hash password"})
		return
	}

	log.Println("🔍 Storing user:", user.Username)
	_, err = database.DB.Exec("INSERT INTO users (username, password) VALUES (?, ?)", user.Username, hashedPassword)
	if err != nil {
		log.Println("❌ SQL Insert Error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create user"})
		return
	}

	log.Println("✅ User registered successfully:", user.Username)
	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

// Login authenticates a user and returns a JWT token
func Login(c *gin.Context) {
	log.Println("📥 Received request to Login")

	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		log.Println("❌ Error parsing request JSON:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	var storedPassword string
	err := database.DB.QueryRow("SELECT password FROM users WHERE username = ?", user.Username).Scan(&storedPassword)
	if err == sql.ErrNoRows {
		log.Println("❌ User not found:", user.Username)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	} else if err != nil {
		log.Println("❌ SQL Query Error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Verify the password
	if err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(user.Password)); err != nil {
		log.Println("❌ Invalid password for user:", user.Username)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// Generate JWT token
	expirationTime := time.Now().Add(1 * time.Hour) // Token valid for 1 hour
	claims := &models.Claims{
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecretKey)
	if err != nil {
		log.Println("❌ Error generating JWT:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}

	log.Println("✅ User logged in successfully:", user.Username)
	c.JSON(http.StatusOK, gin.H{"token": tokenString})
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
