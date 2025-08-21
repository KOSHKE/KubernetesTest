package handlers

import (
	"api-gateway/internal/clients"
	"api-gateway/pkg/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	http.BaseHandler
	userClient clients.UserClient
}

func NewUserHandler(userClient clients.UserClient) *UserHandler {
	return &UserHandler{userClient: userClient}
}

func (h *UserHandler) Register(c *gin.Context) {
	var req http.RegisterRequest
	if !http.ValidateRequest(c, &req) {
		return
	}

	if h.HandleUserClientOperation(c, func() error {
		_, err := h.userClient.Register(c.Request.Context(), req.ToClientRequest())
		return err
	}, "register user") {
		http.RespondCreated(c, gin.H{"message": "User registered successfully"}, "User registered successfully")
	}
}

func (h *UserHandler) Login(c *gin.Context) {
	var req http.LoginRequest
	if !http.ValidateRequest(c, &req) {
		return
	}

	if h.HandleUserClientOperation(c, func() error {
		_, err := h.userClient.Login(c.Request.Context(), req.ToClientRequest())
		return err
	}, "authenticate user") {
		http.RespondSuccess(c, gin.H{"message": "Login successful"}, "Login successful")
	}
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	// No auth: demo-only
	userID, ok := http.RequireUserID(c)
	if !ok {
		return // Error response already sent by RequireUserID
	}

	if h.HandleUserClientOperation(c, func() error {
		_, err := h.userClient.GetUser(c.Request.Context(), userID)
		return err
	}, "get user profile") {
		http.RespondSuccess(c, gin.H{"message": "Profile retrieved successfully"}, "Profile retrieved successfully")
	}
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	// No auth: demo-only
	userID, ok := http.RequireUserID(c)
	if !ok {
		return // Error response already sent by RequireUserID
	}
	var req http.RegisterRequest
	if !http.ValidateRequest(c, &req) {
		return
	}

	if h.HandleUserClientOperation(c, func() error {
		_, err := h.userClient.UpdateUser(c.Request.Context(), userID, req.ToClientRequest())
		return err
	}, "update user profile") {
		http.RespondSuccess(c, gin.H{"message": "Profile updated successfully"}, "Profile updated successfully")
	}
}
