package autoingestlog

import (
	"github.com/xxcheng123/cloudpan189-share/internal/bootstrap"
	appContext "github.com/xxcheng123/cloudpan189-share/internal/framework/context"
	"github.com/xxcheng123/cloudpan189-share/internal/repository/models"
	"github.com/xxcheng123/cloudpan189-share/internal/types/autoingest"
	"gorm.io/gorm"
)

// Service 面向 AutoIngestLog 的日志服务接口（与计划服务分离）
type Service interface {
	// Create 写日志
	Create(ctx appContext.Context, planID int64, level autoingest.LogLevel, content string) (int64, error)

	// List 查日志
	List(ctx appContext.Context, req *ListRequest) ([]*models.AutoIngestLog, error)
	Count(ctx appContext.Context, req *ListRequest) (int64, error)

	// DeleteByPlanIds 根据计划ID删除日志
	DeleteByPlanIds(ctx appContext.Context, planIds []int64) (int64, error)

	// DeleteByIds 根据ID删除日志
	DeleteByIds(ctx appContext.Context, ids []int64) (int64, error)

	// DeleteErrorLogsByPlanId 删除指定计划的所有错误日志
	DeleteErrorLogsByPlanId(ctx appContext.Context, planId int64) (int64, error)
}

// service 与构造函数，保持与 autoingestplan/mountpoint/filetasklog 风格一致
type service struct {
	svc bootstrap.ServiceContext
}

func NewService(svc bootstrap.ServiceContext) Service {
	return &service{
		svc: svc,
	}
}

func (s *service) getDB(ctx appContext.Context) *gorm.DB {
	return s.svc.GetDB(ctx).Model(new(models.AutoIngestLog))
}
