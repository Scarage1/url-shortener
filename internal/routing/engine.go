package routing

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Scarage1/url-shortener/internal/model"
	"github.com/Scarage1/url-shortener/internal/utils"
)

var ErrPasswordRequired = errors.New("password required")
var ErrInvalidPassword = errors.New("invalid password")
var ErrNotYetActive = errors.New("link not yet active")
var ErrExpired = errors.New("link has expired")

type Context struct {
	Now      time.Time
	Country  string
	Password string
}

type Engine struct{}

func NewEngine() *Engine {
	return &Engine{}
}

// Resolve evaluates all routing rules for a URL and returns the final destination.
// Rules are applied in order; a rule may change the destination (geo) or block
// the request entirely (password, schedule).
func (e *Engine) Resolve(url *model.URL, ctx Context) (string, error) {

	destination := url.OriginalURL

	for _, rule := range url.Rules {

		redirect, err := e.applyRule(rule, ctx)

		if err != nil {
			return "", err
		}

		// A non-empty redirect from a rule (e.g. geo) overrides the destination.
		if redirect != "" {
			destination = redirect
		}
	}

	return destination, nil
}

type PasswordRule struct {
	Hash string `json:"hash"`
}

type ScheduleRule struct {
	ActiveFrom *time.Time `json:"active_from"`
	ExpiresAt  *time.Time `json:"expires_at"`
}

// applyRule returns ("", nil) for pass-through rules, (newURL, nil) for
// redirect rules (geo), or ("", err) when the request must be blocked.
func (e *Engine) applyRule(
	rule model.RoutingRule,
	ctx Context,
) (string, error) {

	switch rule.Type {

	case model.RoutingRuleTypePassword:
		var cfg PasswordRule
		if err := decodeRule(rule.Config, &cfg); err != nil {
			return "", err
		}
		if cfg.Hash == "" {
			return "", fmt.Errorf("password rule requires a hash")
		}
		if ctx.Password == "" {
			return "", ErrPasswordRequired
		}
		if !utils.CheckPassword(ctx.Password, cfg.Hash) {
			return "", ErrInvalidPassword
		}
		return "", nil

	case model.RoutingRuleTypeSchedule:
		var cfg ScheduleRule
		if err := decodeRule(rule.Config, &cfg); err != nil {
			return "", err
		}
		if cfg.ActiveFrom != nil && ctx.Now.Before(*cfg.ActiveFrom) {
			return "", ErrNotYetActive
		}
		if cfg.ExpiresAt != nil && !ctx.Now.Before(*cfg.ExpiresAt) {
			return "", ErrExpired
		}
		return "", nil

	case model.RoutingRuleTypeGeo:
		var destinations map[string]string
		if err := decodeRule(rule.Config, &destinations); err != nil {
			return "", err
		}
		if len(destinations) == 0 {
			return "", fmt.Errorf("geo rule requires at least one destination")
		}
		// Route to country-specific destination if available; fall back silently.
		if ctx.Country != "" {
			if dest, ok := destinations[ctx.Country]; ok {
				return dest, nil
			}
		}
		return "", nil

	default:
		return "", fmt.Errorf("unsupported routing rule type: %s", rule.Type)
	}
}

func decodeRule(raw []byte, target interface{}) error {

	if len(raw) == 0 {
		raw = []byte("{}")
	}

	if err := json.Unmarshal(raw, target); err != nil {
		return fmt.Errorf("decode routing rule config: %w", err)
	}

	return nil
}
