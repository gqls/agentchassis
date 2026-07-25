package httperr

import "github.com/gin-gonic/gin"

// JSONError writes a uniform JSON error response used by all middleware and handlers.
func JSONError(c *gin.Context, status int, msg string) {
	c.JSON(status, gin.H{"error": msg})
}
