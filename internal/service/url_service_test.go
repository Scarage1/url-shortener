package service

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/Scarage1/url-shortener/internal/geo"
	"github.com/Scarage1/url-shortener/internal/model"
	"github.com/Scarage1/url-shortener/internal/routing"
	"github.com/Scarage1/url-shortener/internal/security"
	"github.com/Scarage1/url-shortener/internal/utils"
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

type stubScanner struct {
	err error
}

func (s stubScanner) Check(string) error {
	return s.err
}

func (m *mockURLRepo) Create(url *model.URL) error {
	m.urls = append(m.urls, url)
	return nil
}

func (m *mockURLRepo) FindByOriginalURL(originalURL string, orgID uint) (*model.URL, error) {
	for _, u := range m.urls {
		if u.OriginalURL == originalURL && u.OrganizationID == orgID {
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

func (m *mockURLRepo) FindByCodeAndOrg(code string, orgID uint) (*model.URL, error) {
	for _, u := range m.urls {
		if u.ShortCode == code && u.OrganizationID == orgID {
			return u, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockURLRepo) FindByOrg(orgID uint) ([]model.URL, error) {
	var result []model.URL
	for _, u := range m.urls {
		if u.OrganizationID == orgID {
			result = append(result, *u)
		}
	}
	return result, nil
}

func (m *mockURLRepo) CountByOrg(orgID uint) (int64, error) {
	var count int64
	for _, u := range m.urls {
		if u.OrganizationID == orgID {
			count++
		}
	}
	return count, nil
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

func (m *mockURLRepo) DeleteByCodeAndOrg(code string, orgID uint) error {
	for i, u := range m.urls {
		if u.ShortCode == code && u.OrganizationID == orgID {
			m.urls = append(m.urls[:i], m.urls[i+1:]...)
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func newTestService(
	t *testing.T,
	urls []*model.URL,
	scanner security.URLScanner,
) (*URLService, *miniredis.Miniredis) {
	t.Helper()

	mr, err := miniredis.Run()
	require.NoError(t, err, "failed to start miniredis")

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	svc := NewURLService(
		&mockURLRepo{urls: urls},
		client,
		scanner,
		routing.NewEngine(),
		geo.NoopLocator{},
	)

	return svc, mr
}

// ---------------------------------------------------------------------------
// URL ownership tests (multi-tenancy isolation via org)
// ---------------------------------------------------------------------------

func TestGetStats_OwnerCanAccessOwnURL(t *testing.T) {
	url := &model.URL{
		Model:          gorm.Model{ID: 1},
		ShortCode:      "abc123",
		OriginalURL:    "https://example.com",
		OrganizationID: 1,
		CreatedBy:      1,
	}

	svc, mr := newTestService(t, []*model.URL{url}, security.AllowAllScanner{})
	defer mr.Close()

	result, err := svc.GetStats("abc123", 1)

	require.NoError(t, err)
	assert.Equal(t, "https://example.com", result.OriginalURL)
}

func TestGetStats_OtherOrgCannotAccessURL(t *testing.T) {
	url := &model.URL{
		Model:          gorm.Model{ID: 1},
		ShortCode:      "abc123",
		OriginalURL:    "https://example.com",
		OrganizationID: 1,
		CreatedBy:      1,
	}

	svc, mr := newTestService(t, []*model.URL{url}, security.AllowAllScanner{})
	defer mr.Close()

	// Org 2 tries to access org 1's URL
	_, err := svc.GetStats("abc123", 2)

	assert.Error(t, err)
}

func TestDeleteURL_OwnerCanDelete(t *testing.T) {
	url := &model.URL{
		Model:          gorm.Model{ID: 1},
		ShortCode:      "abc123",
		OriginalURL:    "https://example.com",
		OrganizationID: 1,
		CreatedBy:      1,
	}

	svc, mr := newTestService(t, []*model.URL{url}, security.AllowAllScanner{})
	defer mr.Close()

	err := svc.DeleteURL("abc123", 1)

	require.NoError(t, err)
}

func TestDeleteURL_OtherOrgCannotDelete(t *testing.T) {
	url := &model.URL{
		Model:          gorm.Model{ID: 1},
		ShortCode:      "abc123",
		OriginalURL:    "https://example.com",
		OrganizationID: 1,
		CreatedBy:      1,
	}

	svc, mr := newTestService(t, []*model.URL{url}, security.AllowAllScanner{})
	defer mr.Close()

	err := svc.DeleteURL("abc123", 2)

	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// Click counting
// ---------------------------------------------------------------------------

func TestGetStats_MergesRedisClickCount(t *testing.T) {
	url := &model.URL{
		Model:          gorm.Model{ID: 1},
		ShortCode:      "count1",
		OriginalURL:    "https://example.com",
		OrganizationID: 1,
		CreatedBy:      1,
		ClickCount:     10,
	}

	svc, mr := newTestService(t, []*model.URL{url}, security.AllowAllScanner{})
	defer mr.Close()

	mr.Set("clicks:count1", "5")

	result, err := svc.GetStats("count1", 1)

	require.NoError(t, err)
	assert.Equal(t, 15, result.ClickCount)
}

func TestGetStats_FallsBackToDBWhenNoRedisKey(t *testing.T) {
	url := &model.URL{
		Model:          gorm.Model{ID: 1},
		ShortCode:      "count2",
		OriginalURL:    "https://example.com",
		OrganizationID: 1,
		CreatedBy:      1,
		ClickCount:     42,
	}

	svc, mr := newTestService(t, []*model.URL{url}, security.AllowAllScanner{})
	defer mr.Close()

	result, err := svc.GetStats("count2", 1)

	require.NoError(t, err)
	assert.Equal(t, 42, result.ClickCount)
}

// ---------------------------------------------------------------------------
// URL Safety
// ---------------------------------------------------------------------------

func TestCreateShortURL_RejectsUnsafeURL(t *testing.T) {
	svc, mr := newTestService(
		t,
		nil,
		stubScanner{err: errors.New("flagged")},
	)
	defer mr.Close()

	_, err := svc.CreateShortURL("https://malware.example", 1, 1, nil)

	assert.ErrorIs(t, err, ErrUnsafeURL)
}

// ---------------------------------------------------------------------------
// Routing rules
// ---------------------------------------------------------------------------

func TestGetOriginalURL_LoadsRoutingRules(t *testing.T) {
	// Generate a real bcrypt hash so the engine can verify the password.
	hash, err := utils.HashPassword("secret")
	require.NoError(t, err)

	passwordConfig, err := json.Marshal(
		map[string]string{
			"hash": hash,
		},
	)
	require.NoError(t, err)

	url := &model.URL{
		Model:          gorm.Model{ID: 1},
		ShortCode:      "rule01",
		OriginalURL:    "https://example.com",
		OrganizationID: 1,
		CreatedBy:      1,
		Rules: []model.RoutingRule{
			{
				Type:   model.RoutingRuleTypePassword,
				Config: passwordConfig,
			},
		},
	}

	svc, mr := newTestService(t, []*model.URL{url}, security.AllowAllScanner{})
	defer mr.Close()

	// Correct password → redirect should succeed
	result, err := svc.GetOriginalURL("rule01", "secret", "")

	require.NoError(t, err)
	assert.Equal(t, "https://example.com", result.OriginalURL)
	assert.Len(t, result.Rules, 1)
	assert.Equal(t, model.RoutingRuleTypePassword, result.Rules[0].Type)
}

func TestGetOriginalURL_CacheHitStillLoadsRoutingRules(t *testing.T) {
	geoConfig, err := json.Marshal(
		map[string]string{
			"US": "https://us.example.com",
		},
	)
	require.NoError(t, err)

	url := &model.URL{
		Model:          gorm.Model{ID: 1},
		ShortCode:      "rule02",
		OriginalURL:    "https://example.com",
		OrganizationID: 1,
		CreatedBy:      1,
		Rules: []model.RoutingRule{
			{
				Type:   model.RoutingRuleTypeGeo,
				Config: geoConfig,
			},
		},
	}

	svc, mr := newTestService(t, []*model.URL{url}, security.AllowAllScanner{})
	defer mr.Close()

	_, err = svc.GetOriginalURL("rule02", "", "")
	require.NoError(t, err)

	// Warm the cache first, then fetch with the geo rule active
	mr.Set("url:rule02", "https://example.com")

	result, err := svc.GetOriginalURL("rule02", "", "")

	require.NoError(t, err)
	assert.Equal(t, "https://example.com", result.OriginalURL)
	assert.Len(t, result.Rules, 1)
	assert.Equal(t, model.RoutingRuleTypeGeo, result.Rules[0].Type)
}

func TestCreateShortURL_PasswordRuleHashesPassword(t *testing.T) {
	svc, mr := newTestService(t, nil, security.AllowAllScanner{})
	defer mr.Close()

	result, err := svc.CreateShortURL(
		"https://example.com/private",
		1, // orgID
		1, // createdBy
		[]CreateRuleInput{
			{
				Type:     model.RoutingRuleTypePassword,
				Password: "secret123",
			},
		},
	)

	require.NoError(t, err)
	require.Len(t, result.Rules, 1)
	assert.Equal(t, model.RoutingRuleTypePassword, result.Rules[0].Type)

	// Verify the config contains a bcrypt hash, not plaintext
	var cfg routing.PasswordRule
	err = json.Unmarshal(result.Rules[0].Config, &cfg)
	require.NoError(t, err)
	assert.NotEmpty(t, cfg.Hash)
	assert.True(t, utils.CheckPassword("secret123", cfg.Hash))
}

// ---------------------------------------------------------------------------
// Password-protected redirect
// ---------------------------------------------------------------------------

func TestGetOriginalURL_PasswordRuleBlocksWithoutPassword(t *testing.T) {
	hash, err := utils.HashPassword("secret")
	require.NoError(t, err)

	passwordConfig, err := json.Marshal(
		routing.PasswordRule{Hash: hash},
	)
	require.NoError(t, err)

	url := &model.URL{
		Model:          gorm.Model{ID: 1},
		ShortCode:      "lock01",
		OriginalURL:    "https://example.com/private",
		OrganizationID: 1,
		CreatedBy:      1,
		Rules: []model.RoutingRule{
			{
				Type:   model.RoutingRuleTypePassword,
				Config: passwordConfig,
			},
		},
	}

	svc, mr := newTestService(t, []*model.URL{url}, security.AllowAllScanner{})
	defer mr.Close()

	_, err = svc.GetOriginalURL("lock01", "", "")

	assert.ErrorIs(t, err, ErrPasswordRequired)
}

func TestGetOriginalURL_PasswordRuleAllowsMatchingPassword(t *testing.T) {
	hash, err := utils.HashPassword("secret")
	require.NoError(t, err)

	passwordConfig, err := json.Marshal(
		routing.PasswordRule{Hash: hash},
	)
	require.NoError(t, err)

	url := &model.URL{
		Model:          gorm.Model{ID: 1},
		ShortCode:      "lock02",
		OriginalURL:    "https://example.com/private",
		OrganizationID: 1,
		CreatedBy:      1,
		Rules: []model.RoutingRule{
			{
				Type:   model.RoutingRuleTypePassword,
				Config: passwordConfig,
			},
		},
	}

	svc, mr := newTestService(t, []*model.URL{url}, security.AllowAllScanner{})
	defer mr.Close()

	result, err := svc.GetOriginalURL("lock02", "secret", "")

	require.NoError(t, err)
	assert.Equal(t, "https://example.com/private", result.OriginalURL)
}
