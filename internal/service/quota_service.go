package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Scarage1/url-shortener/internal/model"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var ErrLinkLimitReached = errors.New("link limit reached")
var ErrRedirectQuotaExceeded = errors.New("redirect quota exceeded")
var ErrAPIQuotaExceeded = errors.New("API quota exceeded")

// UsageStats contains current month counters for the dashboard.
type UsageStats struct {
	Redirects     int `json:"redirects"`
	RedirectLimit int `json:"redirect_limit"`
	APICalls      int `json:"api_calls"`
	APILimit      int `json:"api_limit"`
	ActiveLinks   int `json:"active_links"`
	LinkLimit     int `json:"link_limit"`
}

// QuotaService enforces plan limits and tracks usage counters.
type QuotaService struct {
	DB    *gorm.DB
	Redis *redis.Client
}

func NewQuotaService(db *gorm.DB, redis *redis.Client) *QuotaService {
	return &QuotaService{DB: db, Redis: redis}
}

// getPlan returns the plan for a given organization via its subscription.
func (q *QuotaService) getPlan(orgID uint) (*model.Plan, error) {

	var sub model.Subscription

	err := q.DB.Where(
		"organization_id = ? AND status = ?",
		orgID,
		model.SubscriptionActive,
	).Preload("Plan").First(&sub).Error

	if err != nil {
		return nil, fmt.Errorf("subscription not found for org %d: %w", orgID, err)
	}

	return &sub.Plan, nil
}

// CanCreateLink checks if the org can create another link.
func (q *QuotaService) CanCreateLink(orgID uint) error {

	plan, err := q.getPlan(orgID)
	if err != nil {
		return err
	}

	if plan.MaxLinks == -1 {
		return nil // unlimited
	}

	var count int64

	err = q.DB.Model(&model.URL{}).
		Where("organization_id = ?", orgID).
		Count(&count).Error

	if err != nil {
		return err
	}

	if int(count) >= plan.MaxLinks {
		return ErrLinkLimitReached
	}

	return nil
}

// TrackRedirect increments the monthly redirect counter.
// Returns true if the org is over their redirect quota (analytics limited).
// The redirect itself should ALWAYS proceed — never block end-user traffic.
func (q *QuotaService) TrackRedirect(orgID uint) (limited bool, err error) {

	plan, err := q.getPlan(orgID)
	if err != nil {
		// If we can't determine the plan, don't block traffic
		return false, nil
	}

	if plan.MaxRedirects == -1 {
		return false, nil // unlimited
	}

	ctx := context.Background()
	key := redirectKey(orgID)

	count, err := q.Redis.Incr(ctx, key).Result()
	if err != nil {
		return false, nil // fail open
	}

	// Set TTL on first increment (35 days covers the month + overlap)
	if count == 1 {
		q.Redis.Expire(ctx, key, 35*24*time.Hour)
	}

	return int(count) > plan.MaxRedirects, nil
}

// TrackAPICall increments the monthly API call counter.
func (q *QuotaService) TrackAPICall(orgID uint) error {

	plan, err := q.getPlan(orgID)
	if err != nil {
		return err
	}

	if plan.MaxAPICalls == -1 {
		return nil // unlimited
	}

	ctx := context.Background()
	key := apiKey(orgID)

	count, err := q.Redis.Incr(ctx, key).Result()
	if err != nil {
		return nil // fail open
	}

	if count == 1 {
		q.Redis.Expire(ctx, key, 35*24*time.Hour)
	}

	if int(count) > plan.MaxAPICalls {
		return ErrAPIQuotaExceeded
	}

	return nil
}

// GetUsage returns current month usage stats for the dashboard.
func (q *QuotaService) GetUsage(orgID uint) (*UsageStats, error) {

	plan, err := q.getPlan(orgID)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	redirects, _ := q.Redis.Get(ctx, redirectKey(orgID)).Int()
	apiCalls, _ := q.Redis.Get(ctx, apiKey(orgID)).Int()

	var linkCount int64
	q.DB.Model(&model.URL{}).
		Where("organization_id = ?", orgID).
		Count(&linkCount)

	return &UsageStats{
		Redirects:     redirects,
		RedirectLimit: plan.MaxRedirects,
		APICalls:      apiCalls,
		APILimit:      plan.MaxAPICalls,
		ActiveLinks:   int(linkCount),
		LinkLimit:     plan.MaxLinks,
	}, nil
}

// GetPlanForOrg returns the current plan for the dashboard.
func (q *QuotaService) GetPlanForOrg(orgID uint) (*model.Plan, error) {
	return q.getPlan(orgID)
}

func redirectKey(orgID uint) string {
	month := time.Now().Format("2006-01")
	return fmt.Sprintf("usage:redirect:%d:%s", orgID, month)
}

func apiKey(orgID uint) string {
	month := time.Now().Format("2006-01")
	return fmt.Sprintf("usage:api:%d:%s", orgID, month)
}
