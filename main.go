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

	token, err := RandToken(64)
	if err != nil {
		logger.Error("unable to generate random token", "error", err)
		os.Exit(1)
	}

	store := cookie.NewStore([]byte(token))
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
