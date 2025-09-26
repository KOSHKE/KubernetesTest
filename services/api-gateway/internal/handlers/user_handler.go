package handlers

import (
	"context"
	"time"

	dto "ecommerce-platform/pkg/common/dto/user-service"
	"ecommerce-platform/services/api-gateway/internal/clients"

	"github.com/gin-gonic/gin"
)

const userDefaultRPCTimeout = 3 * time.Second

type UserHandler struct {
	userClient clients.UserClient
}

func NewUserHandler(userClient clients.UserClient) *UserHandler {
	return &UserHandler{userClient: userClient}
}

// POST /api/v1/auth/register
func (h *UserHandler) Register(c *gin.Context) {
	var req dto.RegisterUserRequest
	if !validateJSONRequest(c, &req) {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), userDefaultRPCTimeout)
	defer cancel()
	resp, err := h.userClient.Register(ctx, &req)
	if !handleGRPCError(c, err) {
		return
	}
	respondCreated(c, gin.H{"user_id": resp.UserID}, "User registered successfully")
}

// POST /api/v1/auth/login
func (h *UserHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if !validateJSONRequest(c, &req) {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), userDefaultRPCTimeout)
	defer cancel()
	resp, err := h.userClient.Login(ctx, &req)
	if !handleGRPCError(c, err) {
		return
	}
	respondSuccess(c, resp, "Login successful")
}

// GET /api/v1/users/:id
func (h *UserHandler) GetUser(c *gin.Context) {
	// Require path param and ownership by JWT subject
	pathID, ok := requireParam(c, "id")
	if !ok {
		return
	}
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	if userID != pathID {
		respondError(c, 403, "forbidden", "Access denied")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), userDefaultRPCTimeout)
	defer cancel()
	resp, err := h.userClient.GetUser(ctx, pathID)
	if !handleGRPCError(c, err) {
		return
	}
	respondSuccess(c, resp, "User retrieved successfully")
}

// POST /api/v1/auth/refresh
func (h *UserHandler) Refresh(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if !validateJSONRequest(c, &req) {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), userDefaultRPCTimeout)
	defer cancel()
	resp, err := h.userClient.RefreshToken(ctx, &req)
	if !handleGRPCError(c, err) {
		return
	}
	respondSuccess(c, resp, "Token refreshed successfully")
}

// POST /api/v1/auth/logout
func (h *UserHandler) Logout(c *gin.Context) {
	// Require authenticated user
	if _, ok := requireUserID(c); !ok {
		return
	}
	var body struct {
		SessionID string `json:"session_id"`
	}
	if !validateJSONRequest(c, &body) {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), userDefaultRPCTimeout)
	defer cancel()
	if !handleGRPCError(c, h.userClient.Logout(ctx, body.SessionID)) {
		return
	}
	respondSuccess(c, gin.H{"message": "Logout successful"}, "Logout successful")
}
