package advance

import (
	"github.com/xxcheng123/cloudpan189-share/internal/consts"
	"github.com/xxcheng123/cloudpan189-share/internal/framework/httpcontext"
	cloudbridgeSvi "github.com/xxcheng123/cloudpan189-share/internal/services/cloudbridge"
	cloudtokenSvi "github.com/xxcheng123/cloudpan189-share/internal/services/cloudtoken"
)

type Handler interface {
	FamilyList() httpcontext.HandlerFunc
	GetFamilyFiles() httpcontext.HandlerFunc
	GetPersonFiles() httpcontext.HandlerFunc
	GetSubscribeUser() httpcontext.HandlerFunc
	GetSubscribeUserAll() httpcontext.HandlerFunc
	GetShareInfo() httpcontext.HandlerFunc
}

var bi = httpcontext.NewBusinessGenerator(consts.BusCodeStorageAdvanceStartCode)

var (
	codeStorageAdvanceCloudTokenNotExist          = bi.Next("云盘令牌不存在")
	codeStorageAdvanceQueryPathFailed             = bi.Next("查询路径失败")
	codeStorageAdvanceQuerySubscribeUserError     = bi.Next("查询订阅信息失败")
	codeStorageAdvanceQuerySubscribeUserListError = bi.Next("查询订阅用户列表失败")
	codeStorageAdvanceGetShareInfoError           = bi.Next("获取分享详情失败")
)

type handler struct {
	cloudBridgeService cloudbridgeSvi.Service
	cloudTokenService  cloudtokenSvi.Service
}

func NewHandler(cloudBridgeService cloudbridgeSvi.Service, cloudTokenService cloudtokenSvi.Service) Handler {
	return &handler{
		cloudBridgeService: cloudBridgeService,
		cloudTokenService:  cloudTokenService,
	}
}
