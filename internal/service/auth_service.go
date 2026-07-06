package service

import (
	"errors"
	"fmt"

	"github.com/Scarage1/url-shortener/internal/model"
	"github.com/Scarage1/url-shortener/internal/repository"
	"github.com/Scarage1/url-shortener/internal/utils"
	"gorm.io/gorm"
)

type AuthService struct {
	UserRepo   *repository.UserRepository
	OrgService *OrgService
	JWTSecret  string
	FreePlanID uint // cached on startup
}

func NewAuthService(
	repo *repository.UserRepository,
	orgService *OrgService,
	jwtSecret string,
	db *gorm.DB,
) *AuthService {

	// Look up the free plan ID once at startup
	var freePlan model.Plan
	if err := db.Where("name = ?", model.PlanFree).First(&freePlan).Error; err != nil {
		// Will be 0 if plans aren't seeded yet — register will fail gracefully
		freePlan.ID = 0
	}

	return &AuthService{
		UserRepo:   repo,
		OrgService: orgService,
		JWTSecret:  jwtSecret,
		FreePlanID: freePlan.ID,
	}
}

func (s *AuthService) Register(
	email string,
	password string,
) error {

	if s.FreePlanID == 0 {
		return fmt.Errorf("free plan not configured")
	}

	hash, err :=
		utils.HashPassword(
			password,
		)

	if err != nil {

		return err
	}

	user :=
		&model.User{

			Email: email,

			PasswordHash: hash,

			EmailVerified: false,
		}

	// Use a transaction: create user + org + membership + subscription
	err = s.UserRepo.DB.Transaction(func(tx *gorm.DB) error {

		if err := tx.Create(user).Error; err != nil {
			return err
		}

		return s.OrgService.CreateDefaultOrg(tx, user, s.FreePlanID)
	})

	if err != nil {
		return errors.New("registration failed: " + err.Error())
	}

	return nil
}

func (s *AuthService) Login(
	email string,
	password string,
) (string, error) {

	user, err :=
		s.UserRepo.FindByEmail(
			email,
		)

	if err != nil {

		return "",
			errors.New(
				"invalid credentials",
			)
	}

	if !utils.CheckPassword(
		password,
		user.PasswordHash,
	) {

		return "",
			errors.New(
				"invalid credentials",
			)
	}

	return utils.GenerateToken(
		user.ID,
		s.JWTSecret,
	)
}
