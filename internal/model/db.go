package model

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/suanova/medrouter/internal/common"
	"github.com/suanova/medrouter/internal/config"
)

var DB *gorm.DB

// InitDB opens the configured database, applies migrations and seeds the
// optional initial administrator account.
func InitDB(cfg *config.Config) error {
	gormCfg := &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "[gorm] ", log.LstdFlags),
			logger.Config{
				SlowThreshold:             500 * time.Millisecond,
				LogLevel:                  logger.Warn,
				IgnoreRecordNotFoundError: true,
				Colorful:                  false,
			},
		),
		NowFunc: func() time.Time { return time.Now().UTC() },
	}

	var dialector gorm.Dialector
	isPostgres := strings.HasPrefix(cfg.SQLDSN, "postgres://") || strings.HasPrefix(cfg.SQLDSN, "postgresql://")

	if cfg.SQLDSN != "" && !isPostgres {
		return fmt.Errorf("SQL_DSN must start with postgres:// or postgresql:// (leave empty for SQLite)")
	}

	if isPostgres {
		dialector = postgres.Open(cfg.SQLDSN)
	} else {
		if err := os.MkdirAll(filepath.Dir(cfg.SQLitePath), 0o755); err != nil {
			return fmt.Errorf("create sqlite directory: %w", err)
		}
		dsn := fmt.Sprintf(
			"file:%s?_pragma=busy_timeout(30000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(1)",
			cfg.SQLitePath,
		)
		dialector = sqlite.Open(dsn)
	}

	db, err := gorm.Open(dialector, gormCfg)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get sql.DB: %w", err)
	}
	if isPostgres {
		sqlDB.SetMaxOpenConns(cfg.DBMaxOpenConns)
		sqlDB.SetMaxIdleConns(maxInt(2, cfg.DBMaxOpenConns/2))
		sqlDB.SetConnMaxLifetime(time.Hour)
	} else {
		// SQLite is a single writer. A single connection avoids SQLITE_BUSY
		// between concurrent writers; read concurrency is not a first-version
		// concern. Do not share one SQLite file between processes.
		sqlDB.SetMaxOpenConns(1)
	}
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	if isPostgres {
		log.Printf("database: postgresql")
	} else {
		log.Printf("database: sqlite path=%s", cfg.SQLitePath)
	}

	if err := db.AutoMigrate(&User{}, &UserSession{}, &VerificationCode{}, &OAuthBinding{}); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}
	DB = db

	if err := seedAdmin(cfg); err != nil {
		return err
	}
	return nil
}

func seedAdmin(cfg *config.Config) error {
	username := common.NormalizeUsername(cfg.InitialAdminUser)
	if username == "" || cfg.InitialAdminPass == "" {
		return nil
	}
	if err := common.ValidateUsername(username); err != nil {
		return fmt.Errorf("INITIAL_ADMIN_USERNAME invalid: %w", err)
	}
	if err := common.ValidatePassword(cfg.InitialAdminPass); err != nil {
		return fmt.Errorf("INITIAL_ADMIN_PASSWORD invalid: %w", err)
	}

	var existing User
	err := DB.Where("username = ?", username).First(&existing).Error
	if err == nil {
		return nil
	}
	if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("look up initial admin: %w", err)
	}

	hash, err := common.HashPassword(cfg.InitialAdminPass)
	if err != nil {
		return fmt.Errorf("hash initial admin password: %w", err)
	}

	admin := User{
		Username:     username,
		DisplayName:  username,
		PasswordHash: hash,
		Email:        common.NormalizeEmail(cfg.InitialAdminEmail),
		Role:         RoleRoot,
		Status:       StatusEnabled,
		AuthVersion:  1,
	}
	if admin.Email != "" {
		now := time.Now().UTC()
		admin.EmailVerifiedAt = &now
	}
	if err := DB.Create(&admin).Error; err != nil {
		return fmt.Errorf("create initial admin: %w", err)
	}
	log.Printf("seeded initial admin account username=%s", username)
	return nil
}

// Close closes the underlying database handle.
func Close() error {
	if DB == nil {
		return nil
	}
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
