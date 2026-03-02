package http

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	embed "github.com/xxcheng123/cloudpan189-share"
	"github.com/xxcheng123/cloudpan189-share/internal/handler/http/taskstate"
	"github.com/xxcheng123/cloudpan189-share/internal/types/loginlog"

	"github.com/xxcheng123/cloudpan189-share/internal/handler/http/autoingest"
	"github.com/xxcheng123/cloudpan189-share/internal/handler/http/cloudtoken"
	"github.com/xxcheng123/cloudpan189-share/internal/handler/http/file"
	loginlogHandler "github.com/xxcheng123/cloudpan189-share/internal/handler/http/loginlog"
	"github.com/xxcheng123/cloudpan189-share/internal/handler/http/media"
	"github.com/xxcheng123/cloudpan189-share/internal/handler/http/setting"
	"github.com/xxcheng123/cloudpan189-share/internal/handler/http/storage"
	"github.com/xxcheng123/cloudpan189-share/internal/handler/http/storage/advance"
	"github.com/xxcheng123/cloudpan189-share/internal/handler/http/usergroup"

	"github.com/xxcheng123/cloudpan189-share/internal/bootstrap"
	"github.com/xxcheng123/cloudpan189-share/internal/framework/httpcontext"
	"github.com/xxcheng123/cloudpan189-share/internal/handler/http/user"

	autoingestlogSvi "github.com/xxcheng123/cloudpan189-share/internal/services/autoingestlog"
	autoingestplanSvi "github.com/xxcheng123/cloudpan189-share/internal/services/autoingestplan"
	cloudbridgeSvi "github.com/xxcheng123/cloudpan189-share/internal/services/cloudbridge"
	cloudtokenSvi "github.com/xxcheng123/cloudpan189-share/internal/services/cloudtoken"
	filetasklogSvi "github.com/xxcheng123/cloudpan189-share/internal/services/filetasklog"
	group2fileSvi "github.com/xxcheng123/cloudpan189-share/internal/services/group2file"
	loginlogSvi "github.com/xxcheng123/cloudpan189-share/internal/services/loginlog"
	mediaconfigSvi "github.com/xxcheng123/cloudpan189-share/internal/services/mediaconfig"
	mediafileSvi "github.com/xxcheng123/cloudpan189-share/internal/services/mediafile"
	mountPointSvi "github.com/xxcheng123/cloudpan189-share/internal/services/mountpoint"
	settingSvi "github.com/xxcheng123/cloudpan189-share/internal/services/setting"
	storagefacadeSvi "github.com/xxcheng123/cloudpan189-share/internal/services/storagefacade"
	userSvi "github.com/xxcheng123/cloudpan189-share/internal/services/user"
	userGroupSvi "github.com/xxcheng123/cloudpan189-share/internal/services/usergroup"
	verifySvi "github.com/xxcheng123/cloudpan189-share/internal/services/verify"
	virtualfileSvi "github.com/xxcheng123/cloudpan189-share/internal/services/virtualfile"
)

func Start(svc bootstrap.ServiceContext) {
	const (
		handlerName = "http"
	)

	var (
		engine = svc.GetHTTPEngine()

		logger     = svc.GetLogger(handlerName)
		taskEngine = svc.GetTaskEngine()

		wrapper = httpcontext.NewHandlerFuncWrapper(logger)
		wrap    = wrapper.Wrap
	)

	var (
		userService           = userSvi.NewService(svc)
		userGroupService      = userGroupSvi.NewService(svc)
		group2FileService     = group2fileSvi.NewService(svc)
		settingService        = settingSvi.NewService(svc)
		virtualFileService    = virtualfileSvi.NewService(svc)
		cloudBridgeService    = cloudbridgeSvi.NewService(svc)
		cloudTokenService     = cloudtokenSvi.NewService(svc)
		mountPointService     = mountPointSvi.NewService(svc, cloudTokenService, cloudBridgeService)
		fileTaskLogService    = filetasklogSvi.NewService(svc)
		storageFacadeService  = storagefacadeSvi.NewService(svc)
		verifyService         = verifySvi.NewService(svc)
		autoIngestPlanService = autoingestplanSvi.NewService(svc)
		autoIngestLogService  = autoingestlogSvi.NewService(svc)
		loginLogService       = loginlogSvi.NewService(svc)
		mediaConfigService    = mediaconfigSvi.NewService(svc)
		mediaFileService      = mediafileSvi.NewService(svc)
	)

	var (
		userHandler           = user.NewHandler(userService, userGroupService, loginLogService)
		settingHandler        = setting.NewHandler(userService, settingService)
		userGroupHandler      = usergroup.NewHandler(userGroupService, group2FileService, userService)
		storageHandler        = storage.NewHandler(taskEngine, virtualFileService, cloudBridgeService, cloudTokenService, mountPointService, fileTaskLogService, storageFacadeService)
		storageAdvanceHandler = advance.NewHandler(cloudBridgeService, cloudTokenService)
		cloudTokenHandler     = cloudtoken.NewHandler(cloudTokenService, mountPointService)
		fileHandler           = file.NewHandler(virtualFileService, verifyService, cloudTokenService, cloudBridgeService, mountPointService, group2FileService, taskEngine)

		taskStateHandler  = taskstate.NewHandler(taskEngine, fileTaskLogService)
		autoIngestHandler = autoingest.NewHandler(taskEngine, autoIngestPlanService, autoIngestLogService, cloudBridgeService)
		loginLogHandler   = loginlogHandler.NewHandler(loginLogService)
		mediaHandler      = media.NewHandler(mediaConfigService, mediaFileService, mountPointService, virtualFileService, verifyService, taskEngine)
	)

	var (
		userMiddleware = newAuthMiddleware(userService)
	)

	openapiRouter := engine.Group("/api", httpcontext.LoggerHandler(logger))

	{
		userRouter := openapiRouter.Group("/user")
		{
			userRouter.POST("/login", wrap(userHandler.RecordLog(loginlog.EventLogin)), wrap(userHandler.Login()))
			userRouter.POST("/refresh_token", wrap(userHandler.RecordLog(loginlog.EventLogin)), wrap(userHandler.RefreshToken()))
		}

		userRouterWithAdminAuth := openapiRouter.Group("/user", wrap(userMiddleware.Auth(true)))
		{
			userRouterWithAdminAuth.POST("/add", wrap(userHandler.Add()))
			userRouterWithAdminAuth.POST("/del", wrap(userHandler.Del()))
			userRouterWithAdminAuth.POST("/update", wrap(userHandler.Update()))
			userRouterWithAdminAuth.POST("/toggle_status", wrap(userHandler.ToggleStatus()))
			userRouterWithAdminAuth.GET("/list", wrap(userHandler.List()))
			userRouterWithAdminAuth.POST("/modify_pass", wrap(userHandler.ModifyPass()))
			userRouterWithAdminAuth.POST("/bind_group", wrap(userHandler.BindGroup()))
		}

		userRouterWithBaseAuth := openapiRouter.Group("/user", wrap(userMiddleware.Auth()))
		{
			userRouterWithBaseAuth.GET("/info", wrap(userHandler.Info()))
			userRouterWithBaseAuth.POST("/modify_own_pass", wrap(userHandler.ModifyOwnPass()))
		}
	}

	{
		userGroupRouter := openapiRouter.Group("/user_group", wrap(userMiddleware.Auth(true)))
		{
			userGroupRouter.POST("/add", wrap(userGroupHandler.Add()))
			userGroupRouter.POST("/delete", wrap(userGroupHandler.Delete()))
			userGroupRouter.POST("/modify_name", wrap(userGroupHandler.ModifyName()))
			userGroupRouter.GET("/list", wrap(userGroupHandler.List()))
			userGroupRouter.POST("/batch_bind_files", wrap(userGroupHandler.BatchBindFiles()))
			userGroupRouter.GET("/bind_files", wrap(userGroupHandler.GetBindFiles()))
		}
	}

	{
		storageRouter := openapiRouter.Group("/storage", wrap(userMiddleware.Auth()))
		{
			storageRouter.POST("/add", wrap(storageHandler.Add()))
			storageRouter.POST("/batch_add", wrap(storageHandler.BatchAdd()))
			storageRouter.POST("/delete", wrap(storageHandler.Delete()))
			storageRouter.POST("/batch_delete", wrap(storageHandler.BatchDelete()))
			storageRouter.POST("/batch_parse_text", wrap(storageHandler.BatchParseFromText()))
			storageRouter.POST("/batch_refresh", wrap(storageHandler.BatchRefresh()))
			storageRouter.POST("/batch_modify_token", wrap(storageHandler.BatchModifyToken()))
			storageRouter.GET("/list", wrap(storageHandler.List()))
			storageRouter.GET("/select_list", wrap(storageHandler.SelectList()))
			storageRouter.POST("/refresh", wrap(storageHandler.Refresh()))
			storageRouter.POST("/toggle_auto_refresh", wrap(storageHandler.ToggleAutoRefresh()))
			storageRouter.POST("/modify_token", wrap(storageHandler.ModifyToken()))
		}

		storageAdvanceRouter := openapiRouter.Group("/storage/advance", wrap(userMiddleware.Auth()))
		{
			storageAdvanceRouter.GET("/person/files", wrap(storageAdvanceHandler.GetPersonFiles()))
			storageAdvanceRouter.GET("/family/files", wrap(storageAdvanceHandler.GetFamilyFiles()))
			storageAdvanceRouter.GET("/family/list", wrap(storageAdvanceHandler.FamilyList()))
			storageAdvanceRouter.GET("/get_subscribe_user", wrap(storageAdvanceHandler.GetSubscribeUser()))
			storageAdvanceRouter.GET("/get_subscribe_user_all", wrap(storageAdvanceHandler.GetSubscribeUserAll()))
			storageAdvanceRouter.GET("/share_info", wrap(storageAdvanceHandler.GetShareInfo()))
		}
	}

	{
		fileRouter := openapiRouter.Group("/file", wrap(userMiddleware.Auth()))
		{
			fileRouter.GET("/search", wrap(fileHandler.Search()))
			fileRouter.POST("/create_download_url", wrap(fileHandler.CreateDownloadURL()))
			fileRouter.GET("/open/*fullPath", wrap(fileHandler.Open()))
			fileRouter.POST("/batch_delete", wrap(fileHandler.BatchDelete()))
		}

		{
			openapiRouter.GET("/file/download/:fileId", wrap(fileHandler.Download()))
		}
	}

	{
		cloudTokenRouter := openapiRouter.Group("/cloud_token", wrap(userMiddleware.Auth(true)))
		{
			cloudTokenRouter.POST("/init_qrcode", wrap(cloudTokenHandler.InitQrcode()))
			cloudTokenRouter.POST("/check_qrcode", wrap(cloudTokenHandler.CheckQrcode()))
			cloudTokenRouter.POST("/username_login", wrap(cloudTokenHandler.UsernameLogin()))
			cloudTokenRouter.POST("/modify_name", wrap(cloudTokenHandler.ModifyName()))
			cloudTokenRouter.POST("/delete", wrap(cloudTokenHandler.Delete()))
			cloudTokenRouter.GET("/list", wrap(cloudTokenHandler.List()))
			cloudTokenRouter.GET("/:id", wrap(cloudTokenHandler.Query()))
		}
	}

	{
		taskStateRouter := openapiRouter.Group("/task_state", wrap(userMiddleware.Auth(true)))
		{
			taskStateRouter.GET("/file_log/list", wrap(taskStateHandler.FileLogList()))
			taskStateRouter.GET("/task_engine/list", wrap(taskStateHandler.TaskEngineList()))
		}
	}

	{
		openapiRouter.POST("/setting/init_system", wrap(settingHandler.InitSystem()))
		openapiRouter.GET("/setting/info", wrap(settingHandler.Info()))
	}
	{
		settingBaseRouter := openapiRouter.Group("/setting", wrap(userMiddleware.Auth()))
		{
			settingBaseRouter.GET("/addition", wrap(settingHandler.Addition()))
		}
	}

	{
		settingAdminRouter := openapiRouter.Group("/setting", wrap(userMiddleware.Auth(true)))
		{
			settingAdminRouter.POST("/modify_title", wrap(settingHandler.ModifyTitle()))
			settingAdminRouter.POST("/modify_base_url", wrap(settingHandler.ModifyBaseURL()))
			settingAdminRouter.POST("/toggle_enable_auth", wrap(settingHandler.ToggleEnableAuth()))
			settingAdminRouter.POST("/modify_addition", wrap(settingHandler.ModifyAddition()))
		}
	}

	{
		autoIngestRouter := openapiRouter.Group("/auto_ingest", wrap(userMiddleware.Auth(true)))
		{
			autoIngestRouter.POST("/plan/create_subscribe", wrap(autoIngestHandler.CreateSubscribePlan()))
			autoIngestRouter.GET("/plan/list", wrap(autoIngestHandler.PlanList()))
			autoIngestRouter.POST("/plan/enable", wrap(autoIngestHandler.EnablePlan()))
			autoIngestRouter.POST("/plan/disable", wrap(autoIngestHandler.DisablePlan()))
			autoIngestRouter.POST("/plan/refresh", wrap(autoIngestHandler.Refresh()))
			autoIngestRouter.POST("/plan/retry_failed", wrap(autoIngestHandler.RetryFailed()))
			autoIngestRouter.POST("/plan/delete", wrap(autoIngestHandler.DeletePlan()))
			autoIngestRouter.POST("/plan/update", wrap(autoIngestHandler.UpdatePlan()))
			autoIngestRouter.GET("/log/list", wrap(autoIngestHandler.LogList()))
			autoIngestRouter.POST("/log/delete_error", wrap(autoIngestHandler.DeleteErrorLogs()))
		}
	}

	{
		loginLogRouter := openapiRouter.Group("/login_log", wrap(userMiddleware.Auth(true)))
		{
			loginLogRouter.GET("/list", wrap(loginLogHandler.List()))
		}
	}

	{
		mediaRouter := openapiRouter.Group("/media", wrap(userMiddleware.Auth(true)))
		{
			mediaRouter.GET("/config/info", wrap(mediaHandler.ConfigInfo()))
			mediaRouter.POST("/config/init", wrap(mediaHandler.ConfigInit()))
			mediaRouter.POST("/config/update", wrap(mediaHandler.ConfigUpdate()))
			mediaRouter.POST("/config/toggle", wrap(mediaHandler.ConfigToggle()))
			mediaRouter.POST("/clear", wrap(mediaHandler.Clear()))
			mediaRouter.POST("/rebuild_strm_file", wrap(mediaHandler.RebuildStrmFile()))
		}
	}

	{
		staticFS, ok := embed.StaticFS()
		if ok {
			assetsFS, _ := fs.Sub(staticFS, "assets")

			engine.StaticFS("/assets", http.FS(assetsFS))

			engine.NoRoute(func(c *gin.Context) {
				if strings.HasPrefix(c.Request.URL.Path, "/api") {
					c.Status(404)

					return
				}

				// 返回 index.html
				file, err := staticFS.Open("index.html")
				if err != nil {
					c.Status(404)

					return
				}

				defer func() {
					_ = file.Close()
				}()

				stat, _ := file.Stat()

				c.Header("Content-Type", "text/html")
				c.DataFromReader(200, stat.Size(), "text/html", file, nil)
			})
		}
	}
}
