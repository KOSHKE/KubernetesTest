package handlers

import (
	"api-gateway/internal/clients"
	"api-gateway/internal/middleware"
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
		response, err := h.userClient.Login(c.Request.Context(), req.ToClientRequest())
		if err != nil {
			return err
		}

		// Debug logging

		// Return tokens in response
		c.JSON(200, gin.H{
			"message": "Login successful",
			"data": gin.H{
				"user": gin.H{
					"id":         response.User.ID,
					"email":      response.User.Email,
					"first_name": response.User.FirstName,
					"last_name":  response.User.LastName,
				},
				"access_token":  response.AccessToken,
				"refresh_token": response.RefreshToken,
				"expires_in":    response.ExpiresIn,
			},
		})
		return nil
	}, "authenticate user") {
		// Response already sent above
	}
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	// Get user ID from JWT context
	userID, ok := middleware.GetUserID(c)
	if !ok {
		http.RespondUnauthorized(c, "User not authenticated")
		return
	}

	if h.HandleUserClientOperation(c, func() error {
		_, err := h.userClient.GetUser(c.Request.Context(), userID)
		return err
	}, "get user profile") {
		http.RespondSuccess(c, gin.H{"message": "Profile retrieved successfully"}, "Profile retrieved successfully")
	}
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	// Get user ID from JWT context
	userID, ok := middleware.GetUserID(c)
	if !ok {
		http.RespondUnauthorized(c, "User not authenticated")
		return
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

func (h *UserHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if !http.ValidateRequest(c, &req) {
		return
	}

	if h.HandleUserClientOperation(c, func() error {
		accessToken, err := h.userClient.RefreshToken(c.Request.Context(), req.RefreshToken)
		if err != nil {
			return err
		}

		// Return new access token
		c.JSON(200, gin.H{
			"message": "Token refreshed successfully",
			"data": gin.H{
				"access_token": accessToken,
				"expires_in":   900, // 15 minutes in seconds
			},
		})
		return nil
	}, "refresh token") {
		// Response already sent above
	}
}

func (h *UserHandler) Logout(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if !http.ValidateRequest(c, &req) {
		return
	}

	if h.HandleUserClientOperation(c, func() error {
		err := h.userClient.Logout(c.Request.Context(), req.RefreshToken)
		return err
	}, "logout user") {
		http.RespondSuccess(c, gin.H{"message": "Logout successful"}, "Logout successful")
	}
}
