package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"os"
	"time"

	"api-gateway/internal/config"
	"api-gateway/internal/session"
	"api-gateway/internal/token"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
)

type AuthHandler struct {
	sessions session.SessionStore
	minter   token.AccessTokenMinter
	cfg      *config.Config
}

func NewAuthHandlerWithDeps(store session.SessionStore, minter token.AccessTokenMinter, cfg *config.Config) *AuthHandler {
	return &AuthHandler{sessions: store, minter: minter, cfg: cfg}
}

func NewAuthHandler() *AuthHandler { return &AuthHandler{} }

// Refresh: opaque refresh token через SessionStore + выпуск нового access через AccessTokenMinter
func (h *AuthHandler) Refresh(c *gin.Context) {
	if h.sessions == nil || h.minter == nil || h.cfg == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "session store not ready"})
		return
	}
	cookie, err := c.Request.Cookie("refresh_token")
	if err != nil || cookie.Value == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing refresh token"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()
	// найти сессию и ротировать
	newRefresh, userID, err := h.sessions.Rotate(ctx, cookie.Value, 7*24*time.Hour)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}
	// запросить новый access у user-service internal mint
	accessToken, mintErr := h.minter.MintAccessToken(ctx, userID)
	secure := os.Getenv("COOKIE_SECURE") == "true"
	httpOnly := true
	c.SetCookie("refresh_token", newRefresh, int((7*24*time.Hour)/time.Second), "/", "", secure, httpOnly)
	if mintErr != nil || accessToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to mint access token"})
		return
	}
	c.SetCookie("access_token", accessToken, 15*60, "/", "", secure, true)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// Logout: удалить refresh из SessionStore и очистить куки
func (h *AuthHandler) Logout(c *gin.Context) {
	secure := os.Getenv("COOKIE_SECURE") == "true"
	httpOnly := true
	if cookie, err := c.Request.Cookie("refresh_token"); err == nil && cookie.Value != "" && h.sessions != nil {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		_ = h.sessions.Delete(ctx, cookie.Value)
		cancel()
	}
	c.SetCookie("access_token", "", -1, "/", "", secure, httpOnly)
	c.SetCookie("refresh_token", "", -1, "/", "", secure, httpOnly)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func generateOpaqueToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
