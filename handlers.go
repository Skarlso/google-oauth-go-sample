package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/sessions"
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
		return nil, fmt.Errorf("failed to read credentials file: %w", err)
	}

	if err := json.Unmarshal(file, &cred); err != nil {
		return nil, fmt.Errorf("failed to unmarshal credentials: %w", err)
	}

	conf := &oauth2.Config{
		ClientID:     cred.Installed.Cid,
		ClientSecret: cred.Installed.Csecret,
		RedirectURL:  "http://127.0.0.1:9090/auth",
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email", // You have to select your own scope from here -> https://developers.google.com/identity/protocols/googlescopes#google_sign-in
			"https://www.googleapis.com/auth/userinfo.profile",
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
	logger := slog.Default()
	session := sessions.Default(c)
	retrievedState := session.Get("state")
	queryState := c.Request.URL.Query().Get("state")
	
	if retrievedState == nil || retrievedState != queryState {
		logger.Warn("Invalid session state", "retrieved", retrievedState, "query", queryState)
		c.HTML(http.StatusUnauthorized, "error.tmpl", gin.H{"message": "Invalid session state."})
		return
	}

	code := c.Request.URL.Query().Get("code")
	if code == "" {
		logger.Warn("Missing authorization code")
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"message": "Missing authorization code."})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	tok, err := h.conf.Exchange(ctx, code)
	if err != nil {
		logger.Error("Failed to exchange authorization code", "error", err)
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"message": "Login failed. Please try again."})
		return
	}

	client := h.conf.Client(ctx, tok)
	userinfo, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		logger.Error("Failed to get user info", "error", err)
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"message": "Failed to get user information."})
		return
	}
	defer userinfo.Body.Close()

	if userinfo.StatusCode != http.StatusOK {
		logger.Error("User info request failed", "status", userinfo.StatusCode)
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"message": "Failed to get user information."})
		return
	}

	data, err := io.ReadAll(userinfo.Body)
	if err != nil {
		logger.Error("Failed to read user info response", "error", err)
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"message": "Failed to read user information."})
		return
	}

	u := User{}
	if err = json.Unmarshal(data, &u); err != nil {
		logger.Error("Failed to unmarshal user info", "error", err)
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"message": "Error processing user information."})
		return
	}

	if u.Email == "" {
		logger.Error("Empty email in user info")
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"message": "Failed to get user email."})
		return
	}

	session.Set("user-id", u.Email)
	if err = session.Save(); err != nil {
		logger.Error("Failed to save session", "error", err)
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"message": "Error while saving session."})
		return
	}

	existingUser, err := h.db.LoadUser(u.Email)
	if err == nil && existingUser != nil {
		logger.Info("Returning user authenticated", "email", u.Email)
		c.HTML(http.StatusOK, "battle.tmpl", gin.H{"email": u.Email, "seen": true})
		return
	}

	if err = h.db.SaveUser(&u); err != nil {
		logger.Error("Failed to save user", "error", err, "email", u.Email)
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"message": "Error while saving user."})
		return
	}

	logger.Info("New user registered", "email", u.Email)
	c.HTML(http.StatusOK, "battle.tmpl", gin.H{"email": u.Email, "seen": false})
}

// LoginHandler handles the login procedure.
func (h *Handlers) LoginHandler(c *gin.Context) {
	logger := slog.Default()
	state, err := RandToken(32)
	if err != nil {
		logger.Error("Failed to generate state token", "error", err)
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"message": "Error while generating session data."})
		return
	}

	session := sessions.Default(c)
	session.Set("state", state)
	if err = session.Save(); err != nil {
		logger.Error("Failed to save session", "error", err)
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"message": "Error while saving session."})
		return
	}

	link := h.getLoginURL(state)
	logger.Info("Login initiated")
	c.HTML(http.StatusOK, "auth.tmpl", gin.H{"link": link})
}

// FieldHandler is a rudimentary handler for logged in users.
func (h *Handlers) FieldHandler(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user-id")
	c.HTML(http.StatusOK, "field.tmpl", gin.H{"user": userID})
}

// LogoutHandler handles user logout by clearing the session.
func (h *Handlers) LogoutHandler(c *gin.Context) {
	logger := slog.Default()
	session := sessions.Default(c)
	
	userID := session.Get("user-id")
	if userID != nil {
		logger.Info("User logged out", "user", userID)
	}
	
	// Clear all session data
	session.Clear()
	if err := session.Save(); err != nil {
		logger.Error("Failed to clear session", "error", err)
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"message": "Error during logout. Please try again."})
		return
	}
	
	// Redirect to home page with a success message
	c.HTML(http.StatusOK, "logout.tmpl", gin.H{})
}
