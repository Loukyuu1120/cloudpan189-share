package file

import (
	"fmt"
	"time"

	"github.com/xxcheng123/cloudpan189-share/internal/framework/taskcontext"
	"github.com/xxcheng123/cloudpan189-share/internal/services/filetasklog"
	"github.com/xxcheng123/cloudpan189-share/internal/services/virtualfile"
	"github.com/xxcheng123/cloudpan189-share/internal/shared"
	"github.com/xxcheng123/cloudpan189-share/internal/types/topic"
	"go.uber.org/zap"
)

// HandleBatchDelete 后台排队删除处理逻辑
func (h *handler) HandleBatchDelete() taskcontext.HandlerFunc {
	return func(ctx *taskcontext.Context) error {
		req := new(topic.FileBatchDeleteRequest)

		if err := ctx.Unmarshal(req); err != nil {
			h.logger.Error("解析删除任务失败", zap.Error(err))
			return nil
		}

		h.logger.Info("消费者开始处理批量删除", zap.Int("count", len(req.IDs)))

		// 创建任务日志
		tracker, logErr := h.fileTaskLogService.Create(
			ctx.GetContext(),
			req.Topic().String(),
			fmt.Sprintf("批量删除 %d 个挂载点", len(req.IDs)),
			filetasklog.WithFile(0),
			filetasklog.WithDesc(fmt.Sprintf("批量删除任务, 共 %d 个挂载点", len(req.IDs))),
		)
		if logErr != nil {
			h.logger.Error("创建任务日志失败", zap.Error(logErr))
		} else {
			_ = h.fileTaskLogService.Running(ctx.GetContext(), tracker)
		}

		defer func() {
			if tracker != nil {
				if err := h.fileTaskLogService.Completed(ctx.GetContext(), tracker, tracker.WithCost()); err != nil {
					h.logger.Error("更新任务日志失败", zap.Error(err))
				}
			}
		}()

		for _, id := range req.IDs {
			targetFileID := id
			fileInfo, fileErr := h.virtualFileService.Query(ctx.GetContext(), targetFileID)

			// 1. 尝试删除挂载点记录
			if err := h.mountPointService.BatchDelete(ctx.GetContext(), []int64{id}); err != nil {
				h.logger.Debug("后台删除挂载点记录异常(或已删除)", zap.Int64("id", id), zap.Error(err))
			}

			// 如果查不到文件信息，说明可能已经被删了，但仍尝试清理残留的子文件
			if fileErr != nil || fileInfo == nil {
				_ = h.clearMountFiles(ctx.GetContext(), targetFileID)
				continue
			}

			// 记录父ID，用于稍后递归清理
			parentId := fileInfo.ParentId

			// 2. 清理挂载点下的子文件 (触发 deleteStrmIterator)
			if err := h.clearMountFiles(ctx.GetContext(), targetFileID); err != nil {
				h.logger.Error("清理挂载点子文件失败", zap.Int64("fid", targetFileID), zap.Error(err))
			}

			// 3. 删除当前的根虚拟文件
			if err := h.virtualFileService.Delete(ctx.GetContext(), targetFileID); err != nil {
				h.logger.Error("删除根虚拟文件失败", zap.Int64("fid", targetFileID), zap.Error(err))
			} else {
				h.logger.Info("后台删除虚拟文件完成", zap.Int64("fid", targetFileID))

				// 4. 本地空目录清理 (确保 strm 删完后执行)
				if shared.MediaConfig != nil && shared.MediaConfig.Enable {
					if err := h.mediaFileService.ClearEmptyDir(ctx.GetContext(), shared.MediaConfig.StoragePath); err != nil {
						h.logger.Warn("清理本地空目录失败", zap.Error(err))
					}
				}

				// 5. [暴力递归] 手动清理数据库中的空祖先目录
				// 逻辑：不依赖 ClearUnusedAncestorFolder，直接查子节点数量
				scanPid := parentId
				for scanPid > 0 {
					// A. 查询当前目录信息（为了拿下一级的父ID）
					pInfo, err := h.virtualFileService.Query(ctx.GetContext(), scanPid)
					if err != nil || pInfo == nil {
						break // 目录不存在，停止
					}
					nextPid := pInfo.ParentId // 记下爷爷ID

					// B. 检查当前目录是否还有子节点
					// 我们只查1个，只要有1个就说明不为空
					children, _ := h.virtualFileService.List(ctx.GetContext(), &virtualfile.ListRequest{
						ParentId:    &scanPid,
						CurrentPage: 1,
						PageSize:    1,
					})

					if len(children) > 0 {
						// 还有孩子，不能删，且再往上肯定也不为空，直接停止
						h.logger.Debug("数据库目录不为空，停止向上清理", zap.Int64("pid", scanPid))
						break
					}

					// C. 没有子节点，直接删除
					if err := h.virtualFileService.Delete(ctx.GetContext(), scanPid); err != nil {
						h.logger.Warn("删除数据库空目录失败", zap.Int64("pid", scanPid), zap.Error(err))
						break
					}

					h.logger.Info("成功清理数据库空目录", zap.Int64("pid", scanPid), zap.String("name", pInfo.Name))

					// D. 继续处理上一级
					scanPid = nextPid
				}
			}
			time.Sleep(20 * time.Millisecond)
		}

		return nil
	}
}
