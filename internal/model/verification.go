package model

import "time"

const (
	PurposeRegister = "register"
)

type VerificationCode struct {
	ID         int64      `gorm:"primaryKey" json:"id"`
	Email      string     `gorm:"size:254;not null;index:idx_verification_lookup,priority:1" json:"email"`
	Purpose    string     `gorm:"size:32;not null;index:idx_verification_lookup,priority:2" json:"purpose"`
	CodeHash   string     `gorm:"size:64;not null" json:"-"`
	IP         string     `gorm:"size:64" json:"ip"`
	Attempts   int        `gorm:"not null;default:0" json:"attempts"`
	ExpiresAt  time.Time  `gorm:"index" json:"expires_at"`
	ConsumedAt *time.Time `json:"-"`
	CreatedAt  time.Time  `gorm:"index:idx_verification_lookup,priority:3" json:"created_at"`
}

func (VerificationCode) TableName() string { return "verification_codes" }

func (v *VerificationCode) IsConsumed() bool { return v.ConsumedAt != nil }
