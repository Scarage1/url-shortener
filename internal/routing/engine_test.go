package routing

import (
	"testing"
	"time"

	"github.com/Scarage1/url-shortener/internal/model"
	"github.com/Scarage1/url-shortener/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func mustTime(s string) *time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return &t
}

// ---------------------------------------------------------------------------
// Basic
// ---------------------------------------------------------------------------

func TestEngineResolve_NoRulesReturnsOriginalURL(t *testing.T) {
	engine := NewEngine()
	url := &model.URL{
		ShortCode:   "abc123",
		OriginalURL: "https://example.com",
	}

	result, err := engine.Resolve(
		url,
		Context{Now: time.Now()},
	)

	require.NoError(t, err)
	assert.Equal(t, "https://example.com", result)
}

// ---------------------------------------------------------------------------
// Schedule rule
// ---------------------------------------------------------------------------

func TestEngineResolve_ScheduleRule_WithinWindow(t *testing.T) {
	engine := NewEngine()
	url := &model.URL{
		ShortCode:   "abc123",
		OriginalURL: "https://example.com",
		Rules: []model.RoutingRule{
			{
				Type: model.RoutingRuleTypeSchedule,
				// Active: started yesterday, expires tomorrow.
				Config: []byte(`{"active_from":"2020-01-01T00:00:00Z","expires_at":"2099-12-31T23:59:59Z"}`),
			},
		},
	}

	result, err := engine.Resolve(url, Context{Now: time.Now()})

	require.NoError(t, err)
	assert.Equal(t, "https://example.com", result)
}

func TestEngineResolve_ScheduleRule_NotYetActive(t *testing.T) {
	engine := NewEngine()
	url := &model.URL{
		ShortCode:   "abc123",
		OriginalURL: "https://example.com",
		Rules: []model.RoutingRule{
			{
				Type:   model.RoutingRuleTypeSchedule,
				Config: []byte(`{"active_from":"2099-01-01T00:00:00Z"}`),
			},
		},
	}

	_, err := engine.Resolve(url, Context{Now: time.Now()})

	assert.ErrorIs(t, err, ErrNotYetActive)
}

func TestEngineResolve_ScheduleRule_Expired(t *testing.T) {
	engine := NewEngine()
	url := &model.URL{
		ShortCode:   "abc123",
		OriginalURL: "https://example.com",
		Rules: []model.RoutingRule{
			{
				Type:   model.RoutingRuleTypeSchedule,
				Config: []byte(`{"expires_at":"2020-01-01T00:00:00Z"}`),
			},
		},
	}

	_, err := engine.Resolve(url, Context{Now: time.Now()})

	assert.ErrorIs(t, err, ErrExpired)
}

// ---------------------------------------------------------------------------
// Password rule
// ---------------------------------------------------------------------------

func TestEngineResolve_PasswordRuleRequiresPassword(t *testing.T) {
	engine := NewEngine()
	hash, err := utils.HashPassword("secret")
	require.NoError(t, err)

	url := &model.URL{
		ShortCode:   "abc123",
		OriginalURL: "https://example.com",
		Rules: []model.RoutingRule{
			{
				Type:   model.RoutingRuleTypePassword,
				Config: []byte(`{"hash":"` + hash + `"}`),
			},
		},
	}

	_, err = engine.Resolve(url, Context{Now: time.Now()})

	assert.ErrorIs(t, err, ErrPasswordRequired)
}

func TestEngineResolve_PasswordRuleAcceptsMatchingPassword(t *testing.T) {
	engine := NewEngine()
	hash, err := utils.HashPassword("secret")
	require.NoError(t, err)

	url := &model.URL{
		ShortCode:   "abc123",
		OriginalURL: "https://example.com",
		Rules: []model.RoutingRule{
			{
				Type:   model.RoutingRuleTypePassword,
				Config: []byte(`{"hash":"` + hash + `"}`),
			},
		},
	}

	result, err := engine.Resolve(
		url,
		Context{
			Now:      time.Now(),
			Password: "secret",
		},
	)

	require.NoError(t, err)
	assert.Equal(t, "https://example.com", result)
}

// ---------------------------------------------------------------------------
// Geo rule
// ---------------------------------------------------------------------------

func TestEngineResolve_GeoRule_RoutesToMatchingCountry(t *testing.T) {
	engine := NewEngine()
	url := &model.URL{
		ShortCode:   "abc123",
		OriginalURL: "https://example.com",
		Rules: []model.RoutingRule{
			{
				Type:   model.RoutingRuleTypeGeo,
				Config: []byte(`{"IN":"https://in.example.com","US":"https://us.example.com"}`),
			},
		},
	}

	result, err := engine.Resolve(url, Context{Now: time.Now(), Country: "IN"})

	require.NoError(t, err)
	assert.Equal(t, "https://in.example.com", result)
}

func TestEngineResolve_GeoRule_FallsBackToOriginalURL(t *testing.T) {
	engine := NewEngine()
	url := &model.URL{
		ShortCode:   "abc123",
		OriginalURL: "https://example.com",
		Rules: []model.RoutingRule{
			{
				Type:   model.RoutingRuleTypeGeo,
				Config: []byte(`{"US":"https://us.example.com"}`),
			},
		},
	}

	// Country "DE" has no destination — falls back to original URL.
	result, err := engine.Resolve(url, Context{Now: time.Now(), Country: "DE"})

	require.NoError(t, err)
	assert.Equal(t, "https://example.com", result)
}

func TestEngineResolve_GeoRule_NoCountryInContext(t *testing.T) {
	engine := NewEngine()
	url := &model.URL{
		ShortCode:   "abc123",
		OriginalURL: "https://example.com",
		Rules: []model.RoutingRule{
			{
				Type:   model.RoutingRuleTypeGeo,
				Config: []byte(`{"US":"https://us.example.com"}`),
			},
		},
	}

	// No country resolved (private IP, lookup failure) — fall back gracefully.
	result, err := engine.Resolve(url, Context{Now: time.Now(), Country: ""})

	require.NoError(t, err)
	assert.Equal(t, "https://example.com", result)
}
