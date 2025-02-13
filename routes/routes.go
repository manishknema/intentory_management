package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/manishknema/inventory_management/auth"
	"github.com/manishknema/inventory_management/handlers"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()
	r.LoadHTMLGlob("templates/*.html")

	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", nil)
	})

	r.GET("/login", func(c *gin.Context) {
		c.HTML(200, "login.html", nil)
	})
	r.GET("/signup", func(c *gin.Context) {
		c.HTML(200, "signup.html", nil)
	})

	r.POST("/signup", handlers.Signup)
	r.POST("/login", handlers.Login)

	r.GET("/check-users", handlers.CheckUsers)

	// Protected Routes (Require JWT)
	authorized := r.Group("/")
	authorized.Use(auth.JWTMiddleware())
	authorized.GET("/items", handlers.GetItems)
	authorized.GET("/items/:id", handlers.GetItem)
	authorized.POST("/items", handlers.CreateItem)
	authorized.PUT("/items/:id", handlers.UpdateItem)
	authorized.DELETE("/items/:id", handlers.DeleteItem)            // Correct individual item delete
	authorized.POST("/items/delete-multiple", handlers.DeleteItems) // Correct multiple item deletion

	return r
}
