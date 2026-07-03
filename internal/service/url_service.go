package service

import (
	"context"
	"github.com/Scarage1/url-shortener/internal/model"
	"github.com/Scarage1/url-shortener/internal/repository"
	"github.com/Scarage1/url-shortener/internal/utils"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"log"
	"time"
)

type URLService struct {
	Repo  *repository.URLRepository
	Redis *redis.Client
}

func NewURLService(repo *repository.URLRepository, redis *redis.Client) *URLService {

	return &URLService{
		Repo:  repo,
		Redis: redis,
	}
}

func (s *URLService) GetOriginalURL(
	shortCode string,
) (*model.URL, error) {

	ctx := context.Background()

	cacheKey := "url:" + shortCode
	clickKey := "clicks:" + shortCode

	// 1. Check Redis cache
	cachedURL, err := s.Redis.Get(
		ctx,
		cacheKey,
	).Result()

	if err == nil {

		if err := s.Redis.Incr(
			ctx,
			clickKey,
		).Err(); err != nil {

			log.Println(
				"Redis counter error:",
				err,
			)
		}

		return &model.URL{
			ShortCode:   shortCode,
			OriginalURL: cachedURL,
		}, nil
	}

	// 2. Cache miss → PostgreSQL
	url, err := s.Repo.FindByShortCode(
		shortCode,
	)

	if err != nil {
		return nil, err
	}

	// 3. Store in Redis
	if err := s.Redis.Set(
		ctx,
		cacheKey,
		url.OriginalURL,
		time.Hour,
	).Err(); err != nil {

		log.Println(
			"Redis cache error:",
			err,
		)
	}

	// 4. Increment Redis counter
	if err := s.Redis.Incr(
		ctx,
		clickKey,
	).Err(); err != nil {

		log.Println(
			"Redis counter error:",
			err,
		)
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

func (s *URLService) GetStats(
	shortCode string,
) (*model.URL, error) {

	return s.Repo.FindByShortCode(
		shortCode,
	)
}
