package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/manishknema/inventory_management/config"
	"github.com/manishknema/inventory_management/database"
	"github.com/manishknema/inventory_management/routes"
	"github.com/ulule/limiter/v3"
	ginlimiter "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

func main() {
	// Load Config and Initialize Database
	config.LoadConfig()
	database.InitDB()

	// Setup Router
	r := routes.SetupRouter()

	// CORS Middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}
		c.Next()
	})

	// Rate Limiting Middleware: 10 requests per minute per IP
	rate := limiter.Rate{Period: 1 * time.Minute, Limit: 10}
	store := memory.NewStore()
	middleware := ginlimiter.NewMiddleware(limiter.New(store, rate))
	r.Use(middleware)

	// Start the server
	log.Println("🚀 Server running on http://localhost:8080")
	r.Run(":8080")
}
