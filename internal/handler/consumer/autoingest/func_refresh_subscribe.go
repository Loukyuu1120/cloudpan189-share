package autoingest

import (
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/xxcheng123/cloudpan189-share/internal/consts"
	"github.com/xxcheng123/cloudpan189-share/internal/framework/taskcontext"
	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/datatypes"
	"github.com/xxcheng123/cloudpan189-share/internal/repository/models"
	"github.com/xxcheng123/cloudpan189-share/internal/types/autoingest"
	"github.com/xxcheng123/cloudpan189-share/internal/types/topic"
	"go.uber.org/zap"
	"gorm.io/gorm"

	cloudbridgeSvi "github.com/xxcheng123/cloudpan189-share/internal/services/cloudbridge"
	storagefacadeSvi "github.com/xxcheng123/cloudpan189-share/internal/services/storagefacade"
)

const (
	retryDelayBase = 100 * time.Millisecond
)

func (h *handler) RefreshSubscribe() taskcontext.HandlerFunc {
	return func(ctx *taskcontext.Context) error {
		req := new(topic.AutoIngestRefreshSubscribeRequest)

		if err := ctx.Unmarshal(req); err != nil {
			return err
		}

		var (
			pageSize   = 100
			pageNum    = 1
			hasTop     = true
			shouldNext = true

			addCount    int64 = 0
			failedCount int64 = 0
		)

		logger := ctx.GetContext().Logger

		plan, err := h.autoIngestPlanService.Query(ctx.GetContext(), req.PlanId)
		if err != nil {
			logger.Error("查询入库计划信息失败", zap.Error(err), zap.Int64("plan_id", req.PlanId))

			return err
		}

		concurrentCount := plan.ConcurrentCount
		if concurrentCount <= 0 {
			concurrentCount = 4
		}
		maxRetryCount := plan.MaxRetryCount
		if maxRetryCount <= 0 {
			maxRetryCount = 3
		}

		var nextOffset = plan.Offset

		if plan.SourceType != autoingest.SourceTypeSubscribe {
			logger.Error("计划类型错误", zap.String("source_type", plan.SourceType.String()))

			return errors.New("计划类型错误")
		}

		addition := new(models.AutoIngestPlanSubscribeAddition)
		_ = plan.Addition.Unmarshal(addition)

		defer func() {
			_ = h.autoIngestPlanService.UpdateOffset(ctx.GetContext(), req.PlanId, nextOffset)
			_ = h.autoIngestPlanService.IncrAddCount(ctx.GetContext(), req.PlanId, addCount)
			_ = h.autoIngestPlanService.IncrFailedCount(ctx.GetContext(), req.PlanId, failedCount)
		}()

		type pendingItem struct {
			item       *cloudbridgeSvi.ShareResourceInfo
			itemOffset int64
		}

		var pendingItems []pendingItem

		for shouldNext {
			list, _, err := h.cloudbridgeService.GetSubscribeUserShareResource(ctx.GetContext(), addition.UpUserId, func(opt *cloudbridgeSvi.SubscribeUserShareResourceOption) {
				opt.PageNum = pageNum
				opt.PageSize = pageSize
			})
			if err != nil {
				logger.Error("获取订阅号内容时失败了~", zap.String("up_user_id", addition.UpUserId), zap.Int("page_num", pageNum), zap.Int("page_size", pageSize))

				return err
			}

			if len(list) == 0 {
				break
			}

			pageNum++

			for _, item := range list {
				if item.IsTop != 1 {
					hasTop = false
				}

				itemOffset := item.ShareTime.Unix()
				if itemOffset > nextOffset {
					nextOffset = itemOffset
				}

				if itemOffset <= plan.Offset {
					if !hasTop {
						shouldNext = false
					}

					continue
				}

				logger.Debug("发现新的待入库文件", zap.String("name", item.Name))

				pendingItems = append(pendingItems, pendingItem{
					item:       item,
					itemOffset: itemOffset,
				})
			}
		}

		if len(pendingItems) == 0 {
			return nil
		}

		logger.Info("开始并发入库", zap.Int("total", len(pendingItems)), zap.Int("concurrent", concurrentCount))

		var (
			wg           sync.WaitGroup
			itemChan     = make(chan pendingItem, len(pendingItems))
			mu           sync.Mutex
			localAddCount    int64
			localFailedCount int64
		)

		for _, item := range pendingItems {
			itemChan <- item
		}
		close(itemChan)

		for i := 0; i < concurrentCount; i++ {
			wg.Add(1)
			go func(workerId int) {
				defer wg.Done()

				for pItem := range itemChan {
					fullPath := path.Join(plan.ParentPath, pItem.item.Name)

					for retry := 0; retry <= maxRetryCount; retry++ {
						if retry > 0 {
							time.Sleep(retryDelayBase * time.Duration(retry))
						}

						exists, err := h.virtualFileService.QueryByPath(ctx.GetContext(), fullPath)
						if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
							logger.Error("查询虚拟文件路径失败 跳过本次自动入库", zap.String("path", fullPath), zap.Error(err))
							break
						}

						if exists != nil {
							if plan.OnConflict == autoingest.OnConflictRename {
								fullPath = path.Join(plan.ParentPath, fmt.Sprintf("%s_%d", pItem.item.Name, time.Now().Unix()))
								continue
							}
							break
						}

						id, err := h.storageFacadeService.CreateStorage(ctx.GetContext(),
							&storagefacadeSvi.CreateStorageRequest{
								LocalPath:  fullPath,
								OsType:     models.OsTypeSubscribeShareFolder,
								CloudToken: plan.TokenId,
								FileId:     pItem.item.ID,
								Addition: datatypes.JSONMap{
									consts.FileAdditionKeyUpUserId: addition.UpUserId,
									consts.FileAdditionKeyShareId:  pItem.item.ShareId,
									consts.FileAdditionKeyIsFolder: pItem.item.IsFolder,
								},
								EnableAutoRefresh: plan.RefreshStrategy.EnableAutoRefresh,
								EnableDeepRefresh: plan.RefreshStrategy.EnableDeepRefresh,
								AutoRefreshDays:   plan.RefreshStrategy.AutoRefreshDays,
								RefreshInterval:   plan.RefreshStrategy.RefreshInterval,
							},
						)
						if err != nil {
							if strings.Contains(err.Error(), "UNIQUE constraint failed") || strings.Contains(err.Error(), "Duplicate entry") {
								break
							}
							if retry < maxRetryCount {
								logger.Warn("入库失败，正在重试", zap.String("path", fullPath), zap.Error(err), zap.Int("retry", retry+1))
								continue
							}

							logger.Error("入库失败", zap.Error(err), zap.String("path", fullPath))
							mu.Lock()
							localFailedCount++
							mu.Unlock()

							if _, logErr := h.authIngestLogService.Create(ctx.GetContext(),
								req.PlanId, autoingest.LogLevelError,
								fmt.Sprintf("新增入库失败：%s, 错误信息：%s", fullPath, err.Error()),
							); logErr != nil {
								logger.Error("创建入库日志失败", zap.Error(logErr))
							}

							break
						}

						taskReq := &topic.FileScanFileRequest{
							FileId: id,
							Deep:   true,
						}

						body, _ := json.Marshal(taskReq)
						if err = h.taskEngine.PushMessage(
							ctx.GetContext().
								WithValue(consts.CtxKeyFullPath, fullPath).
								WithValue(consts.CtxKeyInvokeHandlerName, "入库执行器"),
							taskReq.Topic(), body); err != nil {
							logger.Error("下发文件扫描任务失败", zap.Error(err))
						}

						if _, logErr := h.authIngestLogService.Create(ctx.GetContext(),
							req.PlanId, autoingest.LogLevelInfo,
							fmt.Sprintf("新增入库：%s", fullPath),
						); logErr != nil {
							logger.Error("创建入库日志失败", zap.Error(logErr))
						}

						mu.Lock()
						localAddCount++
						mu.Unlock()

						break
					}
				}
			}(i)
		}

		wg.Wait()

		addCount = localAddCount
		failedCount = localFailedCount

		logger.Info("并发入库完成", zap.Int64("add_count", addCount), zap.Int64("failed_count", failedCount))

		return nil
	}
}
