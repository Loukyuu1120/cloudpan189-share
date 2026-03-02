package taskstate

import (
	"time"

	"github.com/xxcheng123/cloudpan189-share/internal/framework/httpcontext"
	"github.com/xxcheng123/cloudpan189-share/internal/repository/models"
	filetasklogSvi "github.com/xxcheng123/cloudpan189-share/internal/services/filetasklog"
)

type (
	fileLogListRequest = filetasklogSvi.ListRequest

	fileLogListResponse struct {
		Total       int64                 `json:"total" example:"100"`     // 总记录数
		CurrentPage int                   `json:"currentPage" example:"1"` // 当前页码
		PageSize    int                   `json:"pageSize" example:"10"`   // 每页大小
		Data        []*models.FileTaskLog `json:"data"`                    // 任务日志列表数据
	}
)

// FileLogList 获取文件任务日志列表
// @Summary 获取文件任务日志列表
// @Description 分页获取文件任务日志列表，支持多种筛选条件
// @Tags 任务状态管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param currentPage query int false "当前页码，默认为1" default(1)
// @Param pageSize query int false "每页大小，默认为10" default(10)
// @Param noPaginate query bool false "是否不分页，默认false" default(false)
// @Param type query string false "任务类型筛选"
// @Param status query string false "任务状态筛选(pending/running/completed/failed)"
// @Param fileId query int false "文件ID筛选"
// @Param userId query int false "用户ID筛选"
// @Param title query string false "任务标题模糊搜索"
// @Param beginAt query string false "开始时间筛选(格式: 2006-01-02T15:04:05Z07:00)"
// @Param endAt query string false "结束时间筛选(格式: 2006-01-02T15:04:05Z07:00)"
// @Success 200 {object} httpcontext.Response{data=fileLogListResponse} "获取任务日志列表成功"
// @Failure 400 {object} httpcontext.Response "参数验证失败，code=99998"
// @Failure 400 {object} httpcontext.Response "任务列表获取失败，code=7001"
// @Failure 400 {object} httpcontext.Response "任务数量统计失败，code=7002"
// @Failure 401 {object} httpcontext.Response "未授权访问"
// @Failure 403 {object} httpcontext.Response "权限不足"
// @Router /api/task_state/file_log/list [get]
func (h *handler) FileLogList() httpcontext.HandlerFunc {
	return func(ctx *httpcontext.Context) {
		req := new(fileLogListRequest)
		if err := ctx.ShouldBindQuery(req); err != nil {
			ctx.AbortWithInvalidParams(err)

			return
		}

		// 获取任务日志列表
		taskLogList, err := h.fileTaskLogService.List(ctx.GetContext(), req)
		if err != nil {
			ctx.Fail(codeListTasksFailed.WithError(err))

			return
		}

		var total int64
		if !req.NoPaginate {
			// 获取总数
			total, err = h.fileTaskLogService.Count(ctx.GetContext(), req)
			if err != nil {
				ctx.Fail(codeCountTasksFailed.WithError(err))

				return
			}
		} else {
			total = int64(len(taskLogList))
		}

		now := time.Now()

		for _, log := range taskLogList {
			if log.Duration == 0 {
				endTime := log.EndAt
				if endTime == nil {
					endTime = &now
				}
				log.Duration = endTime.UnixMilli() - log.BeginAt.UnixMilli()
		}

		ctx.Success(&fileLogListResponse{
			Total:       total,
			Data:        taskLogList,
			PageSize:    req.PageSize,
			CurrentPage: req.CurrentPage,
		})
	}
}
