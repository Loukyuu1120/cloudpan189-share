package advance

import (
	"github.com/xxcheng123/cloudpan189-share/internal/framework/httpcontext"
	cloudbridgeSvi "github.com/xxcheng123/cloudpan189-share/internal/services/cloudbridge"
)

type getSubscribeUserAllRequest struct {
	SubscribeUser string `form:"subscribeUser" binding:"required" example:"user123"`
}

type getSubscribeUserAllResponse struct {
	Name        string                              `json:"name" example:"订阅用户"`
	Total       int64                               `json:"total"`
	Data        []*cloudbridgeSvi.ShareResourceInfo `json:"data"`
}

// GetSubscribeUserAll 获取订阅用户所有资源列表（不分页，获取全部）
// @Summary 获取订阅用户所有资源列表
// @Description 根据订阅用户名获取该用户的共享资源列表（获取全部，不分页）
// @Tags 存储高级功能
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param subscribeUser query string true "订阅用户名" example("user123")
// @Success 200 {object} httpcontext.Response{data=getSubscribeUserAllResponse} "获取订阅用户资源列表成功"
// @Failure 400 {object} httpcontext.Response "参数验证失败"
// @Failure 400 {object} httpcontext.Response "查询订阅信息失败"
// @Failure 401 {object} httpcontext.Response "未授权访问"
// @Router /api/storage/advance/get_subscribe_user_all [get]
func (h *handler) GetSubscribeUserAll() httpcontext.HandlerFunc {
	return func(ctx *httpcontext.Context) {
		req := new(getSubscribeUserAllRequest)
		if err := ctx.ShouldBindQuery(req); err != nil {
			ctx.AbortWithInvalidParams(err)
			return
		}

		userInfo, err := h.cloudBridgeService.GetSubscribeUserInfo(ctx.GetContext(), req.SubscribeUser)
		if err != nil {
			ctx.Fail(codeStorageAdvanceQuerySubscribeUserError.WithError(err))
			return
		}

		list, count, err := h.cloudBridgeService.GetSubscribeUserShareResourceAll(ctx.GetContext(), req.SubscribeUser)
		if err != nil {
			ctx.Fail(codeStorageAdvanceQuerySubscribeUserListError.WithError(err))
			return
		}

		ctx.Success(&getSubscribeUserAllResponse{
			Name:  userInfo.Name,
			Total: count,
			Data:  list,
		})
	}
}
