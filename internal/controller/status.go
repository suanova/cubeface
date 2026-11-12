package controller

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/suanova/medrouter/internal/common"
	"github.com/suanova/medrouter/internal/config"
	"github.com/suanova/medrouter/internal/model"
	"github.com/suanova/medrouter/internal/service"
)

type StatusController struct {
	cfg   *config.Config
	oauth *service.OAuthService
}

func NewStatusController(cfg *config.Config, oauth *service.OAuthService) *StatusController {
	return &StatusController{cfg: cfg, oauth: oauth}
}

type statusResponse struct {
	SystemName               string                 `json:"system_name"`
	Version                  string                 `json:"version"`
	RegisterEnabled          bool                   `json:"register_enabled"`
	EmailVerificationEnabled bool                   `json:"email_verification_enabled"`
	OAuthProviders           []service.ProviderInfo `json:"oauth_providers"`
	ServerTime               time.Time              `json:"server_time"`
}

func (h *StatusController) Get(c *gin.Context) {
	common.Success(c, statusResponse{
		SystemName:               h.cfg.SystemName,
		Version:                  h.cfg.Version,
		RegisterEnabled:          h.cfg.RegisterEnabled,
		EmailVerificationEnabled: h.cfg.EmailVerificationEnabled,
		OAuthProviders:           h.oauth.Providers(),
		ServerTime:               time.Now().UTC(),
	})
}

func (h *StatusController) Health(c *gin.Context) {
	sqlDB, err := model.DB.DB()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
