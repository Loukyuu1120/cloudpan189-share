package advance

import (
	"github.com/xxcheng123/cloudpan189-share/internal/framework/httpcontext"
	cloudbridgeSvi "github.com/xxcheng123/cloudpan189-share/internal/services/cloudbridge"
)

type getSubscribeUserRequest struct {
	SubscribeUser string `form:"subscribeUser" binding:"required" example:"user123"`
	Name          string `form:"name" example:"三生三世"`
	CurrentPage   int    `form:"currentPage,default=1" binding:"required,min=1" example:"1"`
	PageSize      int    `form:"pageSize,default=30" binding:"required,min=1,max=500" example:"30"`
}

type getSubscribeUserResponse struct {
	Name        string                              `json:"name" example:"订阅用户"`
	Total       int64                               `json:"total"`
	CurrentPage int                                 `json:"currentPage"`
	PageSize    int                                 `json:"pageSize"`
	Data        []*cloudbridgeSvi.ShareResourceInfo `json:"data"`
}

// GetSubscribeUser 获取订阅用户资源列表
// @Summary 获取订阅用户资源列表
// @Description 根据订阅用户名获取该用户的共享资源列表，支持分页和按文件名搜索
// @Tags 存储高级功能
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param subscribeUser query string true "订阅用户名" example("user123")
// @Param name query string false "文件名搜索" example("三生三世")
// @Param currentPage query int false "当前页码，默认为1" default(1)
// @Param pageSize query int false "每页大小，默认为10，最大100" default(10)
// @Success 200 {object} httpcontext.Response{data=getSubscribeUserResponse} "获取订阅用户资源列表成功"
// @Failure 400 {object} httpcontext.Response "参数验证失败，code=99998"
// @Failure 400 {object} httpcontext.Response "查询订阅信息失败，code=8001"
// @Failure 400 {object} httpcontext.Response "查询订阅用户列表失败，code=8002"
// @Failure 401 {object} httpcontext.Response "未授权访问"
// @Failure 403 {object} httpcontext.Response "权限不足"
// @Router /api/storage/advance/get_subscribe_user [get]
func (h *handler) GetSubscribeUser() httpcontext.HandlerFunc {
	return func(ctx *httpcontext.Context) {
		req := new(getSubscribeUserRequest)
		if err := ctx.ShouldBindQuery(req); err != nil {
			ctx.AbortWithInvalidParams(err)

			return
		}

		userInfo, err := h.cloudBridgeService.GetSubscribeUserInfo(ctx.GetContext(), req.SubscribeUser)
		if err != nil {
			ctx.Fail(codeStorageAdvanceQuerySubscribeUserError.WithError(err))

			return
		}

		list, count, err := h.cloudBridgeService.GetSubscribeUserShareResource(ctx.GetContext(), req.SubscribeUser, func(opt *cloudbridgeSvi.SubscribeUserShareResourceOption) {
			opt.FileName = req.Name
			opt.PageSize = req.PageSize
			opt.PageNum = req.CurrentPage
		})
		if err != nil {
			ctx.Fail(codeStorageAdvanceQuerySubscribeUserListError.WithError(err))

			return
		}

		ctx.Success(&getSubscribeUserResponse{
			Name:        userInfo.Name,
			Total:       count,
			Data:        list,
			CurrentPage: req.CurrentPage,
			PageSize:    req.PageSize,
		})
	}
}
