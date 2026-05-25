package middleware

import (
	"github.com/gin-gonic/gin"
)

// AuthMiddleware is a middleware that checks if the user is authenticated
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Here can be some logic to check if the user is authenticated and add the "user_id" to the context
		c.Next()
	}
}
