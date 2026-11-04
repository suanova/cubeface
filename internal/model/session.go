package model

import "time"

const (
	LoginMethodPassword = "password"
	LoginMethodGitHub   = "github"
	LoginMethodGoogle   = "google"
	LoginMethodOIDC     = "oidc"
)

type UserSession struct {
	ID          string     `gorm:"primaryKey;size:36" json:"id"`
	UserID      int64      `gorm:"not null;index" json:"user_id"`
	TokenHash   string     `gorm:"size:64;not null;uniqueIndex" json:"-"`
	AuthVersion int64      `gorm:"not null" json:"-"`
	LoginMethod string     `gorm:"size:32;not null" json:"login_method"`
	IP          string     `gorm:"size:64" json:"ip"`
	UserAgent   string     `gorm:"size:255" json:"user_agent"`
	CreatedAt   time.Time  `json:"created_at"`
	LastSeenAt  time.Time  `json:"last_seen_at"`
	ExpiresAt   time.Time  `gorm:"index" json:"expires_at"`
	RevokedAt   *time.Time `json:"-"`
}

func (UserSession) TableName() string { return "user_sessions" }

func (s *UserSession) IsActive(now time.Time) bool {
	return s.RevokedAt == nil && now.Before(s.ExpiresAt)
}
