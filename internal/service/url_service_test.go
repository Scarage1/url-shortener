package service

import (
	"testing"
	"time"

	"github.com/Scarage1/url-shortener/internal/model"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// In-memory mock repository
// ---------------------------------------------------------------------------

type mockURLRepo struct {
	urls []*model.URL
}

func (m *mockURLRepo) Create(url *model.URL) error {
	m.urls = append(m.urls, url)
	return nil
}

func (m *mockURLRepo) FindByOriginalURL(originalURL string, userID uint) (*model.URL, error) {
	for _, u := range m.urls {
		if u.OriginalURL == originalURL && u.UserID == userID {
			return u, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockURLRepo) FindByShortCode(code string) (*model.URL, error) {
	for _, u := range m.urls {
		if u.ShortCode == code {
			return u, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockURLRepo) FindByCodeAndUser(code string, userID uint) (*model.URL, error) {
	for _, u := range m.urls {
		if u.ShortCode == code && u.UserID == userID {
			return u, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockURLRepo) FindByUser(userID uint) ([]model.URL, error) {
	var result []model.URL
	for _, u := range m.urls {
		if u.UserID == userID {
			result = append(result, *u)
		}
	}
	return result, nil
}

func (m *mockURLRepo) Update(url *model.URL) error { return nil }

func (m *mockURLRepo) IncrementClickCount(
	code string,
	delta int,
	accessedAt time.Time,
) error {
	if delta <= 0 {
		return nil
	}

	for _, u := range m.urls {
		if u.ShortCode == code {
			u.ClickCount += delta
			u.LastAccessed = &accessedAt
			return nil
		}
	}

	return gorm.ErrRecordNotFound
}

func (m *mockURLRepo) DeleteByCodeAndUser(code string, userID uint) error {
	for i, u := range m.urls {
		if u.ShortCode == code && u.UserID == userID {
			m.urls = append(m.urls[:i], m.urls[i+1:]...)
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func newTestService(t *testing.T, urls []*model.URL) (*URLService, *miniredis.Miniredis) {
	t.Helper()

	mr, err := miniredis.Run()
	require.NoError(t, err, "failed to start miniredis")

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	svc := NewURLService(&mockURLRepo{urls: urls}, client)

	return svc, mr
}

// ---------------------------------------------------------------------------
// URL ownership tests (multi-tenancy isolation)
// ---------------------------------------------------------------------------

func TestGetStats_OwnerCanAccessOwnURL(t *testing.T) {
	ownerID := uint(1)
	url := &model.URL{
		Model:       gorm.Model{ID: 1},
		ShortCode:   "abc123",
		OriginalURL: "https://github.com/openai",
		UserID:      ownerID,
		ClickCount:  5,
	}

	svc, mr := newTestService(t, []*model.URL{url})
	defer mr.Close()

	result, err := svc.GetStats("abc123", ownerID)

	require.NoError(t, err)
	assert.Equal(t, "abc123", result.ShortCode)
	assert.Equal(t, "https://github.com/openai", result.OriginalURL)
}

// Critical SaaS test: User B must NOT be able to read User A's stats.
func TestGetStats_OtherUserIsBlocked(t *testing.T) {
	ownerID := uint(1)
	attackerID := uint(2)

	url := &model.URL{
		Model:       gorm.Model{ID: 1},
		ShortCode:   "abc123",
		OriginalURL: "https://github.com/openai",
		UserID:      ownerID,
		ClickCount:  5,
	}

	svc, mr := newTestService(t, []*model.URL{url})
	defer mr.Close()

	_, err := svc.GetStats("abc123", attackerID)

	assert.Error(t, err, "User B must not access User A's URL stats")
}

// ---------------------------------------------------------------------------
// Delete ownership tests
// ---------------------------------------------------------------------------

func TestDeleteURL_OwnerCanDeleteOwnURL(t *testing.T) {
	ownerID := uint(1)
	url := &model.URL{
		Model:     gorm.Model{ID: 1},
		ShortCode: "del123",
		UserID:    ownerID,
	}

	svc, mr := newTestService(t, []*model.URL{url})
	defer mr.Close()

	err := svc.DeleteURL("del123", ownerID)
	assert.NoError(t, err)
}

func TestDeleteURL_OtherUserCannotDelete(t *testing.T) {
	ownerID := uint(1)
	attackerID := uint(2)

	url := &model.URL{
		Model:     gorm.Model{ID: 1},
		ShortCode: "del123",
		UserID:    ownerID,
	}

	svc, mr := newTestService(t, []*model.URL{url})
	defer mr.Close()

	err := svc.DeleteURL("del123", attackerID)
	assert.Error(t, err, "User B must not delete User A's URL")
}

// ---------------------------------------------------------------------------
// Deduplication test
// ---------------------------------------------------------------------------

func TestCreateShortURL_ReturnsSameCodeForDuplicateURL(t *testing.T) {
	ownerID := uint(1)
	original := "https://example.com"

	existing := &model.URL{
		Model:       gorm.Model{ID: 1},
		ShortCode:   "exist1",
		OriginalURL: original,
		UserID:      ownerID,
	}

	svc, mr := newTestService(t, []*model.URL{existing})
	defer mr.Close()

	result, err := svc.CreateShortURL(original, ownerID)

	require.NoError(t, err)
	assert.Equal(t, "exist1", result.ShortCode, "duplicate URL should return existing short code")
}

// ---------------------------------------------------------------------------
// Click counter merge test
// ---------------------------------------------------------------------------

func TestGetStats_ClickCountMergesRedisAndDB(t *testing.T) {
	ownerID := uint(1)
	url := &model.URL{
		Model:      gorm.Model{ID: 1},
		ShortCode:  "click1",
		UserID:     ownerID,
		ClickCount: 10, // DB baseline
	}

	svc, mr := newTestService(t, []*model.URL{url})
	defer mr.Close()

	// Simulate 5 clicks stored in Redis
	mr.Set("clicks:click1", "5")

	result, err := svc.GetStats("click1", ownerID)

	require.NoError(t, err)
	assert.Equal(t, 15, result.ClickCount, "total clicks should be DB baseline + Redis delta")
}
