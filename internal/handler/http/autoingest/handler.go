package autoingest

import (
	"github.com/xxcheng123/cloudpan189-share/internal/consts"
	"github.com/xxcheng123/cloudpan189-share/internal/framework/httpcontext"
	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/taskengine"
	autoingestlogSvi "github.com/xxcheng123/cloudpan189-share/internal/services/autoingestlog"
	autoingestplanSvi "github.com/xxcheng123/cloudpan189-share/internal/services/autoingestplan"
	cloudbridgeSvi "github.com/xxcheng123/cloudpan189-share/internal/services/cloudbridge"
)

type Handler interface {
	CreateSubscribePlan() httpcontext.HandlerFunc

	EnablePlan() httpcontext.HandlerFunc
	DisablePlan() httpcontext.HandlerFunc
	DeletePlan() httpcontext.HandlerFunc
	// PlanList 计划列表
	PlanList() httpcontext.HandlerFunc
	// LogList 日志查询
	LogList() httpcontext.HandlerFunc
	// Refresh 刷新计划（下发订阅刷新请求）
	Refresh() httpcontext.HandlerFunc
	// RetryFailed 重试失败的任务
	RetryFailed() httpcontext.HandlerFunc
	// UpdatePlan 修改计划
	UpdatePlan() httpcontext.HandlerFunc
	// DeleteErrorLogs 删除错误日志
	DeleteErrorLogs() httpcontext.HandlerFunc
}

var bi = httpcontext.NewBusinessGenerator(consts.BusCodeAutoIngestStartCode)

var (
	codePlanDeleteFailed  = bi.Next("删除自动挂载计划失败")
	codePlanEnableFailed  = bi.Next("启用自动挂载计划失败")
	codePlanDisableFailed = bi.Next("停用自动挂载计划失败")
	codePlanListFailed    = bi.Next("获取自动挂载计划列表失败")
	codeLogListFailed     = bi.Next("获取自动挂载日志列表失败")
	codeUpUserIdInvalid   = bi.Next("订阅号查询失败")
	codeCreatePlanFailed  = bi.Next("创建自动挂载计划失败")
	codePlanRefreshFailed = bi.Next("下发订阅刷新任务失败")
	codePlanNotFound      = bi.Next("自动挂载计划不存在")
	codePlanInvalidSource = bi.Next("自动挂载计划来源类型不支持刷新")
	codePlanQueryFailed   = bi.Next("查询自动挂载计划失败")
	codeLogDeleteFailed   = bi.Next("删除日志失败")
)

type handler struct {
	taskEngine         taskengine.TaskEngine
	planService        autoingestplanSvi.Service
	logService         autoingestlogSvi.Service
	cloudBridgeService cloudbridgeSvi.Service
}

func NewHandler(
	taskEngine taskengine.TaskEngine,
	planService autoingestplanSvi.Service,
	logService autoingestlogSvi.Service,
	cloudBridgeService cloudbridgeSvi.Service,
) Handler {
	return &handler{
		taskEngine:         taskEngine,
		planService:        planService,
		logService:         logService,
		cloudBridgeService: cloudBridgeService,
	}
}
