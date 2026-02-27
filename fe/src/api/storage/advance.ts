import { api, type ApiResponse } from '@/utils/api'

// ===== 存储高级功能相关接口 =====

// 文件节点接口
export interface FileNode {
  id: string
  name: string
  parentId: string
  isFolder: number
}

// 分享资源信息接口
export interface ShareResourceInfo {
  id: string
  name: string
  shareId: number
  shareTime: string
  userId: string
  isFolder: boolean
  accessCode: string
}

// 家庭云信息接口
export interface FamilyInfo {
  familyId: string
  remarkName: string
  createTime: string
  expireTime: string
  count: number
  type: number
  useFlag: number
  userRole: number
}

// 获取家庭云列表响应接口
export interface GetFamilyListResponse {
  familyInfoResp: FamilyInfo[]
}

// 获取个人文件列表查询参数
export interface GetPersonFilesQuery {
  pageNum: number // 页码，从1开始
  pageSize: number // 每页数量，最大100
  cloudToken: number // 云盘令牌ID
  parentId?: string // 父目录ID，默认为-11（根目录）
}

// 获取个人文件列表响应接口
export type GetPersonFilesResponse = Models.PaginationResponse<FileNode>

// 获取家庭云文件列表查询参数
export interface GetFamilyFilesQuery {
  pageNum: number // 页码，从1开始
  pageSize: number // 每页数量，最大100
  cloudToken: number // 云盘令牌ID
  familyId: string // 家庭云ID
  parentId?: string // 父目录ID，默认为空（根目录）
}

// 获取家庭云文件列表响应接口
export type GetFamilyFilesResponse = Models.PaginationResponse<FileNode>

// 获取订阅用户资源列表查询参数
export interface GetSubscribeUserQuery {
  subscribeUser: string // 订阅用户名
  name?: string // 文件名搜索
  currentPage?: number // 当前页码，默认为1
  pageSize?: number // 每页大小，默认为10，最大100
}

// 获取订阅用户资源列表响应接口
export interface GetSubscribeUserResponse extends Models.PaginationResponse<ShareResourceInfo> {
  name: string // 订阅用户名
}

// 获取订阅用户所有资源查询参数
export interface GetSubscribeUserAllQuery {
  subscribeUser: string // 订阅用户名
}

// 获取订阅用户所有资源响应接口
export interface GetSubscribeUserAllResponse {
  name: string // 订阅用户名
  total: number
  data: ShareResourceInfo[]
}

// 分享信息接口
export interface ShareInfo {
  id: string
  name: string
  shareId: number
  shareTime: string
  isFolder: boolean
  accessCode: string
}

// 获取分享信息查询参数
export interface GetShareInfoQuery {
  shareCode: string // 分享码
  shareAccessCode?: string // 分享访问码（可选）
}

// 获取家庭云列表查询参数
export interface GetFamilyListQuery {
  cloudToken: number // 云盘令牌ID
}

// ===== 存储高级功能接口 =====

// 获取个人文件列表
export const getPersonFiles = (
  params: GetPersonFilesQuery
): Promise<ApiResponse<GetPersonFilesResponse>> => {
  return api.get('/storage/advance/person/files', { params }).then((res) => res.data)
}

// 获取家庭云列表
export const getFamilyList = (
  params: GetFamilyListQuery
): Promise<ApiResponse<GetFamilyListResponse>> => {
  return api.get('/storage/advance/family/list', { params }).then((res) => res.data)
}

// 获取家庭云文件列表
export const getFamilyFiles = (
  params: GetFamilyFilesQuery
): Promise<ApiResponse<GetFamilyFilesResponse>> => {
  return api.get('/storage/advance/family/files', { params }).then((res) => res.data)
}

// 获取订阅用户资源列表
export const getSubscribeUser = (
  params: GetSubscribeUserQuery
): Promise<ApiResponse<GetSubscribeUserResponse>> => {
  return api.get('/storage/advance/get_subscribe_user', { params }).then((res) => res.data)
}

// 获取订阅用户所有资源（不分页）
export const getSubscribeUserAll = (
  params: GetSubscribeUserAllQuery
): Promise<ApiResponse<GetSubscribeUserAllResponse>> => {
  return api.get('/storage/advance/get_subscribe_user_all', { params }).then((res) => res.data)
}

// 获取分享信息
export const getShareInfo = (params: GetShareInfoQuery): Promise<ApiResponse<ShareInfo>> => {
  return api.get('/storage/advance/share_info', { params }).then((res) => res.data)
}
