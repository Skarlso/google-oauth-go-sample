package main

import (
	"github.com/Skarlso/google-oauth-go-sample/handlers"
	"github.com/Skarlso/google-oauth-go-sample/middleware"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	store := sessions.NewCookieStore([]byte(handlers.RandToken(64)))
	store.Options(sessions.Options{
		Path:   "/",
		MaxAge: 86400 * 7,
	})
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(sessions.Sessions("goquestsession", store))
	router.Static("/css", "./static/css")
	router.Static("/img", "./static/img")
	router.LoadHTMLGlob("templates/*")

	router.GET("/", handlers.IndexHandler)
	router.GET("/login", handlers.LoginHandler)
	router.GET("/auth", handlers.AuthHandler)

	authorized := router.Group("/battle")
	authorized.Use(middleware.AuthorizeRequest())
	{
		authorized.GET("/field", handlers.FieldHandler)
	}

	router.Run("127.0.0.1:9090")
}
