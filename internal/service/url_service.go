package service

import (
	"github.com/Scarage1/url-shortener/internal/model"
	"github.com/Scarage1/url-shortener/internal/repository"
	"github.com/Scarage1/url-shortener/internal/utils"
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

func (s *URLService) CreateShortURL(originalURL string) (*model.URL, error) {

	shortCode, err := utils.GenerateShortCode(6)

	if err != nil {
		return nil, err
	}


	url := &model.URL{
		ShortCode:   shortCode,
		OriginalURL: originalURL,
	}


	err = s.Repo.Create(url)

	if err != nil {
		return nil, err
	}


	return url, nil
}
