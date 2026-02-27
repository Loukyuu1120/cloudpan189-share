package autoingest

import (
	"github.com/xxcheng123/cloudpan189-share/internal/framework/httpcontext"
	autoingestlogSvi "github.com/xxcheng123/cloudpan189-share/internal/services/autoingestlog"
)

type deleteErrorLogsRequest struct {
	PlanId *int64 `json:"planId" example:"1"`
}

// DeleteErrorLogs 删除错误日志
// @Summary 删除错误日志
// @Description 删除指定计划的错误日志，如果不指定计划则删除所有错误日志
// @Tags 自动挂载管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param request body deleteErrorLogsRequest false "删除请求参数"
// @Success 200 {object} httpcontext.Response "删除成功"
// @Failure 400 {object} httpcontext.Response "参数验证失败"
// @Failure 401 {object} httpcontext.Response "未授权访问"
// @Router /api/auto_ingest/log/delete_error [post]
func (h *handler) DeleteErrorLogs() httpcontext.HandlerFunc {
	return func(ctx *httpcontext.Context) {
		req := new(deleteErrorLogsRequest)
		if err := ctx.ShouldBindJSON(req); err != nil {
			ctx.AbortWithInvalidParams(err)
			return
		}

		var count int64
		var err error

		if req.PlanId != nil && *req.PlanId > 0 {
			count, err = h.logService.DeleteErrorLogsByPlanId(ctx.GetContext(), *req.PlanId)
		} else {
			logs, err := h.logService.List(ctx.GetContext(), &autoingestlogSvi.ListRequest{
				Level:       "error",
				PageSize:    10000,
				CurrentPage: 1,
			})
			if err != nil {
				ctx.Fail(codeLogDeleteFailed.WithError(err))
				return
			}
			if len(logs) == 0 {
				ctx.Success(0)
				return
			}
			ids := make([]int64, len(logs))
			for i, log := range logs {
				ids[i] = log.ID
			}
			count, err = h.logService.DeleteByIds(ctx.GetContext(), ids)
		}

		if err != nil {
			ctx.Fail(codeLogDeleteFailed.WithError(err))
			return
		}

		ctx.Success(count)
	}
}
