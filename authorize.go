package main

import (
	"log/slog"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// AuthorizeRequest is used to authorize a request for a certain end-point group.
func AuthorizeRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := slog.Default()
		session := sessions.Default(c)
		
		// Handle potential session errors gracefully
		userID := session.Get("user-id")
		if userID == nil {
			// Clear any corrupted session data
			session.Clear()
			session.Save()
			
			logger.Warn("Unauthorized access attempt", "path", c.Request.URL.Path, "ip", c.ClientIP())
			c.HTML(http.StatusUnauthorized, "error.tmpl", gin.H{"message": "Your session has expired. Please login again."})
			c.Abort()
			return
		}
		
		logger.Info("Authorized access", "user", userID, "path", c.Request.URL.Path)
		c.Next()
	}
}
