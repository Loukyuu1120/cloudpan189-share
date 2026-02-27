import { reactive, computed, type Ref } from 'vue'
import { getSubscribeUser, getSubscribeUserAll } from '@/api/storage/advance'
import type { ShareResourceInfo, GetSubscribeUserResponse } from '@/api/storage/advance'
import type { ApiResponse } from '@/utils/api'
import type { MessageApi } from 'naive-ui'

// 常量定义
const PAGINATION_CONFIG = {
  DEFAULT_PAGE_SIZE: 30,
  PAGE_SIZES: [20, 30, 50, 100, 200, 500] as number[],
  DEFAULT_PAGE: 1,
}

export function useSubscribeResource(subscribeUserId: Ref<string>, message: MessageApi) {
  // 资源列表状态
  const resourceState = reactive({
    userInfo: null as GetSubscribeUserResponse | null,
    list: [] as ShareResourceInfo[],
    loading: false,
    loadingAll: false,
    searchKeyword: '',
    selected: [] as ShareResourceInfo[],
  })

  // 分页配置
  const resourcePagination = reactive({
    page: PAGINATION_CONFIG.DEFAULT_PAGE,
    pageSize: PAGINATION_CONFIG.DEFAULT_PAGE_SIZE,
    itemCount: 0,
  })

  // 计算属性
  const hasSelectedResources = computed(() => resourceState.selected.length > 0)
  const hasResourceList = computed(() => resourceState.list.length > 0)

  // 重置状态
  const resetState = () => {
    resourceState.userInfo = null
    resourceState.list = []
    resourceState.loading = false
    resourceState.searchKeyword = ''
    resourceState.selected = []
    resourcePagination.page = PAGINATION_CONFIG.DEFAULT_PAGE
    resourcePagination.pageSize = PAGINATION_CONFIG.DEFAULT_PAGE_SIZE
    resourcePagination.itemCount = 0
  }

  // 处理API响应的公共方法
  const handleApiResponse = (
    response: ApiResponse<GetSubscribeUserResponse>,
    isInitialSearch = false
  ) => {
    if (response.data) {
      resourceState.userInfo = response.data
      resourceState.list = response.data.data || []
      resourcePagination.itemCount = response.data.total || 0

      if (isInitialSearch) {
        resourcePagination.page = PAGINATION_CONFIG.DEFAULT_PAGE
        resourcePagination.pageSize = PAGINATION_CONFIG.DEFAULT_PAGE_SIZE

        if (resourceState.list.length > 0) {
          message.success(`找到 ${resourcePagination.itemCount} 个资源`)
          return true
        } else {
          message.warning('该用户暂无分享资源')
          return false
        }
      }
      return true
    } else {
      message.error(response.msg || '获取用户资源失败')
      return false
    }
  }

  // 获取资源列表
  const fetchResourceList = async (isInitialSearch = false) => {
    if (!subscribeUserId.value.trim()) return false

    resourceState.loading = true
    try {
      const response = await getSubscribeUser({
        subscribeUser: subscribeUserId.value.trim(),
        name: resourceState.searchKeyword || undefined,
        currentPage: resourcePagination.page,
        pageSize: resourcePagination.pageSize,
      })
      return handleApiResponse(response, isInitialSearch)
    } catch (error) {
      console.error('获取资源列表失败:', error)
      message.error('获取资源列表失败')
      return false
    } finally {
      resourceState.loading = false
    }
  }

  // 搜索资源
  const handleSearchResource = () => {
    resourcePagination.page = PAGINATION_CONFIG.DEFAULT_PAGE
    fetchResourceList()
  }

  // 重置资源搜索
  const handleResetResourceSearch = () => {
    resourceState.searchKeyword = ''
    resourcePagination.page = PAGINATION_CONFIG.DEFAULT_PAGE
    fetchResourceList()
  }

  // 分页处理
  const handleResourcePageChange = (page: number) => {
    resourcePagination.page = page
    fetchResourceList()
  }

  const handleResourcePageSizeChange = (pageSize: number) => {
    resourcePagination.pageSize = pageSize
    resourcePagination.page = PAGINATION_CONFIG.DEFAULT_PAGE
    fetchResourceList()
  }

  // 获取全部资源
  const fetchAllResources = async () => {
    if (!subscribeUserId.value.trim()) return false

    resourceState.loadingAll = true
    resourceState.selected = []
    try {
      const response = await getSubscribeUserAll({
        subscribeUser: subscribeUserId.value.trim(),
      })
      if (response.data) {
        resourceState.list = response.data.data || []
        resourceState.userInfo = {
          name: response.data.name,
          data: response.data.data || [],
          total: response.data.total,
          currentPage: 1,
          pageSize: response.data.total,
        }
        resourcePagination.itemCount = response.data.total || 0
        resourcePagination.pageSize = response.data.total || 0
        
        message.success(`已加载全部 ${response.data.total} 个资源`)
        return true
      } else {
        message.error(response.msg || '获取全部资源失败')
        return false
      }
    } catch (error) {
      console.error('获取全部资源失败:', error)
      message.error('获取全部资源失败')
      return false
    } finally {
      resourceState.loadingAll = false
    }
  }

  const checkedRowKeys = computed(() => resourceState.selected.map((r) => r.id))

  const handleCheckedRowKeysChange = (keys: Array<string | number>) => {
    resourceState.selected = resourceState.list.filter((item) => keys.includes(item.id))
  }

  return {
    // state
    resourceState,
    resourcePagination,
    // computed
    hasSelectedResources,
    hasResourceList,
    // methods
    fetchResourceList,
    fetchAllResources,
    handleSearchResource,
    handleResetResourceSearch,
    handleResourcePageChange,
    handleResourcePageSizeChange,
    resetState,
    checkedRowKeys,
    handleCheckedRowKeysChange,
  }
}
