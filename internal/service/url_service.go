package service

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/Scarage1/url-shortener/internal/model"
	"github.com/Scarage1/url-shortener/internal/routing"
	"github.com/Scarage1/url-shortener/internal/security"
	"github.com/Scarage1/url-shortener/internal/utils"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var ErrUnsafeURL = errors.New("unsafe URL")

// urlRepository is the minimal interface the service needs from the data layer.
// *repository.URLRepository satisfies it automatically — no changes needed in callers.
type urlRepository interface {
	Create(url *model.URL) error
	FindByOriginalURL(originalURL string, userID uint) (*model.URL, error)
	FindByShortCode(code string) (*model.URL, error)
	FindByCodeAndUser(code string, userID uint) (*model.URL, error)
	FindByUser(userID uint) ([]model.URL, error)
	Update(url *model.URL) error
	DeleteByCodeAndUser(code string, userID uint) error
	IncrementClickCount(code string, delta int, accessedAt time.Time) error
}

type URLService struct {
	Repo     urlRepository
	Redis    *redis.Client
	Scanner  security.URLScanner
	Resolver *routing.Engine
}

func NewURLService(
	repo urlRepository,
	redis *redis.Client,
	scanner security.URLScanner,
	resolver *routing.Engine,
) *URLService {

	return &URLService{
		Repo:     repo,
		Redis:    redis,
		Scanner:  scanner,
		Resolver: resolver,
	}
}

func (s *URLService) GetOriginalURL(
	shortCode string,
) (*model.URL, error) {

	ctx := context.Background()

	cacheKey := "url:" + shortCode
	clickKey := "clicks:" + shortCode
	clickRecorded := false

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
		clickRecorded = true

		if s.Resolver == nil {
			return &model.URL{
				ShortCode:   shortCode,
				OriginalURL: cachedURL,
			}, nil
		}
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
	if !clickRecorded {
		if err := s.Redis.Incr(
			ctx,
			clickKey,
		).Err(); err != nil {

			log.Println(
				"Redis counter error:",
				err,
			)
		}
	}

	destination := url.OriginalURL

	if s.Resolver != nil {
		resolvedURL, err := s.Resolver.Resolve(
			url,
			routing.Context{
				Now: time.Now(),
			},
		)
		if err != nil {
			return nil, err
		}
		destination = resolvedURL
	}

	return &model.URL{
		ShortCode:   url.ShortCode,
		OriginalURL: destination,
		Rules:       url.Rules,
	}, nil
}

func (s *URLService) FlushClickCounts(ctx context.Context) error {

	var cursor uint64
	now := time.Now()

	for {
		keys, nextCursor, err := s.Redis.Scan(
			ctx,
			cursor,
			"clicks:*",
			100,
		).Result()

		if err != nil {
			return err
		}

		for _, key := range keys {

			code := strings.TrimPrefix(key, "clicks:")

			val, err := s.Redis.GetDel(ctx, key).Int()

			if err != nil {
				if err != redis.Nil {
					log.Println("flush: getdel error:", err)
				}
				continue
			}

			if val <= 0 {
				continue
			}

			if err := s.Repo.IncrementClickCount(code, val, now); err != nil {
				log.Println("flush: db increment error for", code, ":", err)
				s.Redis.IncrBy(ctx, key, int64(val))
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return nil
}

func (s *URLService) CreateShortURL(
	originalURL string,
	userID uint,
) (*model.URL, error) {

	if s.Scanner != nil {
		err := s.Scanner.Check(originalURL)
		if err != nil {
			return nil, ErrUnsafeURL
		}
	}

	existingURL, err :=
		s.Repo.FindByOriginalURL(
			originalURL,
			userID,
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

		UserID: userID,
	}

	err = s.Repo.Create(url)

	if err != nil {

		return nil, err
	}

	return url, nil
}

// GetStats returns URL analytics with live click count merged from Redis.
// The DB stores the persisted baseline; Redis holds the live delta since
// the last page reload or server restart.
func (s *URLService) GetStats(
	code string,
	userID uint,
) (*model.URL, error) {

	url, err := s.Repo.FindByCodeAndUser(
		code,
		userID,
	)

	if err != nil {
		return nil, err
	}

	// Merge live Redis click counter on top of DB baseline
	ctx := context.Background()
	clickKey := "clicks:" + code

	redisCount, err := s.Redis.Get(ctx, clickKey).Int()

	if err == nil {
		// Redis has a live delta — add it to the DB value
		url.ClickCount += redisCount
	}
	// If Redis key is missing (expired/not yet set), fall back to DB value only

	return url, nil
}

// GetUserLinks returns all URLs created by a user, ordered newest first.
func (s *URLService) GetUserLinks(
	userID uint,
) ([]model.URL, error) {

	return s.Repo.FindByUser(
		userID,
	)
}

// DeleteURL removes a URL by short code, enforcing ownership.
// Also clears the Redis cache and click counter for the code.
func (s *URLService) DeleteURL(
	code string,
	userID uint,
) error {

	err := s.Repo.DeleteByCodeAndUser(
		code,
		userID,
	)

	if err != nil {
		return err
	}

	// Evict Redis cache so stale entries don't redirect to deleted URLs
	ctx := context.Background()

	s.Redis.Del(ctx, "url:"+code)
	s.Redis.Del(ctx, "clicks:"+code)

	return nil
}
