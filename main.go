package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(sessionErrorHandler())

	// Use a fixed key for sessions to prevent invalidation on restart
	// In production, this should come from environment variables or a secure key store
	sessionKey := "your-secret-key-here-change-in-production-32-bytes-long"
	if len(sessionKey) < 32 {
		logger.Error("session key must be at least 32 bytes long")
		os.Exit(1)
	}

	store := cookie.NewStore([]byte(sessionKey))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		Secure:   false, // Set to true in production with HTTPS
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	router.Use(sessions.Sessions("goquestsession", store))
	router.Static("/css", "./static/css")
	router.Static("/img", "./static/img")
	router.LoadHTMLGlob("templates/*")

	db := &Database{}
	handler, err := NewHandler(db)
	if err != nil {
		logger.Error("unable to load credentials", "error", err)
		os.Exit(1)
	}

	router.GET("/", handler.IndexHandler)
	router.GET("/login", handler.LoginHandler)
	router.GET("/auth", handler.AuthHandler)
	router.GET("/logout", handler.LogoutHandler)

	authorized := router.Group("/battle")
	authorized.Use(AuthorizeRequest())
	{
		authorized.GET("/field", handler.FieldHandler)
	}

	srv := &http.Server{
		Addr:         "127.0.0.1:9090",
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("Starting server", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("Server exiting")
}

// sessionErrorHandler handles session-related errors gracefully
func sessionErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				if err, ok := r.(error); ok {
					slog.Default().Warn("Session error recovered", "error", err, "path", c.Request.URL.Path)
					// Clear potentially corrupted session and redirect to home
					if session := sessions.Default(c); session != nil {
						session.Clear()
						session.Save()
					}
				}
			}
		}()
		c.Next()
	}
}
