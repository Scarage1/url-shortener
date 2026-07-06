package model

import "gorm.io/gorm"

const (
	RoutingRuleTypePassword = "password"
	RoutingRuleTypeSchedule = "schedule"
	RoutingRuleTypeGeo      = "geo"
)

type RoutingRule struct {
	gorm.Model
	URLID  uint   `gorm:"not null;index"`
	Type   string `gorm:"not null;index"`
	Config []byte `gorm:"type:jsonb;not null"`
}
