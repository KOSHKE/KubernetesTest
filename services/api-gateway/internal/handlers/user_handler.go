package handlers

import (
	"net/http"
	"os"
	"time"

	"api-gateway/internal/clients"
	"api-gateway/internal/config"
	"api-gateway/internal/session"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
)

type UserHandler struct {
	userClient clients.UserClient
	sessions   session.SessionStore
	cfg        *config.Config
}

func NewUserHandler(userClient clients.UserClient) *UserHandler {
	return &UserHandler{userClient: userClient}
}

func NewUserHandlerWithSessions(userClient clients.UserClient, store session.SessionStore, cfg *config.Config) *UserHandler {
	return &UserHandler{userClient: userClient, sessions: store, cfg: cfg}
}

func (h *UserHandler) Register(c *gin.Context) {
	var req clients.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	if req.Email == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email and password are required"})
		return
	}
	response, err := h.userClient.Register(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}
	setAuthCookies(c, response.Token)
	if h.sessions != nil {
		if err := h.issueRefresh(c, response.User.ID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set session"})
			return
		}
	}
	c.JSON(http.StatusCreated, response)
}

func (h *UserHandler) Login(c *gin.Context) {
	var req clients.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	if req.Email == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email and password are required"})
		return
	}
	response, err := h.userClient.Login(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
	setAuthCookies(c, response.Token)
	if h.sessions != nil {
		if err := h.issueRefresh(c, response.User.ID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set session"})
			return
		}
	}
	c.JSON(http.StatusOK, response)
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	user, err := h.userClient.GetUser(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user profile"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}

func setAuthCookies(c *gin.Context, accessToken string) {
	// Dev-noauth: do nothing
	_ = os.Getenv
}

func (h *UserHandler) issueRefresh(c *gin.Context, userID string) error {
	// Dev-noauth: no refresh cookies
	_ = time.Second
	_ = context.Background
	return nil
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	var req clients.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	user, err := h.userClient.UpdateUser(c.Request.Context(), userID.(string), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user profile"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user, "message": "Profile updated successfully"})
}
