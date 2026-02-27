package cloudbridge

import (
	"github.com/xxcheng123/cloudpan189-interface/client"
	"github.com/xxcheng123/cloudpan189-share/internal/bootstrap"
	"github.com/xxcheng123/cloudpan189-share/internal/framework/context"
	"github.com/xxcheng123/cloudpan189-share/internal/types/converter"
)

type Service interface {
	// PersonFileList 获取个人文件列表
	PersonFileList(ctx context.Context, token client.AuthToken, parentId string, pageNum, pageSize int) (*PersonFileListResponse, error)
	// PersonFileCount 获取个人文件总数
	PersonFileCount(ctx context.Context, token client.AuthToken, parentId string) (int64, error)
	// FamilyFileList 获取家庭云文件列表
	FamilyFileList(ctx context.Context, token client.AuthToken, familyId, parentId string, pageNum, pageSize int) (*FamilyFileListResponse, error)
	// FamilyFileCount 获取家庭云文件总数
	FamilyFileCount(ctx context.Context, token client.AuthToken, familyId, parentId string) (int64, error)
	// FamilyList 获取家庭云列表
	FamilyList(ctx context.Context, token client.AuthToken) (*GetFamilyListResponse, error)

	GetSubscribeUserFiles(ctx context.Context, userId string) ([]converter.VirtualFileConverter, error)
	GetSubscribeShareFiles(ctx context.Context, upUserId string, shareId int64, fileId string, isFolder bool) ([]converter.VirtualFileConverter, error)
	GetShareFiles(ctx context.Context, shareId int64, fileId string, shareMode int, accessCode string, isFolder bool) ([]converter.VirtualFileConverter, error)
	GetCloudFiles(ctx context.Context, cc AuthToken, fileId string) ([]converter.VirtualFileConverter, error)
	GetCloudFamilyFiles(ctx context.Context, cc AuthToken, familyId string, fileId string) ([]converter.VirtualFileConverter, error)

	CheckSubscribeUser(ctx context.Context, subscribeUser string) (string, error)
	CheckSubscribeShare(ctx context.Context, subscribeUser, shareCode string) (shareId int64, isFolder bool, fileId string, err error)
	CheckShare(ctx context.Context, shareCode string, accessCode string) (result *CheckShareResult, err error)
	CheckPerson(ctx context.Context, token AuthToken, fileId string) (string, error)
	CheckFamily(ctx context.Context, token AuthToken, familyId, fileId string) error

	PersonDownloadLink(ctx context.Context, token AuthToken, fileId string) (string, error)
	FamilyDownloadLink(ctx context.Context, token AuthToken, familyId, fileId string) (string, error)
	ShareDownloadLink(ctx context.Context, token AuthToken, shareId int64, fileId string) (string, error)

	GetSubscribeUserInfo(ctx context.Context, userId string) (*SubscribeUserInfo, error)
	GetSubscribeUserShareResource(ctx context.Context, userId string, opts ...SubscribeUserShareResourceOptionFunc) ([]*ShareResourceInfo, int64, error)
	GetSubscribeUserShareResourceAll(ctx context.Context, userId string) ([]*ShareResourceInfo, int64, error)
	GetShareInfo(ctx context.Context, shareCode string, accessCode string) (*ShareInfo, error)
}

type service struct {
	svc    bootstrap.ServiceContext
	client client.Client
}

func NewService(svc bootstrap.ServiceContext) Service {
	return &service{
		svc:    svc,
		client: client.New(),
	}
}

func (s *service) getClient(ctx context.Context) client.Client {
	return client.New().WithClient(ctx.HTTPClient())
}
