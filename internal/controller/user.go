package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/suanova/medrouter/internal/common"
	"github.com/suanova/medrouter/internal/config"
	"github.com/suanova/medrouter/internal/model"
	"github.com/suanova/medrouter/internal/service"
)

type UserController struct {
	cfg      *config.Config
	users    *service.UserService
	sessions *service.SessionService
}

func NewUserController(cfg *config.Config, users *service.UserService, sessions *service.SessionService) *UserController {
	return &UserController{cfg: cfg, users: users, sessions: sessions}
}

type registerRequest struct {
	Username         string `json:"username"`
	Email            string `json:"email"`
	Password         string `json:"password"`
	VerificationCode string `json:"verification_code"`
}

func (h *UserController) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.NewAPIError("INVALID_REQUEST", "Invalid request body"))
		return
	}
	user, apiErr := h.users.Register(service.RegisterInput{
		Username:         req.Username,
		Email:            req.Email,
		Password:         req.Password,
		VerificationCode: req.VerificationCode,
	}, c.ClientIP(), c.Request.UserAgent())
	if apiErr != nil {
		common.Fail(c, apiErr)
		return
	}
	common.Success(c, gin.H{"user": user.ToDTO()})
}

type loginRequest struct {
	Username   string `json:"username"`
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

func (h *UserController) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.NewAPIError("INVALID_REQUEST", "Invalid request body"))
		return
	}
	identifier := req.Username
	if identifier == "" {
		identifier = req.Identifier
	}
	user, result, apiErr := h.users.Login(identifier, req.Password, c.ClientIP(), c.Request.UserAgent())
	if apiErr != nil {
		common.Fail(c, apiErr)
		return
	}
	c.SetCookieData(h.sessions.Cookie(result.Token))
	common.Success(c, gin.H{
		"user":       user.ToDTO(),
		"expires_at": result.Session.ExpiresAt,
	})
}

func (h *UserController) Logout(c *gin.Context) {
	if token, err := c.Cookie(h.sessions.CookieName()); err == nil && token != "" {
		_ = h.sessions.RevokeByToken(token)
	}
	c.SetCookieData(h.sessions.ClearCookie())
	common.Success(c, nil)
}

func (h *UserController) Self(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		common.Fail(c, common.NewAPIErrorStatus(http.StatusUnauthorized, "UNAUTHENTICATED", "Please sign in"))
		return
	}
	common.Success(c, gin.H{"user": user.ToDTO()})
}

func currentUser(c *gin.Context) *model.User {
	value, ok := c.Get("medrouter.user")
	if !ok {
		return nil
	}
	user, _ := value.(*model.User)
	return user
}
