package controller

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"

	"github.com/suanova/medrouter/internal/config"
	"github.com/suanova/medrouter/internal/service"
)

type OAuthController struct {
	cfg      *config.Config
	oauth    *service.OAuthService
	sessions *service.SessionService
}

func NewOAuthController(cfg *config.Config, oauth *service.OAuthService, sessions *service.SessionService) *OAuthController {
	return &OAuthController{cfg: cfg, oauth: oauth, sessions: sessions}
}

func (h *OAuthController) Start(c *gin.Context) {
	provider := c.Param("provider")
	authURL, cookie, apiErr := h.oauth.Start(provider, c.Query("redirect"))
	if apiErr != nil {
		h.redirectError(c, apiErr.Code)
		return
	}
	c.SetCookieData(cookie)
	c.Redirect(http.StatusFound, authURL)
}

func (h *OAuthController) Callback(c *gin.Context) {
	provider := c.Param("provider")
	stateCookie, _ := c.Cookie(h.oauth.StateCookieName())
	redirect, _ /* user */, result, apiErr := h.oauth.Complete(
		c.Request.Context(), provider, c.Query("code"), stateCookie, c.ClientIP(), c.Request.UserAgent(),
	)
	c.SetCookieData(h.oauth.ClearStateCookie())

	if apiErr != nil {
		h.redirectError(c, apiErr.Code)
		return
	}
	if result != nil {
		c.SetCookieData(h.sessions.Cookie(result.Token))
	}
	if redirect == "" {
		redirect = "/dashboard"
	}
	c.Redirect(http.StatusFound, h.cfg.BasePath+redirect)
}

func (h *OAuthController) redirectError(c *gin.Context, code string) {
	target := h.cfg.FrontendPath("/sign-in") + "?oauth_error=" + url.QueryEscape(code)
	c.Redirect(http.StatusFound, target)
}
