package storagefacade

import (
	"time"

	"github.com/pkg/errors"
	"github.com/xxcheng123/cloudpan189-share/internal/consts"
	"github.com/xxcheng123/cloudpan189-share/internal/framework/context"
	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/datatypes"
	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/utils"
	"github.com/xxcheng123/cloudpan189-share/internal/repository/models"
	mountPointSvi "github.com/xxcheng123/cloudpan189-share/internal/services/mountpoint"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type CreateStorageRequest struct {
	LocalPath  string            // 逻辑路径，例如 /foo/bar
	OsType     string            // 协议类型
	CloudToken int64             // 云 token ID
	FileId     string            // 云文件/资源 ID（用于 Addition.CloudId 映射）
	Addition   datatypes.JSONMap // 额外元数据，由上层根据协议类型准备

	EnableAutoRefresh bool `json:"enableAutoRefresh"`
	AutoRefreshDays   int  `json:"autoRefreshDays"`
	RefreshInterval   int  `json:"refreshInterval"`
	EnableDeepRefresh bool `json:"enableDeepRefresh"`
}

var (
	errRequestNil         = errors.New("请求对象为空")
	errInvalidPath        = errors.New("路径不合法，需要 / 开头的路径")
	errMountPointExists   = errors.New("路径已存在")
	errVirtualFileExists  = errors.New("路径已存在")
	errRootPathNotAllowed = errors.New("不允许挂载根路径")
)

// CreateStorage 在一个统一流程中创建 VirtualFile 顶层节点与 MountPoint 记录，返回根 VirtualFile ID。
// 说明：暂不使用数据库事务（底层 service 未暴露 tx），若创建 MountPoint 失败则补偿删除刚创建的 VirtualFile。
func (s *service) CreateStorage(ctx context.Context, req *CreateStorageRequest) (int64, error) {
	// 基本校验
	if req == nil {
		return 0, errRequestNil
	}

	if !utils.CheckIsPath(req.LocalPath) {
		return 0, errInvalidPath
	}

	// 防重：MountPoint 与 VirtualFile
	if mp, err := s.mountPointService.QueryByPath(ctx, req.LocalPath); err != nil {
		ctx.Error("查询挂载点路径失败", zap.Error(err), zap.String("path", req.LocalPath))

		return 0, err
	} else if mp != nil {
		// 路径已存在，返回已存在的挂载点ID而不是报错
		ctx.Info("挂载点路径已存在，返回已存在的记录", zap.String("path", req.LocalPath), zap.Int64("exists_id", mp.ID))
		return mp.FileId, nil
	}

	if vf, err := s.virtualFileService.QueryByPath(ctx, req.LocalPath); !errors.Is(err, gorm.ErrRecordNotFound) {
		if err != nil {
			ctx.Error("查询虚拟文件路径失败", zap.Error(err), zap.String("path", req.LocalPath))

			return 0, err
		}
		// 虚拟文件已存在，返回其ID
		ctx.Info("虚拟文件路径已存在，返回已存在的记录", zap.String("path", req.LocalPath), zap.Int64("exists_id", vf.ID))
		return vf.ID, nil
	}

	// 路径分割与父级创建
	paths, err := utils.SplitPath(req.LocalPath)
	if err != nil {
		ctx.Error("路径分割失败", zap.Error(err), zap.String("path", req.LocalPath))

		return 0, err
	}

	if len(paths) == 0 {
		return 0, errRootPathNotAllowed
	}

	parentId, err := s.virtualFileService.FindOrCreateAncestors(ctx, req.LocalPath)
	if err != nil {
		ctx.Error("创建父级路径失败", zap.Error(err), zap.String("path", req.LocalPath))

		return 0, err
	}

	// 组装 VirtualFile 顶层节点
	now := time.Now()
	vf := &models.VirtualFile{
		Name:       paths[len(paths)-1],
		IsTop:      true,
		Size:       0,
		Hash:       "",
		CreateDate: now,
		ModifyDate: now,
		Rev:        now.Format(consts.RevFormat),
		OsType:     req.OsType,
		IsDir:      true,
		Addition:   req.Addition,
		CloudId:    req.FileId,
	}

	id, err := s.virtualFileService.CreateTop(ctx, parentId, vf)
	if err != nil {
		ctx.Error("创建顶层虚拟文件失败", zap.Error(err), zap.Int64("parentId", parentId))

		return 0, err
	}

	// 创建 MountPoint，失败时做补偿删除虚拟文件
	if _, err = s.mountPointService.Create(ctx, &mountPointSvi.CreateRequest{
		FullPath: req.LocalPath,
		FileId:   id,
		OsType:   req.OsType,
		TokenId:  req.CloudToken,

		EnableAutoRefresh: req.EnableAutoRefresh,
		AutoRefreshDays:   req.AutoRefreshDays,
		RefreshInterval:   req.RefreshInterval,
		EnableDeepRefresh: req.EnableDeepRefresh,
	}); err != nil {
		ctx.Error("创建挂载点失败，执行补偿删除虚拟文件", zap.Error(err), zap.Int64("virtualFileId", id), zap.String("path", req.LocalPath))

		if delErr := s.virtualFileService.Delete(ctx, id); delErr != nil {
			ctx.Error("补偿删除虚拟文件失败", zap.Error(delErr), zap.Int64("virtualFileId", id))
		}

		return 0, err
	}

	return id, nil
}
