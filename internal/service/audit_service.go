package service

import (
	"encoding/json"
	"log"
	"time"

	"github.com/Scarage1/url-shortener/internal/model"
	"gorm.io/gorm"
)

type AuditService struct {
	DB *gorm.DB
}

func NewAuditService(db *gorm.DB) *AuditService {
	return &AuditService{DB: db}
}

// Log records an audit event asynchronously in the database to prevent blocking API requests.
func (s *AuditService) Log(orgID, actorID uint, action, resourceType, resourceID, ip string, meta map[string]interface{}) {
	var metaStr string
	if meta != nil {
		if b, err := json.Marshal(meta); err == nil {
			metaStr = string(b)
		}
	}

	auditLog := model.AuditLog{
		OrganizationID: orgID,
		ActorID:        actorID,
		Action:         action,
		ResourceType:   resourceType,
		ResourceID:     resourceID,
		Metadata:       metaStr,
		IPAddress:      ip,
		CreatedAt:      time.Now(),
	}

	go func() {
		// Use a new DB session connection for the goroutine
		if err := s.DB.Session(&gorm.Session{}).Create(&auditLog).Error; err != nil {
			log.Printf("[AUDIT ERROR] failed to write audit log: %v", err)
		}
	}()
}

// GetLogs returns audit logs for an organization.
func (s *AuditService) GetLogs(orgID uint, limit, offset int) ([]model.AuditLog, error) {
	var logs []model.AuditLog
	err := s.DB.Where("organization_id = ?", orgID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error
	return logs, err
}
