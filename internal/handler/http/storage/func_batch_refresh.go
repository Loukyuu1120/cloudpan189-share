package storage

import (
	"encoding/json"
	"fmt"

	"github.com/xxcheng123/cloudpan189-share/internal/consts"
	"github.com/xxcheng123/cloudpan189-share/internal/framework/httpcontext"
	filetasklogSvi "github.com/xxcheng123/cloudpan189-share/internal/services/filetasklog"
	"github.com/xxcheng123/cloudpan189-share/internal/types/topic"
	"go.uber.org/zap"
)

type batchRefreshRequest struct {
	IDs  []int64 `json:"ids" binding:"required,min=1"`
	Deep bool    `json:"deep"` // 是否深度刷新
}

// BatchRefresh 批量刷新存储挂载
// @Summary 批量刷新存储挂载
// @Description 批量刷新指定的存储挂载点，触发文件扫描任务重新同步文件信息
// @Tags 存储管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param request body batchRefreshRequest true "批量刷新请求参数"
// @Success 200 {object} httpcontext.Response "刷新任务已提交，后台处理中"
// @Failure 400 {object} httpcontext.Response "参数验证失败，code=99998"
// @Failure 400 {object} httpcontext.Response "发送刷新任务失败，code=4024"
// @Failure 401 {object} httpcontext.Response "未授权访问"
// @Failure 403 {object} httpcontext.Response "权限不足"
// @Router /api/storage/batch_refresh [post]
func (h *handler) BatchRefresh() httpcontext.HandlerFunc {
	return func(ctx *httpcontext.Context) {
		req := new(batchRefreshRequest)
		if err := ctx.ShouldBindJSON(req); err != nil {
			ctx.AbortWithInvalidParams(err)
			return
		}

		refreshType := "普通刷新"
		if req.Deep {
			refreshType = "深度刷新"
		}

		// 创建任务日志
		tracker, _ := h.fileTaskLogService.Create(
			ctx.GetContext(),
			fmt.Sprintf("批量%s", refreshType),
			fmt.Sprintf("批量%s %d 个挂载点", refreshType, len(req.IDs)),
			filetasklogSvi.WithFile(req.IDs[0]),
			filetasklogSvi.WithDesc(fmt.Sprintf("ID列表: %v, 类型: %s", req.IDs, refreshType)),
		)

		// 为每个挂载点创建扫描任务
		successCount := 0
		for _, id := range req.IDs {
			mountPoint, err := h.mountPointService.Query(ctx.GetContext(), id)
			if err != nil {
				ctx.GetContext().Warn("查询挂载点失败，跳过", zap.Int64("id", id), zap.Error(err))
				continue
			}

			taskReq := &topic.FileScanFileRequest{
				FileId: id,
				Deep:   req.Deep,
			}
			body, _ := json.Marshal(taskReq)

			err = h.taskEngine.PushMessage(
				ctx.GetContext().
					WithValue(consts.CtxKeyFullPath, mountPoint.FullPath).
					WithValue(consts.CtxKeyInvokeHandlerName, fmt.Sprintf("批量%s", refreshType)),
				taskReq.Topic(),
				body,
			)

			if err != nil {
				ctx.GetContext().Warn("推送刷新任务失败，跳过", zap.Int64("id", id), zap.Error(err))
				continue
			}
			successCount++
		}

		if tracker != nil {
			_ = h.fileTaskLogService.Completed(ctx.GetContext(), tracker)
		}

		ctx.GetContext().Info("批量刷新请求已加入队列", zap.Int("total", len(req.IDs)), zap.Int("success", successCount))
		ctx.Success(fmt.Sprintf("刷新任务已提交，成功 %d 个，失败 %d 个", successCount, len(req.IDs)-successCount))
	}
}
