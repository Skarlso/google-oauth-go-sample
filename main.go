package main

import (
	"log"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	token, err := RandToken(64)
	if err != nil {
		log.Fatal("unable to generate random token: ", err)
	}
	store := sessions.NewCookieStore([]byte(token))
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

	handler, err := NewHandler()
	if err != nil {
		log.Fatal("unable to load credentials: %w", err)
	}
	router.GET("/", handler.IndexHandler)
	router.GET("/login", handler.LoginHandler)
	router.GET("/auth", handler.AuthHandler)

	authorized := router.Group("/battle")
	authorized.Use(AuthorizeRequest())
	{
		authorized.GET("/field", handler.FieldHandler)
	}

	if err := router.Run("127.0.0.1:9090"); err != nil {
		log.Fatal(err)
	}
}
