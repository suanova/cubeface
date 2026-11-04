package model

import "time"

type OAuthBinding struct {
	ID             int64     `gorm:"primaryKey" json:"id"`
	Provider       string    `gorm:"size:32;not null;uniqueIndex:idx_oauth_identity,priority:1" json:"provider"`
	ProviderUserID string    `gorm:"size:191;not null;uniqueIndex:idx_oauth_identity,priority:2" json:"provider_user_id"`
	UserID         int64     `gorm:"not null;index" json:"user_id"`
	Email          string    `gorm:"size:254" json:"email"`
	CreatedAt      time.Time `json:"created_at"`
}

func (OAuthBinding) TableName() string { return "oauth_bindings" }
