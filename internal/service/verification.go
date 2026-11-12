package service

import (
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/suanova/medrouter/internal/common"
	"github.com/suanova/medrouter/internal/config"
	"github.com/suanova/medrouter/internal/model"
)

type VerificationService struct {
	cfg   *config.Config
	email *EmailService
}

func NewVerificationService(cfg *config.Config, email *EmailService) *VerificationService {
	return &VerificationService{cfg: cfg, email: email}
}

// Send issues a registration verification code after checking format,
// existence, cooldown and per-IP quota.
func (s *VerificationService) Send(rawEmail, purpose, ip string) *common.APIError {
	if purpose != model.PurposeRegister {
		return common.NewAPIError("INVALID_REQUEST", "Unsupported verification purpose")
	}
	if !s.cfg.EmailVerificationEnabled {
		return common.NewAPIError("EMAIL_VERIFICATION_DISABLED", "Email verification is disabled")
	}

	email := common.NormalizeEmail(rawEmail)
	if err := common.ValidateEmail(email); err != nil {
		return err
	}

	var count int64
	if err := model.DB.Model(&model.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return common.NewAPIErrorStatus(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to check email")
	}
	if count > 0 {
		return common.NewAPIError("EMAIL_TAKEN", "This email is already registered")
	}

	now := time.Now().UTC()

	var ipCount int64
	if err := model.DB.Model(&model.VerificationCode{}).
		Where("ip = ? AND created_at > ?", truncate(ip, 64), now.Add(-time.Hour)).
		Count(&ipCount).Error; err != nil {
		return common.NewAPIErrorStatus(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to check rate limit")
	}
	if int(ipCount) >= s.cfg.VerificationIPHourly {
		return common.NewAPIErrorStatus(http.StatusTooManyRequests, "RATE_LIMITED", "Too many verification emails requested from this address")
	}

	var latest model.VerificationCode
	err := model.DB.Where("email = ? AND purpose = ?", email, purpose).
		Order("created_at DESC").First(&latest).Error
	if err == nil && now.Sub(latest.CreatedAt) < s.cfg.VerificationResend {
		wait := int(s.cfg.VerificationResend.Seconds() - now.Sub(latest.CreatedAt).Seconds())
		if wait < 1 {
			wait = 1
		}
		return common.NewAPIErrorStatus(http.StatusTooManyRequests, "RATE_LIMITED",
			"Please wait before requesting another code")
	} else if err != nil && err != gorm.ErrRecordNotFound {
		return common.NewAPIErrorStatus(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to check cooldown")
	}

	code, err := common.NewVerificationCode()
	if err != nil {
		return common.NewAPIErrorStatus(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate verification code")
	}
	record := &model.VerificationCode{
		Email:     email,
		Purpose:   purpose,
		CodeHash:  common.VerificationCodeHash(s.cfg.SessionSecret, email, purpose, code),
		IP:        truncate(ip, 64),
		ExpiresAt: now.Add(s.cfg.VerificationCodeTTL),
		CreatedAt: now,
	}

	if err := model.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.VerificationCode{}).
			Where("email = ? AND purpose = ? AND consumed_at IS NULL", email, purpose).
			Update("consumed_at", now).Error; err != nil {
			return err
		}
		return tx.Create(record).Error
	}); err != nil {
		return common.NewAPIErrorStatus(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to store verification code")
	}

	if err := s.email.SendVerificationCode(email, code, purpose); err != nil {
		// Do not leave a code the user never received.
		_ = model.DB.Delete(&model.VerificationCode{}, record.ID).Error
		return common.NewAPIErrorStatus(http.StatusInternalServerError, "EMAIL_SEND_FAILED", "Failed to send verification email")
	}
	return nil
}

// Verify checks a code without consuming it. Failed attempts are persisted
// immediately so that a later transaction rollback cannot reset the counter.
func (s *VerificationService) Verify(rawEmail, purpose, code string) (*model.VerificationCode, *common.APIError) {
	if !s.cfg.EmailVerificationEnabled {
		return nil, common.NewAPIError("EMAIL_VERIFICATION_DISABLED", "Email verification is disabled")
	}
	email := common.NormalizeEmail(rawEmail)
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, common.NewAPIError("VERIFICATION_REQUIRED", "Please request and enter a verification code")
	}

	var record model.VerificationCode
	err := model.DB.Where("email = ? AND purpose = ? AND consumed_at IS NULL", email, purpose).
		Order("created_at DESC").First(&record).Error
	if err != nil {
		return nil, common.NewAPIError("VERIFICATION_REQUIRED", "Please request a verification code first")
	}

	now := time.Now().UTC()
	if now.After(record.ExpiresAt) {
		return nil, common.NewAPIError("VERIFICATION_EXPIRED", "The verification code has expired, please request a new one")
	}
	if record.Attempts >= s.cfg.VerificationMaxAttempts {
		return nil, common.NewAPIError("VERIFICATION_TOO_MANY_ATTEMPTS", "Too many incorrect attempts, please request a new code")
	}

	expected := common.VerificationCodeHash(s.cfg.SessionSecret, email, purpose, code)
	if !common.ConstantTimeEqual(expected, record.CodeHash) {
		_ = model.DB.Model(&model.VerificationCode{}).
			Where("id = ? AND consumed_at IS NULL", record.ID).
			UpdateColumn("attempts", gorm.Expr("attempts + 1")).Error
		if record.Attempts+1 >= s.cfg.VerificationMaxAttempts {
			return nil, common.NewAPIError("VERIFICATION_TOO_MANY_ATTEMPTS", "Too many incorrect attempts, please request a new code")
		}
		return nil, common.NewAPIError("VERIFICATION_INVALID", "Incorrect verification code")
	}
	return &record, nil
}

// ConsumeTx marks a verified code as used inside the caller's transaction.
func (s *VerificationService) ConsumeTx(tx *gorm.DB, record *model.VerificationCode) *common.APIError {
	now := time.Now().UTC()
	res := tx.Model(&model.VerificationCode{}).
		Where("id = ? AND consumed_at IS NULL", record.ID).
		Update("consumed_at", now)
	if res.Error != nil {
		return common.NewAPIErrorStatus(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to consume verification code")
	}
	if res.RowsAffected == 0 {
		return common.NewAPIError("VERIFICATION_INVALID", "The verification code has already been used")
	}
	return nil
}
