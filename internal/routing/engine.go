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

type Context struct {
	Now      time.Time
	Country  string
	Password string
}

type Engine struct{}

func NewEngine() *Engine {

	return &Engine{}
}

func (e *Engine) Resolve(url *model.URL, ctx Context) (string, error) {

	for _, rule := range url.Rules {
		if err := e.applyRule(rule, ctx); err != nil {
			return "", err
		}
	}

	return url.OriginalURL, nil
}

type PasswordRule struct {
	Hash string `json:"hash"`
}

type ScheduleRule struct {
	ActiveFrom *time.Time `json:"active_from"`
	ExpiresAt  *time.Time `json:"expires_at"`
}

type GeoRule struct {
	Destinations map[string]string `json:"destinations"`
}

func (e *Engine) applyRule(
	rule model.RoutingRule,
	ctx Context,
) error {

	switch rule.Type {
	case model.RoutingRuleTypePassword:
		var cfg PasswordRule
		if err := decodeRule(rule.Config, &cfg); err != nil {
			return err
		}

		if cfg.Hash == "" {
			return fmt.Errorf("password rule requires a hash")
		}

		if ctx.Password == "" {
			return ErrPasswordRequired
		}

		if !utils.CheckPassword(ctx.Password, cfg.Hash) {
			return ErrInvalidPassword
		}

		return nil
	case model.RoutingRuleTypeSchedule:
		var cfg ScheduleRule
		return decodeRule(rule.Config, &cfg)
	case model.RoutingRuleTypeGeo:
		var raw map[string]string
		if err := decodeRule(rule.Config, &raw); err != nil {
			return err
		}

		geo := GeoRule{
			Destinations: raw,
		}

		if len(geo.Destinations) == 0 {
			return fmt.Errorf("geo rule requires at least one destination")
		}

		return nil
	default:
		return fmt.Errorf("unsupported routing rule type: %s", rule.Type)
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
