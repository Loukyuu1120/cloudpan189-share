package autoingestlog

import (
	appContext "github.com/xxcheng123/cloudpan189-share/internal/framework/context"
	"github.com/xxcheng123/cloudpan189-share/internal/repository/models"
)

func (s *service) DeleteByPlanIds(ctx appContext.Context, planIds []int64) (int64, error) {
	if len(planIds) == 0 {
		return 0, nil
	}
	db := s.getDB(ctx).Where("plan_id IN ?", planIds)
	result := db.Delete(&models.AutoIngestLog{})
	return result.RowsAffected, result.Error
}

func (s *service) DeleteByIds(ctx appContext.Context, ids []int64) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	db := s.getDB(ctx).Where("id IN ?", ids)
	result := db.Delete(&models.AutoIngestLog{})
	return result.RowsAffected, result.Error
}

func (s *service) DeleteErrorLogsByPlanId(ctx appContext.Context, planId int64) (int64, error) {
	db := s.getDB(ctx).Where("plan_id = ? AND level = ?", planId, "error")
	result := db.Delete(&models.AutoIngestLog{})
	return result.RowsAffected, result.Error
}
