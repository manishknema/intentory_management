package models

import "github.com/golang-jwt/jwt/v5"

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// Claims struct for JWT token
type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}
