package service

import (
	"net/http"
	"time"

	"github.com/cubeface/cubeface/internal/common"
	"github.com/cubeface/cubeface/internal/config"
	"github.com/cubeface/cubeface/internal/model"
)

type SessionService struct {
	cfg *config.Config
}

type SessionResult struct {
	Token   string
	Session *model.UserSession
}

func NewSessionService(cfg *config.Config) *SessionService {
	return &SessionService{cfg: cfg}
}

func (s *SessionService) hashToken(token string) string {
	return common.HMACHex(s.cfg.SessionSecret, "session:"+token)
}

// Create issues a new opaque session token and stores only its HMAC.
func (s *SessionService) Create(user *model.User, method, ip, userAgent string) (*SessionResult, *common.APIError) {
	token, err := common.RandomToken(32)
	if err != nil {
		return nil, common.NewAPIErrorStatus(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create session")
	}
	now := time.Now().UTC()
	session := &model.UserSession{
		ID:          common.NewUUID(),
		UserID:      user.ID,
		TokenHash:   s.hashToken(token),
		AuthVersion: user.AuthVersion,
		LoginMethod: method,
		IP:          truncate(ip, 64),
		UserAgent:   truncate(userAgent, 255),
		CreatedAt:   now,
		LastSeenAt:  now,
		ExpiresAt:   now.Add(s.cfg.SessionTTL),
	}
	if err := model.DB.Create(session).Error; err != nil {
		return nil, common.NewAPIErrorStatus(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create session")
	}
	return &SessionResult{Token: token, Session: session}, nil
}

// Resolve validates a raw cookie token and returns the live user and session.
// It performs sliding expiry, but only touches the row once per configured
// interval.
func (s *SessionService) Resolve(token string) (*model.User, *model.UserSession, *common.APIError) {
	if token == "" {
		return nil, nil, common.NewAPIErrorStatus(http.StatusUnauthorized, "UNAUTHENTICATED", "Please sign in")
	}
	var session model.UserSession
	err := model.DB.Where("token_hash = ?", s.hashToken(token)).First(&session).Error
	if err != nil {
		return nil, nil, common.NewAPIErrorStatus(http.StatusUnauthorized, "SESSION_EXPIRED", "Your session has expired")
	}
	now := time.Now().UTC()
	if !session.IsActive(now) {
		return nil, nil, common.NewAPIErrorStatus(http.StatusUnauthorized, "SESSION_EXPIRED", "Your session has expired")
	}

	var user model.User
	if err := model.DB.First(&user, session.UserID).Error; err != nil {
		return nil, nil, common.NewAPIErrorStatus(http.StatusUnauthorized, "SESSION_EXPIRED", "Your session has expired")
	}
	if !user.IsEnabled() {
		return nil, nil, common.NewAPIErrorStatus(http.StatusUnauthorized, "USER_DISABLED", "This account has been disabled")
	}
	if session.AuthVersion != user.AuthVersion {
		_ = model.DB.Model(&model.UserSession{}).
			Where("id = ? AND revoked_at IS NULL", session.ID).
			Update("revoked_at", now).Error
		return nil, nil, common.NewAPIErrorStatus(http.StatusUnauthorized, "SESSION_EXPIRED", "Your session has expired")
	}

	if now.Sub(session.LastSeenAt) >= s.cfg.SessionTouchInterval {
		expires := now.Add(s.cfg.SessionTTL)
		if err := model.DB.Model(&model.UserSession{}).
			Where("id = ? AND revoked_at IS NULL", session.ID).
			Updates(map[string]any{"last_seen_at": now, "expires_at": expires}).Error; err == nil {
			session.LastSeenAt = now
			session.ExpiresAt = expires
		}
	}
	return &user, &session, nil
}

// RevokeByToken revokes the session identified by a raw token. It is
// idempotent: an unknown or already revoked token is not an error.
func (s *SessionService) RevokeByToken(token string) error {
	if token == "" {
		return nil
	}
	now := time.Now().UTC()
	return model.DB.Model(&model.UserSession{}).
		Where("token_hash = ? AND revoked_at IS NULL", s.hashToken(token)).
		Update("revoked_at", now).Error
}

func (s *SessionService) RevokeAllForUser(userID int64) error {
	now := time.Now().UTC()
	return model.DB.Model(&model.UserSession{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", now).Error
}

func (s *SessionService) Cookie(token string) *http.Cookie {
	return &http.Cookie{
		Name:     s.cfg.SessionCookieName,
		Value:    token,
		Path:     cookiePath(s.cfg.BasePath),
		MaxAge:   int(s.cfg.SessionTTL.Seconds()),
		HttpOnly: true,
		Secure:   s.cfg.SessionCookieSecure,
		SameSite: http.SameSiteLaxMode,
	}
}

func (s *SessionService) ClearCookie() *http.Cookie {
	return &http.Cookie{
		Name:     s.cfg.SessionCookieName,
		Value:    "",
		Path:     cookiePath(s.cfg.BasePath),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   s.cfg.SessionCookieSecure,
		SameSite: http.SameSiteLaxMode,
	}
}

// CookieName is the name of the session cookie.
func (s *SessionService) CookieName() string { return s.cfg.SessionCookieName }

func cookiePath(basePath string) string {
	if basePath == "" {
		return "/"
	}
	return basePath + "/"
}

// CookiePath is exported for handlers that need to match the session scope.
func (s *SessionService) CookiePath() string { return cookiePath(s.cfg.BasePath) }

func truncate(value string, max int) string {
	runes := []rune(value)
	if len(runes) <= max {
		return value
	}
	return string(runes[:max])
}
