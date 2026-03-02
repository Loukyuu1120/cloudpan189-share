package mountpoint

import (
	"regexp"
	"strings"
	"time"

	"github.com/xxcheng123/cloudpan189-share/internal/bootstrap"
	"github.com/xxcheng123/cloudpan189-share/internal/framework/context"
	"github.com/xxcheng123/cloudpan189-share/internal/repository/models"

	cloudbridgeSvi "github.com/xxcheng123/cloudpan189-share/internal/services/cloudbridge"
	cloudtokenSvi "github.com/xxcheng123/cloudpan189-share/internal/services/cloudtoken"
	"github.com/xxcheng123/cloudpan189-share/internal/types/topic"

	"gorm.io/gorm"
)

type Service interface {
	Create(ctx context.Context, req *CreateRequest) (int64, error)
	Query(ctx context.Context, fileId int64) (*models.MountPoint, error)
	QueryByPath(ctx context.Context, fullPath string) (*models.MountPoint, error)
	List(ctx context.Context, req *ListRequest) ([]*models.MountPoint, error)
	Count(ctx context.Context, req *ListRequest) (int64, error)
	Delete(ctx context.Context, fileId int64) error
	BatchDelete(ctx context.Context, ids []int64) error
	EnableAutoRefresh(ctx context.Context, fileId int64, enable bool) error
	GetAutoRefreshList(ctx context.Context, req *GetAutoRefreshListRequest) ([]*models.MountPoint, error)
	UpdateRefreshConfig(ctx context.Context, fileId int64, config RefreshConfig) error
	ModifyToken(ctx context.Context, fid int64, tokenId int64) error
	BatchParseText(ctx context.Context, req *topic.BatchParseTextRequest) ([]*topic.BatchParseItem, error)
	UpdateRefreshTime(ctx context.Context, fileId int64) error
}

type service struct {
	svc                bootstrap.ServiceContext
	cloudTokenService  cloudtokenSvi.Service
	cloudBridgeService cloudbridgeSvi.Service
}

func NewService(
	svc bootstrap.ServiceContext,
	cloudTokenService cloudtokenSvi.Service,
	cloudBridgeService cloudbridgeSvi.Service,
) Service {
	return &service{
		svc:                svc,
		cloudTokenService:  cloudTokenService,
		cloudBridgeService: cloudBridgeService,
	}
}

// UpdateRefreshTime 更新挂载点的刷新时间
func (s *service) UpdateRefreshTime(ctx context.Context, fileId int64) error {
	return s.getDB(ctx).Where("file_id = ?", fileId).Update("updated_at", time.Now()).Error
}

func (s *service) getDB(ctx context.Context) *gorm.DB {
	return s.svc.GetDB(ctx).Model(new(models.MountPoint))
}

var (
	reFolderID      = regexp.MustCompile(`^\d+$`)
	reShareLink     = regexp.MustCompile(`cloud\.189\.cn\/t\/([a-zA-Z0-9]+)`)
	reAccessCode    = regexp.MustCompile(`(?:\S+码|code)[:：]\s*([a-zA-Z0-9]+)`)
	reSubscribeLink = regexp.MustCompile(`content\.21cn\.com.*[?&]uuid=([a-zA-Z0-9]+)`)
)

// 实现 BatchParseText
func (s *service) BatchParseText(ctx context.Context, req *topic.BatchParseTextRequest) ([]*topic.BatchParseItem, error) {
	// 1. 获取 Token 信息 (用于 CheckPerson)
	tokenInfo, err := s.cloudTokenService.Query(ctx, req.CloudToken)
	if err != nil {
		return nil, err
	}
	// 构造 AuthToken
	authToken := cloudbridgeSvi.NewAuthToken(tokenInfo.AccessToken, tokenInfo.ExpiresIn)

	var results []*topic.BatchParseItem
	lines := strings.Split(req.Content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 预处理：统一中文符号
		cleanLine := strings.ReplaceAll(line, "（", "(")
		cleanLine = strings.ReplaceAll(cleanLine, "）", ")")
		cleanLine = strings.ReplaceAll(cleanLine, "：", ":")

		var (
			shareCode   string
			accessCode  string
			fileId      string
			isShare     bool
			isFolder    bool
			isSubscribe bool
			subscribeId string
		)

		// 0. 尝试匹配订阅号链接 (优先级最高)
		if matches := reSubscribeLink.FindStringSubmatch(cleanLine); len(matches) > 1 {
			subscribeId = matches[1]
			isSubscribe = true
		}

		// 1. 尝试匹配分享链接 (全行搜索)
		if !isSubscribe {
			if matches := reShareLink.FindStringSubmatch(cleanLine); len(matches) > 1 {
				shareCode = matches[1]
				isShare = true
			}
		}

		// 2. 如果不是分享链接，尝试匹配纯数字文件夹ID
		if !isShare {
			// 这里需要严谨一点，如果是纯数字或者是 "数字" 这种格式
			// 简单处理：如果是纯数字
			if reFolderID.MatchString(line) {
				fileId = line
				isFolder = true
			} else {
				// 如果是 "folder_id 12345" 这种格式，尝试提取
				parts := strings.Fields(line)
				if len(parts) > 0 && reFolderID.MatchString(parts[0]) {
					fileId = parts[0]
					isFolder = true
				}
			}
		}

		// 3. 提取访问码 (仅针对分享链接)
		if isShare {
			if codeMatch := reAccessCode.FindStringSubmatch(cleanLine); len(codeMatch) > 1 {
				accessCode = codeMatch[1]
			} else {
				parts := strings.Fields(cleanLine)
				if len(parts) > 1 {
					lastPart := strings.Trim(parts[len(parts)-1], "()")
					if len(lastPart) == 4 {
						accessCode = lastPart
					}
				}
			}
		}
		if isShare {
			info, err := s.cloudBridgeService.GetShareInfo(ctx, shareCode, accessCode)

			name := ""
			if err == nil && info != nil {
				name = info.Name
			} else {
				name = "未知分享_" + shareCode
			}

			results = append(results, &topic.BatchParseItem{
				Name:            name,
				OsType:          models.OsTypeShareFolder,
				ShareCode:       shareCode,
				ShareAccessCode: accessCode,
			})

		} else if isFolder {
			name, err := s.cloudBridgeService.CheckPerson(ctx, authToken, fileId)

			if err != nil || name == "" {
				name = "未知文件夹_" + fileId
			}

			results = append(results, &topic.BatchParseItem{
				Name:   name,
				OsType: models.OsTypePersonFolder,
				FileId: fileId,
			})
		} else if isSubscribe {
			userInfo, err := s.cloudBridgeService.GetSubscribeUserInfo(ctx, subscribeId)
			name := ""
			if err == nil && userInfo != nil {
				name = userInfo.Name
			} else {
				name = "未知订阅号_" + subscribeId
			}

			// 使用 subscribe 类型，只需 SubscribeUser，不需要 ShareCode
			results = append(results, &topic.BatchParseItem{
				Name:          name,
				OsType:        models.OsTypeSubscribe,
				SubscribeUser: subscribeId,
			})
		}
	}

	return results, nil
}
