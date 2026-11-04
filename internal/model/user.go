package model

import (
	"time"
)

const (
	RoleGuest  = 0
	RoleCommon = 1
	RoleRoot   = 100

	StatusEnabled  = 1
	StatusDisabled = 2
)

type User struct {
	ID              int64      `gorm:"primaryKey" json:"id"`
	Username        string     `gorm:"size:32;not null;uniqueIndex" json:"username"`
	DisplayName     string     `gorm:"size:64;not null" json:"display_name"`
	PasswordHash    string     `gorm:"size:100;not null" json:"-"`
	Email           string     `gorm:"size:254;uniqueIndex:idx_users_email,where:email <> ''" json:"email"`
	EmailVerifiedAt *time.Time `json:"-"`
	Role            int        `gorm:"not null;default:1" json:"role"`
	Status          int        `gorm:"not null;default:1" json:"status"`
	AuthVersion     int64      `gorm:"not null;default:1" json:"-"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"-"`
	LastLoginAt     *time.Time `json:"last_login_at"`
}

func (User) TableName() string { return "users" }

func (u *User) IsEnabled() bool { return u.Status == StatusEnabled }

func (u *User) IsRoot() bool { return u.Role >= RoleRoot }

func (u *User) EmailVerified() bool { return u.EmailVerifiedAt != nil && !u.EmailVerifiedAt.IsZero() }

func (u *User) DisplayNameOrUsername() string {
	if u.DisplayName != "" {
		return u.DisplayName
	}
	return u.Username
}

type UserDTO struct {
	ID            int64      `json:"id"`
	Username      string     `json:"username"`
	DisplayName   string     `json:"display_name"`
	Email         string     `json:"email"`
	Role          int        `json:"role"`
	Status        int        `json:"status"`
	EmailVerified bool       `json:"email_verified"`
	CreatedAt     time.Time  `json:"created_at"`
	LastLoginAt   *time.Time `json:"last_login_at"`
}

// ToDTO strips secrets (password hash, auth version) before a user is
// serialized to an API response.
func (u *User) ToDTO() UserDTO {
	return UserDTO{
		ID:            u.ID,
		Username:      u.Username,
		DisplayName:   u.DisplayNameOrUsername(),
		Email:         u.Email,
		Role:          u.Role,
		Status:        u.Status,
		EmailVerified: u.EmailVerified(),
		CreatedAt:     u.CreatedAt.UTC(),
		LastLoginAt:   u.LastLoginAt,
	}
}
