package topic

import (
	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/taskengine"
)

type Request interface {
	Topic() taskengine.Topic
}

type FileScanFileRequest struct {
	FileId int64 `json:"fileId"` // 入口地址
	Deep   bool  `json:"deep"`   // 是否深度扫描
}

func (r FileScanFileRequest) Topic() taskengine.Topic {
	return taskengine.Topic(KeyFileScanFile)
}

type FileClearFileRequest struct {
	FileId int64 `json:"fileId"` // 入口地址
}

func (r FileClearFileRequest) Topic() taskengine.Topic {
	return taskengine.Topic(KeyFileClearFile)
}

type AutoIngestRefreshSubscribeRequest struct {
	PlanId int64 `json:"planId"`
}

func (r AutoIngestRefreshSubscribeRequest) Topic() taskengine.Topic {
	return taskengine.Topic(KeyAutoIngestRefreshSubscribe)
}

type MediaClearRequest struct{}

func (r MediaClearRequest) Topic() taskengine.Topic {
	return taskengine.Topic(KeyMediaClear)
}

type MediaRebuildStrmFileRequest struct{}

func (r MediaRebuildStrmFileRequest) Topic() taskengine.Topic {
	return taskengine.Topic(KeyMediaRebuildStrmFile)
}

// 2. 定义请求结构体
type FileBatchDeleteRequest struct {
	IDs []int64 `json:"ids"`
}

// 3. 实现接口
func (r FileBatchDeleteRequest) Topic() taskengine.Topic {
	return taskengine.Topic(KeyFileBatchDelete)
}

// 批量解析文本请求 (仅用于 API，不用于 Task)
type BatchParseTextRequest struct {
	Content    string `json:"content" binding:"required"`    // 文本内容
	CloudToken int64  `json:"cloudToken" binding:"required"` // 需要用到token去查询信息
}

// 批量解析响应项 (仅用于 API，不用于 Task)
type BatchParseItem struct {
	Name            string `json:"name"`            // 识别出的名称
	OsType          string `json:"osType"`          // 识别出的类型: share_folder 或 person_folder
	ShareCode       string `json:"shareCode"`       // 分享码
	ShareAccessCode string `json:"shareAccessCode"` // 提取码
	FileId          string `json:"fileId"`          // 文件夹ID
	SubscribeUser   string `json:"subscribeUser"`   // 订阅号用户ID
}
