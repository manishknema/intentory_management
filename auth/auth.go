package auth

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/manishknema/inventory_management/config"
)

// GenerateToken creates a JWT token for a user
func GenerateJWT(username string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := jwt.MapClaims{
		"username": username,
		"exp":      expirationTime.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Ensure secret key is correctly set
	if config.SecretKey == "" {
		log.Println("❌ ERROR: JWT Secret Key is empty!")
		return "", fmt.Errorf("missing secret key")
	}

	tokenString, err := token.SignedString([]byte(config.SecretKey))
	if err != nil {
		log.Println("❌ ERROR: Unable to sign token:", err)
		return "", err
	}

	log.Println("✅ JWT Token Generated for:", username)
	return tokenString, nil
}

// JWTMiddleware validates JWT token from Authorization header
func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing token"})
			c.Abort()
			return
		}

		// Remove "Bearer " prefix if present
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")

		// Parse and validate JWT token
		log.Println("🔍 Received JWT Token:", tokenString)
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				log.Println("❌ ERROR: Unexpected JWT signing method:", token.Header["alg"])
				return nil, fmt.Errorf("unexpected signing method")
			}
			return []byte(config.SecretKey), nil
		})

		if err != nil {
			log.Println("❌ ERROR: Invalid JWT Token:", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		if !token.Valid {
			log.Println("❌ ERROR: JWT Token is invalid!")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token is invalid"})
			c.Abort()
			return
		}

		log.Println("✅ SUCCESS: Valid JWT Token")
		c.Next()
	}
}
