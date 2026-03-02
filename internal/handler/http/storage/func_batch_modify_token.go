package storage

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/xxcheng123/cloudpan189-share/internal/framework/httpcontext"
	filetasklogSvi "github.com/xxcheng123/cloudpan189-share/internal/services/filetasklog"
	"gorm.io/gorm"
)

type batchModifyTokenRequest struct {
	IDs     []int64 `json:"ids" binding:"required,min=1"`
	TokenID int64   `json:"tokenId" binding:"required"` // 新的令牌ID
}

// BatchModifyToken 批量修改存储挂载点令牌
// @Summary 批量修改存储挂载点令牌
// @Description 批量修改指定存储挂载点关联的云盘令牌
// @Tags 存储管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param request body batchModifyTokenRequest true "批量修改令牌请求参数"
// @Success 200 {object} httpcontext.Response "令牌修改成功"
// @Failure 400 {object} httpcontext.Response "参数验证失败，code=99998"
// @Failure 400 {object} httpcontext.Response "云盘令牌不存在，code=4013"
// @Failure 400 {object} httpcontext.Response "查询云盘令牌失败，code=4019"
// @Failure 401 {object} httpcontext.Response "未授权访问"
// @Failure 403 {object} httpcontext.Response "权限不足"
// @Router /api/storage/batch_modify_token [post]
func (h *handler) BatchModifyToken() httpcontext.HandlerFunc {
	return func(ctx *httpcontext.Context) {
		req := new(batchModifyTokenRequest)
		if err := ctx.ShouldBindJSON(req); err != nil {
			ctx.AbortWithInvalidParams(err)
			return
		}

		// 验证云盘令牌是否存在
		if req.TokenID != 0 {
			if _, err := h.cloudTokenService.Query(ctx.GetContext(), req.TokenID); err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					ctx.Fail(busCodeStorageCloudTokenNotExist.WithError(err))
				} else {
					ctx.Fail(busCodeStorageQueryCloudTokenError.WithError(err))
				}
				return
			}
		}

		// 批量修改挂载点的令牌
		successCount := 0
		failCount := 0
		for _, id := range req.IDs {
			// 验证挂载点是否存在
			if _, err := h.mountPointService.Query(ctx.GetContext(), id); err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					failCount++
					continue
				} else {
					failCount++
					continue
				}
			}

			// 修改挂载点的令牌
			if err := h.mountPointService.ModifyToken(ctx.GetContext(), id, req.TokenID); err != nil {
				failCount++
				continue
			}
			successCount++
		}

		// 创建任务日志
		tracker, _ := h.fileTaskLogService.Create(
			ctx.GetContext(),
			"批量修改令牌",
			fmt.Sprintf("批量修改 %d 个挂载点的令牌", len(req.IDs)),
			filetasklogSvi.WithFile(req.IDs[0]),
			filetasklogSvi.WithDesc(fmt.Sprintf("成功: %d, 失败: %d, 新令牌ID: %d", successCount, failCount, req.TokenID)),
		)
		if tracker != nil {
			_ = h.fileTaskLogService.Completed(ctx.GetContext(), tracker)
		}

		ctx.Success(fmt.Sprintf("修改成功 %d 个，失败 %d 个", successCount, failCount))
	}
}
