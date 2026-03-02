// 任务类型常量定义
// 对应后端 internal/types/topic/consts.go

export const TASK_TYPES = {
  FILE_SCAN: 'topic::file::scan::file',
  FILE_CLEAR: 'topic::file::clear::file',
  FILE_BATCH_DELETE: 'topic::file::batch_delete::file',
} as const

// 任务类型选项配置
export const TASK_TYPE_OPTIONS = [
  { label: '文件扫描', value: TASK_TYPES.FILE_SCAN },
  { label: '文件清空', value: TASK_TYPES.FILE_CLEAR },
  { label: '批量删除', value: TASK_TYPES.FILE_BATCH_DELETE },
]

// 任务类型显示文本映射
export const TASK_TYPE_TEXT_MAP = {
  [TASK_TYPES.FILE_SCAN]: '文件扫描',
  [TASK_TYPES.FILE_CLEAR]: '文件清空',
  [TASK_TYPES.FILE_BATCH_DELETE]: '批量删除',
} as const

// 任务类型标签类型映射
export const TASK_TYPE_TAG_MAP = {
  [TASK_TYPES.FILE_SCAN]: 'info',
  [TASK_TYPES.FILE_CLEAR]: 'warning',
  [TASK_TYPES.FILE_BATCH_DELETE]: 'error',
} as const

// 任务状态常量定义
export const TASK_STATUS = {
  PENDING: 'pending',
  RUNNING: 'running',
  COMPLETED: 'completed',
  FAILED: 'failed',
} as const

// 任务状态选项配置
export const TASK_STATUS_OPTIONS = [
  { label: '待处理', value: TASK_STATUS.PENDING },
  { label: '运行中', value: TASK_STATUS.RUNNING },
  { label: '已完成', value: TASK_STATUS.COMPLETED },
  { label: '失败', value: TASK_STATUS.FAILED },
]

// 任务状态显示文本映射
export const TASK_STATUS_TEXT_MAP = {
  [TASK_STATUS.PENDING]: '待处理',
  [TASK_STATUS.RUNNING]: '运行中',
  [TASK_STATUS.COMPLETED]: '已完成',
  [TASK_STATUS.FAILED]: '失败',
} as const

// 任务状态标签类型映射
export const TASK_STATUS_TAG_MAP = {
  [TASK_STATUS.PENDING]: 'warning',
  [TASK_STATUS.RUNNING]: 'info',
  [TASK_STATUS.COMPLETED]: 'success',
  [TASK_STATUS.FAILED]: 'error',
} as const

// 任务类型定义
export type TaskType = (typeof TASK_TYPES)[keyof typeof TASK_TYPES]

// 任务状态定义
export type TaskStatus = (typeof TASK_STATUS)[keyof typeof TASK_STATUS]
