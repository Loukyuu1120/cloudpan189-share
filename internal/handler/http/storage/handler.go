package storage

import (
	"github.com/xxcheng123/cloudpan189-share/internal/consts"
	"github.com/xxcheng123/cloudpan189-share/internal/framework/httpcontext"
	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/taskengine"
	"github.com/xxcheng123/cloudpan189-share/internal/repository/models"
	cloudbridgeSvi "github.com/xxcheng123/cloudpan189-share/internal/services/cloudbridge"
	cloudtokenSvi "github.com/xxcheng123/cloudpan189-share/internal/services/cloudtoken"
	filetasklogSvi "github.com/xxcheng123/cloudpan189-share/internal/services/filetasklog"
	mountPointSvi "github.com/xxcheng123/cloudpan189-share/internal/services/mountpoint"
	storageFacadeSvi "github.com/xxcheng123/cloudpan189-share/internal/services/storagefacade"
	virtualfileSvi "github.com/xxcheng123/cloudpan189-share/internal/services/virtualfile"
)

type Handler interface {
	Add() httpcontext.HandlerFunc
	BatchAdd() httpcontext.HandlerFunc
	Delete() httpcontext.HandlerFunc
	BatchDelete() httpcontext.HandlerFunc
	BatchParseFromText() httpcontext.HandlerFunc
	List() httpcontext.HandlerFunc
	SelectList() httpcontext.HandlerFunc
	Refresh() httpcontext.HandlerFunc
	ToggleAutoRefresh() httpcontext.HandlerFunc
	ModifyToken() httpcontext.HandlerFunc
}

var bi = httpcontext.NewBusinessGenerator(consts.BusCodeStorageStartCode)

var (
	busCodeStorageQueryPathFailed          = bi.Next("查询路径失败")
	busCodeStorageSubscribeUserEmpty       = bi.Next("订阅用户不能为空")
	busCodeStorageQuerySubscribeUserError  = bi.Next("查询订阅用户信息失败")
	busCodeStorageSubscribeShareIncomplete = bi.Next("订阅分享参数不完整")
	busCodeStorageQuerySubscribeShareError = bi.Next("查询订阅分享信息失败")
	busCodeStorageShareCodeEmpty           = bi.Next("分享码不能为空")
	busCodeStoragePersonParamsIncomplete   = bi.Next("个人云盘参数不完整")
	busCodeStorageFamilyParamsIncomplete   = bi.Next("家庭云盘参数不完整")
	busCodeStorageOsTypeNotMatch           = bi.Next("osType 不匹配")
	busCodeStorageOsTypeUnsupport          = bi.Next("osType 不支持")
	busCodeStorageCloudTokenNotExist       = bi.Next("cloudToken 不存在")
	busCodeStorageCloudTokenEmpty          = bi.Next("cloudToken 不能为空")
	busCodeStoragePersonFileQueryError     = bi.Next("查询个人文件失败")
	busCodeStorageFamilyFileQueryError     = bi.Next("查询家庭文件失败")
	busCodeStorageAddTaskFailed            = bi.Next("添加扫描任务失败")
	busCodeStorageAddMountPointFailed      = bi.Next("添加挂载点失败")
	busCodeStorageQueryMountPointError     = bi.Next("查询挂载点失败")
	busCodeStorageMountPointNotFound       = bi.Next("挂载节不存在")
	busCodeStorageMountPointDeleteFail     = bi.Next("挂载点删除失败")
	busCodeStorageSendTaskFail             = bi.Next("下发任务失败")
	busCodeStorageQueryCloudTokenError     = bi.Next("查询 cloudToken 失败")
	busCodeStorageQueryFileTaskLogError    = bi.Next("查询文件任务日志失败")
	busCodeStorageToggleAutoRefreshError   = bi.Next("切换自动刷新失败")
	busCodeStorageUpdateRefreshIntervalErr = bi.Next("更新刷新间隔失败")
	busCodeStorageTimeFormatErr            = bi.Next("时间格式错误")
	busCodeStorageQueryFileCountError      = bi.Next("查询文件数量失败")
	busCodeStorageModifyTokenFailed        = bi.Next("修改令牌失败")
)

const (
	protocolSubscribe      = models.OsTypeSubscribe
	protocolShare          = models.OsTypeShareFolder
	protocolPerson         = models.OsTypePersonFolder
	protocolFamily         = models.OsTypeFamilyFolder
	protocolSubscribeShare = models.OsTypeSubscribeShareFolder
)

type handler struct {
	taskEngine           taskengine.TaskEngine
	virtualFileService   virtualfileSvi.Service
	cloudBridgeService   cloudbridgeSvi.Service
	cloudTokenService    cloudtokenSvi.Service
	mountPointService    mountPointSvi.Service
	fileTaskLogService   filetasklogSvi.Service
	storageFacadeService storageFacadeSvi.Service
}

func NewHandler(
	taskEngine taskengine.TaskEngine,
	virtualFileService virtualfileSvi.Service,
	cloudBridgeService cloudbridgeSvi.Service,
	cloudTokenService cloudtokenSvi.Service,
	mountPointService mountPointSvi.Service,
	fileTaskLogService filetasklogSvi.Service,
	storageFacadeService storageFacadeSvi.Service,
) Handler {
	return &handler{
		virtualFileService:   virtualFileService,
		cloudBridgeService:   cloudBridgeService,
		cloudTokenService:    cloudTokenService,
		mountPointService:    mountPointService,
		taskEngine:           taskEngine,
		fileTaskLogService:   fileTaskLogService,
		storageFacadeService: storageFacadeService,
	}
}
