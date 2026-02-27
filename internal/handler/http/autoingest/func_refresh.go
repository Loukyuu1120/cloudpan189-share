package autoingest

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/xxcheng123/cloudpan189-share/internal/framework/httpcontext"
	"github.com/xxcheng123/cloudpan189-share/internal/repository/models"
	"github.com/xxcheng123/cloudpan189-share/internal/types/autoingest"
	"github.com/xxcheng123/cloudpan189-share/internal/types/topic"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type refreshPlanRequest struct {
	PlanId int64 `json:"planId" binding:"required" example:"1"`
}

type retryFailedRequest struct {
	PlanId int64 `json:"planId" binding:"required" example:"1"`
}

// Refresh 下发订阅计划刷新任务
// @Summary 下发订阅计划刷新任务
// @Description 传入计划ID，查询计划信息并下发 AutoIngestRefreshSubscribeRequest 任务
// @Tags 自动挂载管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param request body refreshPlanRequest true "刷新计划请求参数"
// @Success 200 {object} httpcontext.Response "任务已下发"
// @Failure 400 {object} httpcontext.Response "参数验证失败，code=99998"
// @Failure 400 {object} httpcontext.Response "自动挂载计划不存在，code=xxxx"
// @Failure 400 {object} httpcontext.Response "自动挂载计划来源类型不支持刷新，code=xxxx"
// @Failure 400 {object} httpcontext.Response "下发订阅刷新任务失败，code=xxxx"
// @Failure 401 {object} httpcontext.Response "未授权访问"
// @Failure 403 {object} httpcontext.Response "权限不足"
// @Router /api/auto_ingest/plan/refresh [post]
func (h *handler) Refresh() httpcontext.HandlerFunc {
	return func(ctx *httpcontext.Context) {
		req := new(refreshPlanRequest)
		if err := ctx.ShouldBindJSON(req); err != nil {
			ctx.AbortWithInvalidParams(err)

			return
		}

		plan, err := h.planService.Query(ctx.GetContext(), req.PlanId)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				ctx.Fail(codePlanNotFound.WithError(err))

				return
			}

			ctx.Fail(codePlanQueryFailed.WithError(err))

			return
		}

		// todo 目前仅支持订阅来源类型
		if plan.SourceType != autoingest.SourceTypeSubscribe {
			ctx.Fail(codePlanInvalidSource)

			return
		}

		// 解析订阅附加信息，获取 UpUserId
		var addition models.AutoIngestPlanSubscribeAddition

		_ = plan.Addition.Unmarshal(&addition)

		taskReq := &topic.AutoIngestRefreshSubscribeRequest{
			PlanId: plan.ID,
		}

		body, _ := json.Marshal(taskReq)
		if err = h.taskEngine.PushMessage(ctx.GetContext(), taskReq.Topic(), body); err != nil {
			ctx.Fail(codePlanRefreshFailed.WithError(err))

			return
		}

		ctx.Success()
	}
}

// RetryFailed 重试失败的入库任务
// @Summary 重试失败的入库任务
// @Description 重置偏移量并重新扫描所有文件，用于修复之前的失败
// @Tags 自动挂载管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param request body retryFailedRequest true "重试请求参数"
// @Success 200 {object} httpcontext.Response "任务已下发"
// @Failure 400 {object} httpcontext.Response "参数验证失败，code=99998"
// @Failure 400 {object} httpcontext.Response "自动挂载计划不存在，code=xxxx"
// @Failure 400 {object} httpcontext.Response "重试任务失败，code=xxxx"
// @Failure 401 {object} httpcontext.Response "未授权访问"
// @Failure 403 {object} httpcontext.Response "权限不足"
// @Router /api/auto_ingest/plan/retry_failed [post]
func (h *handler) RetryFailed() httpcontext.HandlerFunc {
	return func(ctx *httpcontext.Context) {
		req := new(retryFailedRequest)
		if err := ctx.ShouldBindJSON(req); err != nil {
			ctx.AbortWithInvalidParams(err)
			return
		}

		plan, err := h.planService.Query(ctx.GetContext(), req.PlanId)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				ctx.Fail(codePlanNotFound.WithError(err))
				return
			}
			ctx.Fail(codePlanQueryFailed.WithError(err))
			return
		}

		if plan.SourceType != autoingest.SourceTypeSubscribe {
			ctx.Fail(codePlanInvalidSource)
			return
		}

		// 将偏移量重置为0，重新扫描所有文件
		oldOffset := plan.Offset
		if err := h.planService.UpdateOffset(ctx.GetContext(), req.PlanId, 0); err != nil {
			ctx.GetContext().Error("重置偏移量失败", zap.Error(err))
			ctx.Fail(codePlanRefreshFailed.WithError(err))
			return
		}

		// 记录重置操作
		if _, logErr := h.logService.Create(ctx.GetContext(), req.PlanId, autoingest.LogLevelWarn,
			fmt.Sprintf("手动重试：将偏移量从 %d 重置为 0", oldOffset)); logErr != nil {
			ctx.GetContext().Error("创建重试日志失败", zap.Error(logErr))
		}

		// 下发刷新任务
		taskReq := &topic.AutoIngestRefreshSubscribeRequest{
			PlanId: plan.ID,
		}

		body, _ := json.Marshal(taskReq)
		if err = h.taskEngine.PushMessage(ctx.GetContext(), taskReq.Topic(), body); err != nil {
			ctx.Fail(codePlanRefreshFailed.WithError(err))
			return
		}

		ctx.Success("重试任务已下发，偏移量已重置")
	}
}
