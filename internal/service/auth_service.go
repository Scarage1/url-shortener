package service

import (
	"errors"

	"github.com/Scarage1/url-shortener/internal/model"
	"github.com/Scarage1/url-shortener/internal/repository"
	"github.com/Scarage1/url-shortener/internal/utils"
)

type AuthService struct {
	UserRepo *repository.UserRepository
}

func NewAuthService(
	repo *repository.UserRepository,
) *AuthService {

	return &AuthService{
		UserRepo: repo,
	}
}

func (s *AuthService) Register(
	email string,
	password string,
) error {

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
		}

	return s.UserRepo.Create(
		user,
	)
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
	)
}
