package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Scarage1/url-shortener/internal/email"
	"github.com/Scarage1/url-shortener/internal/model"
	"github.com/Scarage1/url-shortener/internal/repository"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	require.NoError(t, err)

	// Set connection pool limits to ensure the memory DB connection stays alive
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)

	err = db.AutoMigrate(
		&model.User{},
		&model.URL{},
		&model.RoutingRule{},
		&model.Plan{},
		&model.Organization{},
		&model.OrganizationMember{},
		&model.Subscription{},
		&model.RefreshToken{},
		&model.AuditLog{},
	)
	require.NoError(t, err)

	// Seed free plan
	err = db.Create(&model.Plan{
		Name:        model.PlanFree,
		DisplayName: "Free",
	}).Error
	require.NoError(t, err)

	return db
}

func TestAuthService_LoginAndRefresh(t *testing.T) {
	db := setupTestDB(t)
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	userRepo := repository.NewUserRepository(db)
	orgSvc := NewOrgService(db)
	noopEmail := email.NoopSender{}
	auditSvc := NewAuditService(db)

	authSvc := NewAuthService(userRepo, orgSvc, auditSvc, "super-secret-key-that-is-long-enough", db, rClient, noopEmail, "http://localhost:8080")

	// 1. Register a user
	emailAddr := "test@example.com"
	password := "password123"
	err = authSvc.Register(emailAddr, password)
	require.NoError(t, err)

	// 2. Login
	accessToken, refreshToken, err := authSvc.Login(emailAddr, password)
	require.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)

	// 3. Refresh Session
	newAccess, newRefresh, err := authSvc.RefreshSession(refreshToken)
	require.NoError(t, err)
	assert.NotEmpty(t, newAccess)
	assert.NotEmpty(t, newRefresh)
	assert.NotEqual(t, refreshToken, newRefresh)

	// Old refresh token should be revoked and fail
	_, _, err = authSvc.RefreshSession(refreshToken)
	assert.Error(t, err)

	// 4. Logout
	err = authSvc.Logout(newRefresh)
	require.NoError(t, err)

	// Revoked refresh token should fail to refresh
	_, _, err = authSvc.RefreshSession(newRefresh)
	assert.Error(t, err)
}

func TestAuthService_RevokeAllSessions(t *testing.T) {
	db := setupTestDB(t)
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	userRepo := repository.NewUserRepository(db)
	orgSvc := NewOrgService(db)
	noopEmail := email.NoopSender{}
	auditSvc := NewAuditService(db)

	authSvc := NewAuthService(userRepo, orgSvc, auditSvc, "super-secret-key-that-is-long-enough", db, rClient, noopEmail, "http://localhost:8080")

	emailAddr := "user@example.com"
	password := "password123"
	err = authSvc.Register(emailAddr, password)
	require.NoError(t, err)

	// Login twice (simulate two devices)
	_, refresh1, err := authSvc.Login(emailAddr, password)
	require.NoError(t, err)
	_, refresh2, err := authSvc.Login(emailAddr, password)
	require.NoError(t, err)

	// Get user
	user, err := userRepo.FindByEmail(emailAddr)
	require.NoError(t, err)

	// Revoke all
	err = authSvc.RevokeAllSessions(user.ID)
	require.NoError(t, err)

	// Both should fail
	_, _, err = authSvc.RefreshSession(refresh1)
	assert.Error(t, err)
	_, _, err = authSvc.RefreshSession(refresh2)
	assert.Error(t, err)
}

func TestAuthService_EmailVerification(t *testing.T) {
	db := setupTestDB(t)
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	userRepo := repository.NewUserRepository(db)
	orgSvc := NewOrgService(db)
	noopEmail := email.NoopSender{}
	auditSvc := NewAuditService(db)

	authSvc := NewAuthService(userRepo, orgSvc, auditSvc, "super-secret-key-that-is-long-enough", db, rClient, noopEmail, "http://localhost:8080")

	emailAddr := "verify@example.com"
	password := "password123"
	err = authSvc.Register(emailAddr, password)
	require.NoError(t, err)

	user, err := userRepo.FindByEmail(emailAddr)
	require.NoError(t, err)
	assert.False(t, user.EmailVerified)

	// We need to extract the verification token from redis.
	// Keys are: verify_token:<token> -> userID
	ctx := context.Background()
	var keys []string
	assert.Eventually(t, func() bool {
		var err error
		keys, err = rClient.Keys(ctx, "verify_token:*").Result()
		return err == nil && len(keys) == 1
	}, 2*time.Second, 50*time.Millisecond)

	require.Len(t, keys, 1)

	// verify_token:<token>
	token := keys[0][len("verify_token:"):]

	// Verify email
	err = authSvc.VerifyEmail(token)
	require.NoError(t, err)

	// Check if verified
	db.First(&user, user.ID)
	assert.True(t, user.EmailVerified)

	// Try to verify again with same token (should fail since deleted)
	err = authSvc.VerifyEmail(token)
	assert.Error(t, err)
}
