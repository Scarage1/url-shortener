package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/Scarage1/url-shortener/internal/email"
	"github.com/Scarage1/url-shortener/internal/model"
	"github.com/Scarage1/url-shortener/internal/repository"
	"github.com/Scarage1/url-shortener/internal/utils"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// UserProfile is the response for GET /me.
type UserProfile struct {
	Email    string  `json:"email"`
	Verified bool    `json:"verified"`
	Org      OrgInfo `json:"organization"`
	Plan     string  `json:"plan"`
}

type OrgInfo struct {
	Name string `json:"name"`
	Role string `json:"role"`
}

type AuthService struct {
	UserRepo   *repository.UserRepository
	OrgService *OrgService
	JWTSecret  string
	FreePlanID uint
	DB         *gorm.DB
	Redis      *redis.Client
	Email      email.Sender
	BaseURL    string
}

func NewAuthService(
	repo *repository.UserRepository,
	orgService *OrgService,
	jwtSecret string,
	db *gorm.DB,
	redisClient *redis.Client,
	emailSender email.Sender,
	baseURL string,
) *AuthService {

	var freePlan model.Plan
	if err := db.Where("name = ?", model.PlanFree).First(&freePlan).Error; err != nil {
		freePlan.ID = 0
	}

	return &AuthService{
		UserRepo:   repo,
		OrgService: orgService,
		JWTSecret:  jwtSecret,
		FreePlanID: freePlan.ID,
		DB:         db,
		Redis:      redisClient,
		Email:      emailSender,
		BaseURL:    baseURL,
	}
}

func (s *AuthService) Register(
	email string,
	password string,
) error {

	if s.FreePlanID == 0 {
		return fmt.Errorf("free plan not configured")
	}

	hash, err := utils.HashPassword(password)
	if err != nil {
		return err
	}

	user := &model.User{
		Email:         email,
		PasswordHash:  hash,
		EmailVerified: false,
	}

	err = s.UserRepo.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		return s.OrgService.CreateDefaultOrg(tx, user, s.FreePlanID)
	})

	if err != nil {
		return errors.New("registration failed: " + err.Error())
	}

	// Send verification email (fire and forget — don't block registration)
	go s.sendVerificationEmail(user.ID, email)

	return nil
}

func (s *AuthService) Login(
	email string,
	password string,
) (string, error) {

	user, err := s.UserRepo.FindByEmail(email)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	if !utils.CheckPassword(password, user.PasswordHash) {
		return "", errors.New("invalid credentials")
	}

	return utils.GenerateToken(user.ID, s.JWTSecret)
}

// ---------------------------------------------------------------------------
// GET /me — User profile
// ---------------------------------------------------------------------------

func (s *AuthService) GetProfile(userID uint) (*UserProfile, error) {

	var user model.User
	if err := s.DB.First(&user, userID).Error; err != nil {
		return nil, errors.New("user not found")
	}

	var member model.OrganizationMember
	if err := s.DB.Where("user_id = ?", userID).First(&member).Error; err != nil {
		return nil, errors.New("organization not found")
	}

	var org model.Organization
	if err := s.DB.First(&org, member.OrganizationID).Error; err != nil {
		return nil, errors.New("organization not found")
	}

	var sub model.Subscription
	planName := "free"
	if err := s.DB.Where(
		"organization_id = ? AND status = ?",
		org.ID,
		model.SubscriptionActive,
	).Preload("Plan").First(&sub).Error; err == nil {
		planName = sub.Plan.Name
	}

	return &UserProfile{
		Email:    user.Email,
		Verified: user.EmailVerified,
		Org: OrgInfo{
			Name: org.Name,
			Role: member.Role,
		},
		Plan: planName,
	}, nil
}

// ---------------------------------------------------------------------------
// Email verification
// ---------------------------------------------------------------------------

func (s *AuthService) sendVerificationEmail(userID uint, toEmail string) {

	token, err := s.generateToken("verify", userID, 24*time.Hour)
	if err != nil {
		return
	}

	link := fmt.Sprintf("%s/api/v1/auth/verify?token=%s", s.BaseURL, token)

	body := fmt.Sprintf(`
		<h2>Verify your email</h2>
		<p>Click the link below to verify your email address:</p>
		<p><a href="%s">Verify Email</a></p>
		<p>This link expires in 24 hours.</p>
	`, link)

	_ = s.Email.Send(toEmail, "Verify your email", body)
}

func (s *AuthService) ResendVerification(userID uint) error {

	var user model.User
	if err := s.DB.First(&user, userID).Error; err != nil {
		return errors.New("user not found")
	}

	if user.EmailVerified {
		return errors.New("email already verified")
	}

	go s.sendVerificationEmail(userID, user.Email)
	return nil
}

func (s *AuthService) VerifyEmail(token string) error {

	userID, err := s.validateToken("verify", token)
	if err != nil {
		return err
	}

	return s.DB.Model(&model.User{}).
		Where("id = ?", userID).
		Update("email_verified", true).Error
}

// ---------------------------------------------------------------------------
// Forgot / reset password
// ---------------------------------------------------------------------------

func (s *AuthService) ForgotPassword(emailAddr string) error {

	user, err := s.UserRepo.FindByEmail(emailAddr)
	if err != nil {
		// Don't reveal whether the email exists
		return nil
	}

	token, err := s.generateToken("reset", user.ID, time.Hour)
	if err != nil {
		return nil
	}

	link := fmt.Sprintf("%s/reset-password?token=%s", s.BaseURL, token)

	body := fmt.Sprintf(`
		<h2>Reset your password</h2>
		<p>Click the link below to reset your password:</p>
		<p><a href="%s">Reset Password</a></p>
		<p>This link expires in 1 hour.</p>
		<p>If you didn't request this, ignore this email.</p>
	`, link)

	go func() {
		_ = s.Email.Send(emailAddr, "Reset your password", body)
	}()

	return nil
}

func (s *AuthService) ResetPassword(token string, newPassword string) error {

	userID, err := s.validateToken("reset", token)
	if err != nil {
		return err
	}

	hash, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	return s.DB.Model(&model.User{}).
		Where("id = ?", userID).
		Update("password_hash", hash).Error
}

// ---------------------------------------------------------------------------
// Token helpers (Redis-backed, one-time-use)
// ---------------------------------------------------------------------------

func (s *AuthService) generateToken(prefix string, userID uint, ttl time.Duration) (string, error) {

	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	token := hex.EncodeToString(bytes)
	key := fmt.Sprintf("%s_token:%s", prefix, token)

	ctx := context.Background()
	err := s.Redis.Set(ctx, key, userID, ttl).Err()
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *AuthService) validateToken(prefix string, token string) (uint, error) {

	key := fmt.Sprintf("%s_token:%s", prefix, token)
	ctx := context.Background()

	userID, err := s.Redis.Get(ctx, key).Uint64()
	if err != nil {
		return 0, errors.New("invalid or expired token")
	}

	// One-time use — delete after validation
	s.Redis.Del(ctx, key)

	return uint(userID), nil
}
