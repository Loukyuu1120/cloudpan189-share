package file

import (
	"encoding/json"

	"github.com/xxcheng123/cloudpan189-share/internal/consts"
	"github.com/xxcheng123/cloudpan189-share/internal/framework/httpcontext"
	"github.com/xxcheng123/cloudpan189-share/internal/types/topic"
	"go.uber.org/zap"
)

type batchDeleteRequest struct {
	IDs []int64 `json:"ids" binding:"required,min=1"`
}

// BatchDelete 批量删除文件
// @Router /api/file/batch_delete [post]
func (h *handler) BatchDelete() httpcontext.HandlerFunc {
	return func(ctx *httpcontext.Context) {
		req := new(batchDeleteRequest)
		if err := ctx.ShouldBindJSON(req); err != nil {
			ctx.AbortWithInvalidParams(err)
			return
		}

		// 先删除数据库中的挂载点记录
		if err := h.mountPointService.BatchDelete(ctx.GetContext(), req.IDs); err != nil {
			ctx.GetContext().Error("批量删除挂载点记录失败", zap.Error(err), zap.Int64s("ids", req.IDs))
			ctx.Fail(busCodeBatchDeleteError.WithError(err))
			return
		}

		ctx.GetContext().Info("批量删除挂载点记录成功", zap.Int("count", len(req.IDs)))

		// 构造消息队列请求
		task := &topic.FileBatchDeleteRequest{IDs: req.IDs}
		body, _ := json.Marshal(task)

		fullPath := ctx.GetContext().String(consts.CtxKeyFullPath, "unknown")

		// 推送消息到队列
		err := h.taskEngine.PushMessage(
			ctx.GetContext().
				WithValue(consts.CtxKeyFullPath, fullPath).
				WithValue(consts.CtxKeyInvokeHandlerName, "API批量删除文件"),
			task.Topic(),
			body,
		)

		if err != nil {
			ctx.GetContext().Error("推送文件批量删除任务失败", zap.Error(err))
			// [修正] 使用 busCodeBatchDeleteError 替代 NewError
			ctx.Fail(busCodeBatchDeleteError.WithError(err))
			return
		}

		ctx.GetContext().Info("批量删除文件请求已加入队列", zap.Int("count", len(req.IDs)))
		ctx.Success("删除任务已提交，后台处理中")
	}
}
