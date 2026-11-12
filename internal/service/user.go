package service

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/suanova/medrouter/internal/common"
	"github.com/suanova/medrouter/internal/config"
	"github.com/suanova/medrouter/internal/model"
)

type UserService struct {
	cfg          *config.Config
	sessions     *SessionService
	verification *VerificationService
	dummyHash    string
	loginLimiter *common.RateLimiter
	regLimiter   *common.RateLimiter
}

type RegisterInput struct {
	Username         string
	Email            string
	Password         string
	VerificationCode string
}

func NewUserService(cfg *config.Config, sessions *SessionService, verification *VerificationService) *UserService {
	dummyToken, _ := common.RandomToken(24)
	dummyHash, _ := common.HashPassword(dummyToken)
	return &UserService{
		cfg:          cfg,
		sessions:     sessions,
		verification: verification,
		dummyHash:    dummyHash,
		loginLimiter: common.NewRateLimiter(cfg.LoginRatePerMinute, time.Minute),
		regLimiter:   common.NewRateLimiter(cfg.RegisterRatePerHour, time.Hour),
	}
}

func (s *UserService) Register(input RegisterInput, ip, userAgent string) (*model.User, *common.APIError) {
	if !s.cfg.RegisterEnabled {
		return nil, common.NewAPIError("REGISTER_DISABLED", "Registration is currently disabled")
	}
	if ok, _ := s.regLimiter.Allow(truncate(ip, 64)); !ok {
		return nil, common.NewAPIErrorStatus(http.StatusTooManyRequests, "RATE_LIMITED", "Too many registration attempts, please try again later")
	}

	rawUsername := strings.TrimSpace(input.Username)
	if err := common.ValidateUsername(rawUsername); err != nil {
		return nil, err
	}
	username := common.NormalizeUsername(rawUsername)
	password := input.Password
	if err := common.ValidatePassword(password); err != nil {
		return nil, err
	}

	email := common.NormalizeEmail(input.Email)
	if s.cfg.EmailVerificationEnabled {
		if err := common.ValidateEmail(email); err != nil {
			return nil, err
		}
		if strings.TrimSpace(input.VerificationCode) == "" {
			return nil, common.NewAPIError("VERIFICATION_REQUIRED", "Please request and enter a verification code")
		}
	} else if email != "" {
		if err := common.ValidateEmail(email); err != nil {
			return nil, err
		}
	}

	// Cheap existence pre-checks produce precise errors. The unique indexes are
	// the actual guarantee and are handled again below.
	var count int64
	if err := model.DB.Model(&model.User{}).Where("username = ?", username).Count(&count).Error; err != nil {
		return nil, common.NewAPIErrorStatus(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to check username")
	}
	if count > 0 {
		return nil, common.NewAPIError("USERNAME_TAKEN", "This username is already taken")
	}
	if email != "" {
		if err := model.DB.Model(&model.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
			return nil, common.NewAPIErrorStatus(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to check email")
		}
		if count > 0 {
			return nil, common.NewAPIError("EMAIL_TAKEN", "This email is already registered")
		}
	}

	var codeRecord *model.VerificationCode
	if s.cfg.EmailVerificationEnabled {
		var verifyErr *common.APIError
		codeRecord, verifyErr = s.verification.Verify(email, model.PurposeRegister, input.VerificationCode)
		if verifyErr != nil {
			return nil, verifyErr
		}
	}

	hash, err := common.HashPassword(password)
	if err != nil {
		return nil, common.NewAPIErrorStatus(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to process password")
	}

	now := time.Now().UTC()
	user := &model.User{
		Username:     username,
		DisplayName:  truncate(rawUsername, 64),
		PasswordHash: hash,
		Email:        email,
		Role:         model.RoleCommon,
		Status:       model.StatusEnabled,
		AuthVersion:  1,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if s.cfg.EmailVerificationEnabled && codeRecord != nil {
		user.EmailVerifiedAt = &now
	}

	txErr := model.DB.Transaction(func(tx *gorm.DB) error {
		if codeRecord != nil {
			if apiErr := s.verification.ConsumeTx(tx, codeRecord); apiErr != nil {
				return apiErr
			}
		}
		return tx.Create(user).Error
	})
	if txErr != nil {
		var apiErr *common.APIError
		if errors.As(txErr, &apiErr) {
			return nil, apiErr
		}
		if model.DB.Model(&model.User{}).Where("username = ?", username).Count(&count).Error == nil && count > 0 {
			return nil, common.NewAPIError("USERNAME_TAKEN", "This username is already taken")
		}
		if email != "" && model.DB.Model(&model.User{}).Where("email = ?", email).Count(&count).Error == nil && count > 0 {
			return nil, common.NewAPIError("EMAIL_TAKEN", "This email is already registered")
		}
		return nil, common.NewAPIErrorStatus(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create account")
	}

	return user, nil
}

// Login accepts a username or an email address as the identifier.
func (s *UserService) Login(identifier, password, ip, userAgent string) (*model.User, *SessionResult, *common.APIError) {
	identifier = strings.TrimSpace(identifier)
	if identifier == "" || password == "" {
		return nil, nil, common.NewAPIError("INVALID_CREDENTIALS", "Invalid username or password")
	}
	key := truncate(ip, 64) + "|" + strings.ToLower(identifier)
	if ok, _ := s.loginLimiter.Allow(key); !ok {
		return nil, nil, common.NewAPIErrorStatus(http.StatusTooManyRequests, "RATE_LIMITED", "Too many sign-in attempts, please try again later")
	}

	normalized := strings.ToLower(identifier)
	var user model.User
	err := model.DB.Where("username = ? OR email = ?", normalized, normalized).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Equalize timing so unknown accounts cannot be distinguished.
			common.CheckPassword(s.dummyHash, password)
			return nil, nil, common.NewAPIError("INVALID_CREDENTIALS", "Invalid username or password")
		}
		return nil, nil, common.NewAPIErrorStatus(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to sign in")
	}

	if !common.CheckPassword(user.PasswordHash, password) {
		return nil, nil, common.NewAPIError("INVALID_CREDENTIALS", "Invalid username or password")
	}
	if !user.IsEnabled() {
		return nil, nil, common.NewAPIError("USER_DISABLED", "This account has been disabled")
	}

	result, apiErr := s.sessions.Create(&user, model.LoginMethodPassword, ip, userAgent)
	if apiErr != nil {
		return nil, nil, apiErr
	}

	now := time.Now().UTC()
	if err := model.DB.Model(&model.User{}).Where("id = ?", user.ID).Update("last_login_at", now).Error; err != nil {
		return nil, nil, common.NewAPIErrorStatus(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update sign-in time")
	}
	user.LastLoginAt = &now
	return &user, result, nil
}

func (s *UserService) FindByID(id int64) (*model.User, error) {
	var user model.User
	if err := model.DB.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// FindOrCreateByOAuth resolves an OAuth identity to a local user, creating a
// passwordless account on first sign-in.
func (s *UserService) FindOrCreateByOAuth(identity OAuthIdentity) (*model.User, *common.APIError) {
	email := common.NormalizeEmail(identity.Email)

	var binding model.OAuthBinding
	err := model.DB.Where("provider = ? AND provider_user_id = ?", identity.Provider, identity.ProviderUserID).
		First(&binding).Error
	if err == nil {
		user, loadErr := s.FindByID(binding.UserID)
		if loadErr != nil {
			return nil, common.NewAPIErrorStatus(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to load account")
		}
		if !user.IsEnabled() {
			return nil, common.NewAPIError("USER_DISABLED", "This account has been disabled")
		}
		return user, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, common.NewAPIErrorStatus(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to resolve sign-in")
	}

	if email != "" && identity.EmailVerified {
		var existing model.User
		findErr := model.DB.Where("email = ?", email).First(&existing).Error
		if findErr == nil {
			if !existing.IsEnabled() {
				return nil, common.NewAPIError("USER_DISABLED", "This account has been disabled")
			}
			if err := model.DB.Transaction(func(tx *gorm.DB) error {
				b := &model.OAuthBinding{
					Provider:       identity.Provider,
					ProviderUserID: identity.ProviderUserID,
					UserID:         existing.ID,
					Email:          email,
				}
				if err := tx.Create(b).Error; err != nil {
					return err
				}
				if !existing.EmailVerified() {
					now := time.Now().UTC()
					if err := tx.Model(&model.User{}).Where("id = ?", existing.ID).
						Update("email_verified_at", now).Error; err != nil {
						return err
					}
					existing.EmailVerifiedAt = &now
				}
				return nil
			}); err != nil {
				// A concurrent callback may have created the binding first.
				var concurrent model.OAuthBinding
				if e := model.DB.Where("provider = ? AND provider_user_id = ?", identity.Provider, identity.ProviderUserID).
					First(&concurrent).Error; e == nil {
					user, loadErr := s.FindByID(concurrent.UserID)
					if loadErr == nil {
						return user, nil
					}
				}
				return nil, common.NewAPIErrorStatus(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to link account")
			}
			return &existing, nil
		}
		if !errors.Is(findErr, gorm.ErrRecordNotFound) {
			return nil, common.NewAPIErrorStatus(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to resolve sign-in")
		}
	}

	randomPassword, err := common.RandomToken(32)
	if err != nil {
		return nil, common.NewAPIErrorStatus(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create account")
	}
	hash, err := common.HashPassword(randomPassword)
	if err != nil {
		return nil, common.NewAPIErrorStatus(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create account")
	}

	baseUsername := sanitizeOAuthUsername(identity)
	displayName := truncate(strings.TrimSpace(identity.DisplayName), 64)
	if displayName == "" {
		displayName = baseUsername
	}

	now := time.Now().UTC()
	var created *model.User
	txErr := model.DB.Transaction(func(tx *gorm.DB) error {
		username, err := uniqueUsername(tx, baseUsername)
		if err != nil {
			return err
		}
		user := &model.User{
			Username:     username,
			DisplayName:  displayName,
			PasswordHash: hash,
			Email:        email,
			Role:         model.RoleCommon,
			Status:       model.StatusEnabled,
			AuthVersion:  1,
			CreatedAt:    now,
			UpdatedAt:    now,
			LastLoginAt:  &now,
		}
		if email != "" && identity.EmailVerified {
			user.EmailVerifiedAt = &now
		}
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		b := &model.OAuthBinding{
			Provider:       identity.Provider,
			ProviderUserID: identity.ProviderUserID,
			UserID:         user.ID,
			Email:          email,
			CreatedAt:      now,
		}
		if err := tx.Create(b).Error; err != nil {
			return err
		}
		created = user
		return nil
	})
	if txErr != nil {
		// Concurrent first sign-in or a concurrent email registration won the
		// race; re-resolve instead of failing the user.
		var concurrent model.OAuthBinding
		if e := model.DB.Where("provider = ? AND provider_user_id = ?", identity.Provider, identity.ProviderUserID).
			First(&concurrent).Error; e == nil {
			if user, loadErr := s.FindByID(concurrent.UserID); loadErr == nil {
				return user, nil
			}
		}
		if email != "" {
			var concurrentUser model.User
			if e := model.DB.Where("email = ?", email).First(&concurrentUser).Error; e == nil {
				if !concurrentUser.IsEnabled() {
					return nil, common.NewAPIError("USER_DISABLED", "This account has been disabled")
				}
				return nil, common.NewAPIError("OAUTH_EMAIL_UNAVAILABLE", "This email is already registered, please sign in with your password first")
			}
		}
		return nil, common.NewAPIErrorStatus(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create account")
	}
	return created, nil
}

func sanitizeOAuthUsername(identity OAuthIdentity) string {
	candidate := strings.ToLower(strings.TrimSpace(identity.Username))
	if candidate == "" && identity.Email != "" {
		candidate = strings.ToLower(strings.SplitN(identity.Email, "@", 2)[0])
	}
	if candidate == "" {
		candidate = "user"
	}

	var b strings.Builder
	lastDash := false
	for _, r := range candidate {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '_', r == '.':
			b.WriteRune(r)
			lastDash = false
		case r == '-' || r == ' ':
			if !lastDash {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}
	cleaned := strings.Trim(b.String(), "-.")
	if len(cleaned) < 3 {
		suffix, _ := common.RandomToken(4)
		suffix = strings.ToLower(strings.NewReplacer("-", "", "_", "").Replace(suffix))
		if len(suffix) > 6 {
			suffix = suffix[:6]
		}
		cleaned = "user-" + suffix
	}
	if len(cleaned) > 26 {
		cleaned = strings.Trim(cleaned[:26], "-.")
	}
	if len(cleaned) < 3 {
		cleaned = "user"
	}
	// Must start with a letter or digit.
	first := cleaned[0]
	if !(first >= 'a' && first <= 'z') && !(first >= '0' && first <= '9') {
		cleaned = "u" + cleaned
	}
	if err := common.ValidateUsername(cleaned); err != nil {
		cleaned = "user"
	}
	return cleaned
}

func uniqueUsername(tx *gorm.DB, base string) (string, error) {
	for attempt := 0; attempt < 8; attempt++ {
		candidate := base
		if attempt > 0 {
			suffix, err := common.RandomToken(4)
			if err != nil {
				return "", err
			}
			suffix = strings.ToLower(strings.NewReplacer("-", "", "_", "").Replace(suffix))
			if len(suffix) > 5 {
				suffix = suffix[:5]
			}
			candidate = fmt.Sprintf("%s-%s", base, suffix)
			if len(candidate) > 32 {
				candidate = candidate[:32]
				candidate = strings.TrimRight(candidate, "-.")
			}
		}
		var count int64
		if err := tx.Model(&model.User{}).Where("username = ?", candidate).Count(&count).Error; err != nil {
			return "", err
		}
		if count == 0 {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("could not allocate a unique username")
}
