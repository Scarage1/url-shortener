package service

import (
	"github.com/Scarage1/url-shortener/internal/model"
	"github.com/Scarage1/url-shortener/internal/repository"
	"github.com/Scarage1/url-shortener/internal/utils"
	"gorm.io/gorm"
)


type URLService struct {
	Repo *repository.URLRepository
}


func NewURLService(repo *repository.URLRepository) *URLService {

	return &URLService{
		Repo: repo,
	}
}

func (s *URLService) GetOriginalURL(shortCode string) (*model.URL, error) {

	url, err := s.Repo.FindByShortCode(shortCode)

	if err != nil {
		return nil, err
	}

	return url, nil
}

func (s *URLService) CreateShortURL(
	originalURL string,
) (*model.URL, error) {


	existingURL, err :=
		s.Repo.FindByOriginalURL(
			originalURL,
		)


	if err == nil {

		return existingURL, nil
	}


	if err != gorm.ErrRecordNotFound {

		return nil, err
	}


	var shortCode string


	for {


		code, err :=
			utils.GenerateShortCode(6)


		if err != nil {
			return nil, err
		}


		_, err =
			s.Repo.FindByShortCode(code)


		if err == gorm.ErrRecordNotFound {

			shortCode = code

			break
		}


		if err != nil {

			return nil, err
		}
	}


	url := &model.URL{

		ShortCode: shortCode,

		OriginalURL: originalURL,
	}


	err = s.Repo.Create(url)


	if err != nil {

		return nil, err
	}


	return url, nil
}