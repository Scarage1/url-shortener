package routing

import (
	"testing"
	"time"

	"github.com/Scarage1/url-shortener/internal/model"
	"github.com/Scarage1/url-shortener/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func TestEngineResolve_ValidatesRuleConfig(t *testing.T) {
	engine := NewEngine()
	url := &model.URL{
		ShortCode:   "abc123",
		OriginalURL: "https://example.com",
		Rules: []model.RoutingRule{
			{
				Type:   model.RoutingRuleTypeGeo,
				Config: []byte(`{"US":"https://us.example.com"}`),
			},
			{
				Type:   model.RoutingRuleTypeSchedule,
				Config: []byte(`{"active_from":"2026-07-10T00:00:00Z","expires_at":"2026-08-10T00:00:00Z"}`),
			},
		},
	}

	result, err := engine.Resolve(
		url,
		Context{Now: time.Now()},
	)

	require.NoError(t, err)
	assert.Equal(t, "https://example.com", result)
}

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
