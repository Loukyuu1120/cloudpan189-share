package models

import (
	"time"

	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/datatypes"
	"github.com/xxcheng123/cloudpan189-share/internal/types/autoingest"
)

type RefreshStrategy struct {
	EnableAutoRefresh bool `gorm:"column:enable_auto_refresh;not null;default:false" json:"enableAutoRefresh"`
	AutoRefreshDays   int  `gorm:"column:auto_refresh_days;not null;default:7;comment:'单位:刷新持续天数'" json:"autoRefreshDays"`
	RefreshInterval   int  `gorm:"column:refresh_interval;not null;default:30;comment:'单位:分钟，最小值30'" json:"refreshInterval"`
	EnableDeepRefresh bool `gorm:"column:enable_deep_refresh;not null;default:false" json:"enableDeepRefresh"`
}

type AutoIngestPlan struct {
	ID                 int64                 `gorm:"primaryKey" json:"id"`
	Name               string                `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Enabled            bool                  `gorm:"column:enabled;type:tinyint(1);default:1" json:"enabled"`
	AutoIngestInterval int64                 `gorm:"column:auto_ingest_interval;type:bigint(20);default:30" json:"autoIngestInterval"` // 单位分钟
	SourceType         autoingest.SourceType `gorm:"column:source_type;not null;type:varchar(64)" json:"sourceType"`
	Offset             int64                 `gorm:"column:offset;type:bigint(20);default:0;" json:"offset"`                            // 偏移量
	ParentPath         string                `gorm:"column:parent_path;type:varchar(255);not null;default:'';" json:"parentPath"`       // 父目录路径
	OnConflict         autoingest.OnConflict `gorm:"column:on_conflict;type:varchar(255);not null;default:'rename';" json:"onConflict"` // 冲突处理策略
	ConcurrentCount    int                   `gorm:"column:concurrent_count;default:4" json:"concurrentCount"`                         // 并发数
	MaxRetryCount      int                   `gorm:"column:max_retry_count;default:3" json:"maxRetryCount"`                           // 最大重试次数
	AddCount           int64                 `gorm:"column:add_count;type:bigint(20);default:0;" json:"addCount"`                       // 新增挂载数
	FailedCount        int64                 `gorm:"column:failed_count;type:bigint(20);default:0;" json:"failedCount"`                 // 失败挂载数
	Addition           datatypes.JSONMap     `gorm:"column:addition;type:json;default:'';" json:"addition"`
	RefreshStrategy    RefreshStrategy       `gorm:"embedded;embeddedPrefix:refresh_strategy_" json:"refreshStrategy"`
	TokenId            int64                 `gorm:"column:token_id;type:bigint(20);not null;default:0;" json:"tokenId"`
	CreatedAt          time.Time             `gorm:"column:created_at;autoCreateTime;type:datetime;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt          time.Time             `gorm:"column:updated_at;autoUpdateTime;type:datetime;default:CURRENT_TIMESTAMP;on update:CURRENT_TIMESTAMP" json:"updatedAt"`
}

func (a *AutoIngestPlan) TableName() string {
	return "auto_ingest_plans"
}

type AutoIngestLog struct {
	ID        int64               `gorm:"primaryKey" json:"id"`
	PlanId    int64               `gorm:"column:plan_id;type:bigint(20);not null;default:0;" json:"planId"`
	Level     autoingest.LogLevel `gorm:"column:level;type:varchar(64);not null;default:info" json:"level"`
	Content   string              `gorm:"column:content;type:text;not null;default:'';" json:"content"`
	CreatedAt time.Time           `gorm:"column:created_at;autoCreateTime;type:datetime;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt time.Time           `gorm:"column:updated_at;autoUpdateTime;type:datetime;default:CURRENT_TIMESTAMP;on update:CURRENT_TIMESTAMP" json:"updatedAt"`
}

func (a *AutoIngestLog) TableName() string {
	return "auto_ingest_logs"
}

type AutoIngestPlanSubscribeAddition struct {
	UpUserId string `json:"upUserId"`
}

func (a *AutoIngestPlanSubscribeAddition) JSONMap() datatypes.JSONMap {
	m, _ := datatypes.FromStruct(a)

	return m
}
