package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Scarage1/url-shortener/internal/model"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// DashboardStats is the response for GET /api/v1/dashboard.
// One call replaces 5 separate frontend API calls.
type DashboardStats struct {
	TotalLinks      int        `json:"total_links"`
	MonthlyClicks   int        `json:"monthly_clicks"`
	RemainingClicks int        `json:"remaining_clicks"`
	PlanName        string     `json:"plan"`
	TopLinks        []TopLink  `json:"top_links"`
}

type TopLink struct {
	ShortCode   string `json:"short_code"`
	OriginalURL string `json:"original_url"`
	Clicks      int    `json:"clicks"`
	CreatedAt   string `json:"created_at"`
}

type DashboardService struct {
	DB    *gorm.DB
	Redis *redis.Client
}

func NewDashboardService(db *gorm.DB, redis *redis.Client) *DashboardService {
	return &DashboardService{DB: db, Redis: redis}
}

func (s *DashboardService) GetDashboard(orgID uint) (*DashboardStats, error) {

	// 1. Total active links
	var totalLinks int64
	s.DB.Model(&model.URL{}).
		Where("organization_id = ?", orgID).
		Count(&totalLinks)

	// 2. Monthly click count from Redis
	month := time.Now().Format("2006-01")
	redirectKey := fmt.Sprintf("usage:redirect:%d:%s", orgID, month)

	ctx := context.Background()
	monthlyClicks, _ := s.Redis.Get(ctx, redirectKey).Int()

	// 3. Plan limit for remaining calculation
	var sub model.Subscription
	planName := "free"
	maxRedirects := 25_000

	if err := s.DB.Where(
		"organization_id = ? AND status = ?",
		orgID,
		model.SubscriptionActive,
	).Preload("Plan").First(&sub).Error; err == nil {
		planName = sub.Plan.Name
		maxRedirects = sub.Plan.MaxRedirects
	}

	remaining := maxRedirects - monthlyClicks
	if remaining < 0 || maxRedirects == -1 {
		remaining = 0
		if maxRedirects == -1 {
			remaining = -1 // unlimited
		}
	}

	// 4. Top 5 links by click count
	var urls []model.URL
	s.DB.Where("organization_id = ?", orgID).
		Order("click_count DESC").
		Limit(5).
		Find(&urls)

	topLinks := make([]TopLink, 0, len(urls))
	for _, u := range urls {

		// Merge live Redis clicks
		clickKey := "clicks:" + u.ShortCode
		redisClicks, _ := s.Redis.Get(ctx, clickKey).Int()

		topLinks = append(topLinks, TopLink{
			ShortCode:   u.ShortCode,
			OriginalURL: u.OriginalURL,
			Clicks:      u.ClickCount + redisClicks,
			CreatedAt:   u.CreatedAt.Format(time.RFC3339),
		})
	}

	return &DashboardStats{
		TotalLinks:      int(totalLinks),
		MonthlyClicks:   monthlyClicks,
		RemainingClicks: remaining,
		PlanName:        planName,
		TopLinks:        topLinks,
	}, nil
}
