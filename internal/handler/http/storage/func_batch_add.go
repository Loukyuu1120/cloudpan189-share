package storage

import (
	stdContext "context"
	"encoding/json"
	"fmt"

	"github.com/xxcheng123/cloudpan189-share/internal/consts"
	"github.com/xxcheng123/cloudpan189-share/internal/framework/context"
	"github.com/xxcheng123/cloudpan189-share/internal/framework/httpcontext"
	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/datatypes"
	storagefacadeSvi "github.com/xxcheng123/cloudpan189-share/internal/services/storagefacade"
	"github.com/xxcheng123/cloudpan189-share/internal/types/topic"
	"go.uber.org/zap"
)

type batchAddRequest struct {
	Items []addRequest `json:"items" binding:"required,min=1,dive"`
}

type batchAddResponse struct {
	SuccessCount int               `json:"successCount"`
	FailCount    int               `json:"failCount"`
	Results      []addResponseItem `json:"results"`
}

type addResponseItem struct {
	LocalPath string `json:"localPath"`
	ID        int64  `json:"id,omitempty"`
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
}

// BatchAdd 批量添加存储挂载
// @Summary 批量添加存储挂载
// @Description 批量添加存储挂载点，通过后台任务异步执行文件扫描
// @Tags 存储管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param request body batchAddRequest true "批量存储挂载信息"
// @Success 200 {object} httpcontext.Response{data=batchAddResponse} "批量添加结果"
// @Failure 400 {object} httpcontext.Response "参数验证失败"
// @Failure 401 {object} httpcontext.Response "未授权访问"
// @Failure 403 {object} httpcontext.Response "权限不足"
// @Router /api/storage/batch_add [post]
func (h *handler) BatchAdd() httpcontext.HandlerFunc {
	return func(ctx *httpcontext.Context) {
		req := new(batchAddRequest)
		if err := ctx.ShouldBindJSON(req); err != nil {
			ctx.GetContext().Error("batch_add JSON解析失败", zap.Error(err))
			ctx.AbortWithInvalidParams(err)
			return
		}

		if req.Items == nil || len(req.Items) == 0 {
			ctx.Fail(busCodeStorageQueryPathFailed.WithError(fmt.Errorf("items不能为空")))
			return
		}

		ctx.GetContext().Info("batch_add请求收到，快速返回", zap.Int("items", len(req.Items)))

		results := make([]addResponseItem, 0, len(req.Items))
		successIds := make([]int64, 0, len(req.Items))

		for _, item := range req.Items {
			result := addResponseItem{
				LocalPath: item.LocalPath,
			}

			addition := datatypes.JSONMap{}
			fileId := item.FileId

			switch item.OsType {
			case protocolSubscribe:
				addition[consts.FileAdditionKeyUpUserId] = item.SubscribeUser
			case protocolSubscribeShare:
				addition[consts.FileAdditionKeyUpUserId] = item.SubscribeUser
				addition[consts.FileAdditionKeyShareId] = item.ShareCode
				addition[consts.FileAdditionKeyIsFolder] = true
			case protocolShare:
				addition[consts.FileAdditionKeyShareId] = item.ShareCode
				addition[consts.FileAdditionKeyAccessCode] = item.ShareAccessCode
				addition[consts.FileAdditionKeyIsFolder] = true
			case protocolPerson:
				addition = datatypes.JSONMap{}
			case protocolFamily:
				addition[consts.FileAdditionKeyFamilyId] = item.FamilyId
			}

			id, err := h.storageFacadeService.CreateStorage(ctx.GetContext(), &storagefacadeSvi.CreateStorageRequest{
				LocalPath:         item.LocalPath,
				OsType:            item.OsType,
				CloudToken:        item.CloudToken,
				FileId:            fileId,
				Addition:          addition,
				EnableAutoRefresh: item.EnableAutoRefresh,
				AutoRefreshDays:   item.AutoRefreshDays,
				RefreshInterval:   item.RefreshInterval,
				EnableDeepRefresh: item.EnableDeepRefresh,
			})
			if err != nil {
				result.Error = err.Error()
				results = append(results, result)
				continue
			}

			result.ID = id
			result.Success = true
			successIds = append(successIds, id)
			results = append(results, result)
		}

		successCount := 0
		failCount := 0
		for _, r := range results {
			if r.Success {
				successCount++
			} else {
				failCount++
			}
		}

		for _, id := range successIds {
			taskReq := &topic.FileScanFileRequest{
				FileId: id,
				Deep:   true,
			}
			body, _ := json.Marshal(taskReq)
			bgCtx := context.NewContext(stdContext.Background())
			_ = h.taskEngine.PushMessage(
				bgCtx.WithValue(consts.CtxKeyFullPath, "").
					WithValue(consts.CtxKeyInvokeHandlerName, "批量挂载扫描"),
				taskReq.Topic(), body)
		}
		ctx.GetContext().Info("批量挂载已推送扫描任务", zap.Int("count", len(successIds)))

		ctx.Success(&batchAddResponse{
			SuccessCount: successCount,
			FailCount:    failCount,
			Results:      results,
		})
	}
}
