package cloudbridge

import (
	"errors"
	"fmt"
	"time"

	"github.com/xxcheng123/cloudpan189-interface/client"
	"github.com/xxcheng123/cloudpan189-share/internal/framework/context"
	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/utils"
	"go.uber.org/zap"
)

type SubscribeUserInfo struct {
	ID     int64  `json:"id"`
	UserId string `json:"userId"`
	Name   string `json:"name"`
}

func (s *service) GetSubscribeUserInfo(ctx context.Context, userId string) (*SubscribeUserInfo, error) {
	if info, err := s.getClient(ctx).SubscribeGetUser(ctx, userId); err != nil {
		ctx.Error("查询订阅用户信息失败", zap.String("userId", userId), zap.Error(err))

		return nil, err
	} else {
		return &SubscribeUserInfo{
			ID:     info.Data.Id,
			UserId: info.Data.UserId,
			Name:   info.Data.Name,
		}, nil
	}
}

type SubscribeUserShareResourceOption struct {
	PageNum  int
	PageSize int
	FileName string
}

type SubscribeUserShareResourceOptionFunc func(opt *SubscribeUserShareResourceOption)

type ShareResourceInfo struct {
	UserId     string    `json:"userId"`
	Name       string    `json:"name"`
	IsFolder   bool      `json:"isFolder"`
	AccessCode string    `json:"accessCode"`
	ShareId    int64     `json:"shareId"`
	ID         string    `json:"id"`
	ShareTime  time.Time `json:"shareTime"`
	IsTop      int       `json:"isTop"`
}

func (s *service) GetSubscribeUserShareResource(ctx context.Context, userId string, opts ...SubscribeUserShareResourceOptionFunc) ([]*ShareResourceInfo, int64, error) {
	option := &SubscribeUserShareResourceOption{
		PageNum:  1,
		PageSize: 30,
	}

	for _, opt := range opts {
		opt(option)
	}

	resp, err := s.getClient(ctx).GetUpResourceShare(ctx, userId, int64(option.PageNum), int64(option.PageSize), func(req *client.GetUpResourceShareRequest) {
		req.FileName = option.FileName
	})
	if err != nil {
		ctx.Error("获取订阅号下级分享失败", zap.String("user_id", userId), zap.Error(err))

		return nil, 0, err
	}

	if resp == nil || resp.Data == nil {
		ctx.Error("获取订阅号下级分享返回为空", zap.String("user_id", userId))

		return nil, 0, errors.New("获取订阅号下级分享返回为空")
	}

	list := make([]*ShareResourceInfo, 0)

	for _, item := range resp.Data.FileList {
		shareTime, _ := time.Parse(time.DateTime, item.ShareDate)

		list = append(list, &ShareResourceInfo{
			UserId:     userId,
			Name:       utils.SanitizeFileName(item.Name),
			IsFolder:   item.Folder == 1,
			AccessCode: item.AccessURL,
			ShareId:    item.ShareId,
			ID:         fmt.Sprint(item.Id),
			ShareTime:  shareTime,
			IsTop:      item.IsTop,
		})
	}

	return list, resp.Data.Count, nil
}

// GetSubscribeUserShareResourceAll 获取订阅用户全部分享资源（分页获取全部）
func (s *service) GetSubscribeUserShareResourceAll(ctx context.Context, userId string) ([]*ShareResourceInfo, int64, error) {
	pageSize := int64(100)
	allList := make([]*ShareResourceInfo, 0)
	var totalCount int64

	ctx.Info("开始获取订阅号全部分享", zap.String("user_id", userId), zap.Int64("page_size", pageSize))

	for pageNum := int64(1); ; pageNum++ {
		resp, err := s.getClient(ctx).GetUpResourceShare(ctx, userId, pageNum, pageSize, func(req *client.GetUpResourceShareRequest) {})
		if err != nil {
			ctx.Error("获取订阅号下级分享失败", zap.String("user_id", userId), zap.Int64("page_num", pageNum), zap.Error(err))
			return nil, 0, err
		}

		ctx.Info("订阅号分享API原始响应", zap.String("user_id", userId), zap.Int64("page_num", pageNum), zap.Any("resp", resp))

		if resp == nil || resp.Data == nil {
			ctx.Error("获取订阅号下级分享返回数据为空", zap.String("user_id", userId), zap.Int64("page_num", pageNum))
			return allList, totalCount, nil
		}

		// 只在第一页获取总数
		if pageNum == 1 {
			totalCount = resp.Data.Count
			ctx.Info("订阅号分享总数", zap.String("user_id", userId), zap.Int64("total", totalCount), zap.Int("file_list_len", len(resp.Data.FileList)))
		}

		// 如果当前页为空，退出循环
		if len(resp.Data.FileList) == 0 {
			ctx.Info("获取订阅号下级分享当前页为空", zap.String("user_id", userId), zap.Int64("page_num", pageNum))
			break
		}

		ctx.Info("获取订阅号分享分页数据", zap.String("user_id", userId), zap.Int64("page_num", pageNum), zap.Int("current_page_count", len(resp.Data.FileList)))

		for _, item := range resp.Data.FileList {
			shareTime, _ := time.Parse(time.DateTime, item.ShareDate)

			allList = append(allList, &ShareResourceInfo{
				UserId:     userId,
				Name:       utils.SanitizeFileName(item.Name),
				IsFolder:   item.Folder == 1,
				AccessCode: item.AccessURL,
				ShareId:    item.ShareId,
				ID:         fmt.Sprint(item.Id),
				ShareTime:  shareTime,
				IsTop:      item.IsTop,
			})
		}

		// 如果当前页数据量 < pageSize，说明已经是最后一页
		if len(resp.Data.FileList) < int(pageSize) {
			break
		}
	}

	ctx.Info("获取订阅号全部分享完成", zap.String("user_id", userId), zap.Int64("total_count", totalCount), zap.Int("actual_count", len(allList)))

	return allList, totalCount, nil
}
