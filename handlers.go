package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Credentials which stores google ids.
type Credentials struct {
	Installed struct {
		Cid     string `json:"client_id"`
		Csecret string `json:"client_secret"`
	} `json:"installed"`
}

// Handlers contains the handlers and config for the handlers.
type Handlers struct {
	conf *oauth2.Config
	db   *Database
}

func NewHandler(db *Database) (*Handlers, error) {
	cred := &Credentials{}
	file, err := os.ReadFile("./creds.json")
	if err != nil {
		log.Printf("File error: %v\n", err)
		return nil, err
	}
	if err := json.Unmarshal(file, &cred); err != nil {
		log.Println("unable to marshal data")
		return nil, err
	}

	conf := &oauth2.Config{
		ClientID:     cred.Installed.Cid,
		ClientSecret: cred.Installed.Csecret,
		RedirectURL:  "http://127.0.0.1:9090/auth", // can come from Credentials file redirect URLs.
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email", // You have to select your own scope from here -> https://developers.google.com/identity/protocols/googlescopes#google_sign-in
		},
		Endpoint: google.Endpoint,
	}

	return &Handlers{
		conf: conf,
		db:   db,
	}, nil
}

// RandToken generates a random @l length token.
func RandToken(l int) (string, error) {
	b := make([]byte, l)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

func (h *Handlers) getLoginURL(state string) string {
	return h.conf.AuthCodeURL(state)
}

// IndexHandler handles the location /.
func (h *Handlers) IndexHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "index.tmpl", gin.H{})
}

// AuthHandler handles authentication of a user and initiates a session.
func (h *Handlers) AuthHandler(c *gin.Context) {
	// Handle the exchange code to initiate transport.
	session := sessions.Default(c)
	retrievedState := session.Get("state")
	queryState := c.Request.URL.Query().Get("state")
	if retrievedState != nil && retrievedState != queryState {
		log.Printf("Invalid session state: retrieved: %s; Param: %s", retrievedState, queryState)
		c.HTML(http.StatusUnauthorized, "error.tmpl", gin.H{"message": "Invalid session state."})
		return
	}

	code := c.Request.URL.Query().Get("code")
	ctx, done := context.WithTimeout(context.Background(), 15*time.Second)
	defer done()

	tok, err := h.conf.Exchange(ctx, code)
	if err != nil {
		log.Println(err)
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"message": "Login failed. Please try again."})
		return
	}

	client := h.conf.Client(ctx, tok)
	userinfo, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	defer userinfo.Body.Close()

	data, err := io.ReadAll(userinfo.Body)
	if err != nil {
		log.Println(err)
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"message": "Failed to read request body."})
		return
	}

	u := User{}
	if err = json.Unmarshal(data, &u); err != nil {
		log.Println(err)
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"message": "Error marshalling response. Please try again."})
		return
	}

	session.Set("user-id", u.Email)
	err = session.Save()
	if err != nil {
		log.Println(err)
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"message": "Error while saving session. Please try again."})
		return
	}

	if _, err := h.db.LoadUser(u.Email); err == nil {
		c.HTML(http.StatusOK, "battle.tmpl", gin.H{"email": u.Email, "seen": true})
		return
	}

	err = h.db.SaveUser(&u)
	if err != nil {
		log.Println(err)
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"message": "Error while saving user. Please try again."})
		return
	}

	c.HTML(http.StatusOK, "battle.tmpl", gin.H{"email": u.Email, "seen": false})
}

// LoginHandler handles the login procedure.
func (h *Handlers) LoginHandler(c *gin.Context) {
	state, err := RandToken(32)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"message": "Error while generating random data."})
		return
	}

	session := sessions.Default(c)
	session.Set("state", state)
	err = session.Save()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"message": "Error while saving session."})
		return
	}
	link := h.getLoginURL(state)

	c.HTML(http.StatusOK, "auth.tmpl", gin.H{"link": link})
}

// FieldHandler is a rudimentary handler for logged in users.
func (h *Handlers) FieldHandler(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user-id")
	c.HTML(http.StatusOK, "field.tmpl", gin.H{"user": userID})
}
