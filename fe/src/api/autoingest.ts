import { api, type ApiResponse } from '@/utils/api'

// ===== 自动挂载（Auto Ingest）相关接口 =====

// 公共类型
export type AutoIngestLogLevel = 'info' | 'warn' | 'error'
export type AutoIngestOnConflict = 'rename' | 'abandon'

// 分页查询参数（通用）
export interface PaginationQuery {
  currentPage?: number // 当前页码，默认 1
  pageSize?: number // 每页大小，默认 10
}

// ===== 计划管理 =====

// 获取计划列表 - 查询参数
export interface PlanListQuery extends PaginationQuery {
  name?: string // 按名称模糊搜索
}

export interface PlanLogResult extends Models.AutoIngestLog {
  planName: string // 计划名称
}

// 创建订阅型计划 - 刷新策略
export interface RefreshStrategyRequest {
  enableAutoRefresh?: boolean
  autoRefreshDays?: number // >= 1
  refreshInterval?: number // 单位分钟，>= 30
  enableDeepRefresh?: boolean
}

// 创建订阅型计划 - 请求体
export interface CreateSubscribePlanRequest {
  name: string // 计划名称
  parentPath: string // 挂载父目录路径，例如：/Movies
  upUserId: string // 上传用户ID
  cloudToken?: number // 云盘令牌ID（可选）
  onConflict?: AutoIngestOnConflict // 冲突时处理策略
  autoIngestInterval?: number // 自动挂载间隔，单位分钟（>=5）
  oneClickAddHistory?: boolean // 是否一键添加历史
  refreshStrategy?: RefreshStrategyRequest // 刷新策略（可选）
}

// 创建订阅型计划 - 响应体
export interface CreateSubscribePlanResponse {
  id: number
}

// 更新计划 - 请求体（仅允许以下字段）
export interface UpdatePlanRequest {
  id: number
  name?: string
  autoIngestInterval?: number
  parentPath?: string
  onConflict?: AutoIngestOnConflict
  tokenId?: number
  refreshStrategy?: RefreshStrategyRequest
}

// ===== 日志管理 =====

// 获取日志列表 - 查询参数
export interface LogListQuery extends PaginationQuery {
  planId?: number // 计划ID
  level?: AutoIngestLogLevel // 日志级别：info/warn/error
}

// ===== 自动挂载管理接口 =====

// 获取自动挂载计划列表
export const getAutoIngestPlanList = (
  params?: PlanListQuery
): Promise<ApiResponse<Models.PaginationResponse<Models.AutoIngestPlan>>> => {
  return api.get('/auto_ingest/plan/list', { params }).then((res) => res.data)
}

// 创建订阅型自动挂载计划
export const createSubscribePlan = (
  data: CreateSubscribePlanRequest
): Promise<ApiResponse<CreateSubscribePlanResponse>> => {
  return api.post('/auto_ingest/plan/create_subscribe', data).then((res) => res.data)
}

// 启用自动挂载计划
export const enableAutoIngestPlan = (data: { id: number }): Promise<ApiResponse> => {
  return api.post('/auto_ingest/plan/enable', data).then((res) => res.data)
}

// 停用自动挂载计划
export const disableAutoIngestPlan = (data: { id: number }): Promise<ApiResponse> => {
  return api.post('/auto_ingest/plan/disable', data).then((res) => res.data)
}

// 手动触发订阅计划刷新
export const refreshAutoIngestPlan = (data: { planId: number }): Promise<ApiResponse> => {
  return api.post('/auto_ingest/plan/refresh', data).then((res) => res.data)
}

// 删除自动挂载计划
export const deleteAutoIngestPlan = (data: { id: number }): Promise<ApiResponse> => {
  return api.post('/auto_ingest/plan/delete', data).then((res) => res.data)
}

// 修改自动挂载计划（仅允许部分字段）
export const updateAutoIngestPlan = (data: UpdatePlanRequest): Promise<ApiResponse> => {
  return api.post('/auto_ingest/plan/update', data).then((res) => res.data)
}

// 获取自动挂载日志列表
export const getAutoIngestLogList = (
  params?: LogListQuery
): Promise<ApiResponse<Models.PaginationResponse<PlanLogResult>>> => {
  return api.get('/auto_ingest/log/list', { params }).then((res) => res.data)
}

// 重试失败的任务
export const retryFailedAutoIngest = (data: { planId?: number }): Promise<ApiResponse> => {
  return api.post('/auto_ingest/plan/retry_failed', data).then((res) => res.data)
}

// 删除错误日志
export const deleteErrorLogs = (data: { planId?: number }): Promise<ApiResponse<number>> => {
  return api.post('/auto_ingest/log/delete_error', data).then((res) => res.data)
}
