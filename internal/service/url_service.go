package service

import (
	"github.com/Scarage1/url-shortener/internal/model"
	"github.com/Scarage1/url-shortener/internal/repository"
	"github.com/Scarage1/url-shortener/internal/utils"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"time"
	"context"
)


type URLService struct {
	Repo *repository.URLRepository
	Redis *redis.Client
}


func NewURLService(repo *repository.URLRepository, redis *redis.Client) *URLService {

	return &URLService{
		Repo: repo,
		Redis: redis,
	}
}

func (s *URLService) GetOriginalURL(
	shortCode string,
) (*model.URL,error){


	cachedURL,err :=
		s.Redis.Get(
			context.Background(),
			shortCode,
		).Result()


	if err == nil {


		return &model.URL{
			ShortCode: shortCode,
			OriginalURL: cachedURL,
		},nil
	}


	url,err :=
		s.Repo.FindByShortCode(
			shortCode,
		)


	if err != nil {

		return nil,err
	}


	s.Redis.Set(
		context.Background(),
		shortCode,
		url.OriginalURL,
		time.Hour,
	)


	now := time.Now()

	url.ClickCount++

	url.LastAccessed=&now


	s.Repo.Update(url)


	return url,nil
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

func (s *URLService) GetStats(
	shortCode string,
) (*model.URL,error){

	return s.Repo.FindByShortCode(
		shortCode,
	)
}