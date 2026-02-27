<template>
  <div class="subscribe-mount-container">
    <div class="subscribe-mount-content">
      <!-- 第一步：输入订阅用户ID -->
      <div v-if="currentStep === 1" class="step-content">
        <div class="step-header">
          <n-text strong>第一步：输入订阅号ID</n-text>
          <n-text depth="3">请输入要挂载的天翼云盘订阅号ID</n-text>
        </div>
        <div class="input-section">
          <n-input
            v-model:value="subscribeUserId"
            placeholder="请输入订阅号ID"
            clearable
            size="large"
            @keyup.enter="handleSearchUser"
          >
            <template #prefix>
              <n-icon :size="16"><PersonOutline /></n-icon>
            </template>
          </n-input>
          <n-button
            type="primary"
            size="large"
            :loading="searchLoading"
            :disabled="!isValidSubscribeUserId"
            @click="handleSearchUser"
            style="margin-top: 16px; width: 100%"
          >
            <template #icon
              ><n-icon><SearchOutline /></n-icon
            ></template>
            查询订阅号资源列表
          </n-button>
        </div>
      </div>

      <!-- 第二步：展示资源列表 -->
      <div v-if="currentStep === 2" class="step-content">
        <div class="step-header">
          <n-text strong>第二步：选择要挂载的资源</n-text>
          <div class="user-info-line">
            <n-text depth="3">用户：{{ resourceState.userInfo?.name || subscribeUserId }}</n-text>
            <n-text v-if="hasSelectedResources" type="primary" class="selected-count">
              <n-icon :size="14" color="#2196f3" style="vertical-align: middle; margin-right: 4px">
                <CheckmarkCircleOutline />
              </n-icon>
              已选中 {{ resourceState.selected.length }} 个文件
            </n-text>
          </div>
        </div>

        <!-- 批量操作区域 -->
        <div class="batch-actions">
          <div class="search-section">
            <n-input
              v-model:value="resourceState.searchKeyword"
              placeholder="搜索资源名称"
              clearable
              @keyup.enter="handleSearchResource"
              style="width: 200px"
            >
              <template #prefix
                ><n-icon :size="16"><SearchOutline /></n-icon
              ></template>
            </n-input>
            <n-button type="primary" @click="handleSearchResource" style="margin-left: 8px"
              >搜索</n-button
            >
            <n-button @click="handleResetResourceSearch" style="margin-left: 8px">重置</n-button>
          </div>
          <div class="batch-select-actions">
            <n-button
              type="info"
              :loading="resourceState.loadingAll"
              :disabled="resourceState.loadingAll"
              @click="handleFetchAllResources"
            >
              获取全部资源
            </n-button>
          </div>
        </div>

        <!-- 资源表格 -->
        <div class="table-container">
          <n-spin :show="resourceState.loading">
            <n-data-table
              v-if="hasResourceList"
              :columns="resourceColumns"
              :data="resourceState.list"
              :pagination="false"
              :bordered="false"
              size="small"
              class="resource-table"
              :row-key="(row: ShareResourceInfo) => row.id"
              :checked-row-keys="checkedRowKeys"
              @update:checked-row-keys="handleCheckedRowKeysChange"
            />
            <n-empty v-else description="暂无资源数据" size="large" style="min-height: 200px">
              <template #icon
                ><n-icon size="48" :depth="3"><DocumentTextOutline /></n-icon
              ></template>
            </n-empty>
          </n-spin>
        </div>

        <!-- 分页 -->
        <div v-if="hasResourceList && !isAllLoadedComputed" class="pagination-section">
          <n-pagination
            v-model:page="resourcePagination.page"
            v-model:page-size="resourcePagination.pageSize"
            :item-count="resourcePagination.itemCount"
            :page-sizes="PAGINATION_CONFIG.PAGE_SIZES"
            show-size-picker
            @update:page="handleResourcePageChange"
            @update:page-size="handleResourcePageSizeChange"
          />
        </div>
        <div v-if="hasResourceList && isAllLoadedComputed" class="pagination-section">
          <n-text depth="3">已加载全部 {{ resourcePagination.itemCount }} 个资源</n-text>
        </div>
      </div>
    </div>

    <!-- 操作按钮 -->
    <div class="modal-actions">
      <n-button v-if="currentStep === 2" @click="handleBackToStep1">
        <template #icon
          ><n-icon><ArrowBackOutline /></n-icon
        ></template>
        返回上一步
      </n-button>
      <n-button @click="handleCancel">取消</n-button>
      <n-button
        v-if="currentStep === 2"
        type="primary"
        :disabled="!hasSelectedResources"
        @click="handleConfirm"
      >
        绑定挂载点
      </n-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, h, onMounted } from 'vue'
import {
  NIcon,
  NText,
  NInput,
  NButton,
  NSpin,
  NEmpty,
  NPagination,
  NDataTable,
  useMessage,
  type DataTableColumns,
} from 'naive-ui'
import {
  DocumentTextOutline,
  PersonOutline,
  SearchOutline,
  FolderOutline,
  DocumentOutline,
  CheckmarkCircleOutline,
  ArrowBackOutline,
} from '@vicons/ionicons5'
import type { ShareResourceInfo } from '@/api/storage/advance'
import { formatDateTime } from '@/utils/time'
import { OS_TYPES } from '@/utils/osType'
import { useMountPointBind } from '@/composables/useMountPointBind'
import { useSubscribeResource } from '@/composables/useSubscribeResource'

// 常量定义
const PAGINATION_CONFIG = {
  PAGE_SIZES: [20, 30, 50, 100] as number[],
}

// Emits
interface Emits {
  (e: 'confirm', data: { success: boolean }): void
  (e: 'cancel'): void
}
const emit = defineEmits<Emits>()

// Hooks
const message = useMessage()
const mountPointBind = useMountPointBind()

// 步骤控制
const currentStep = ref(1)

// 第一步：用户搜索状态
const subscribeUserId = ref('')
const searchLoading = ref(false)
const isValidSubscribeUserId = computed(() => subscribeUserId.value.trim().length > 0)

// 第二步：资源管理
const {
  resourceState,
  resourcePagination,
  hasSelectedResources,
  hasResourceList,
  fetchResourceList,
  fetchAllResources,
  handleSearchResource,
  handleResetResourceSearch,
  handleResourcePageChange,
  handleResourcePageSizeChange,
  resetState: resetResourceState,
  checkedRowKeys,
  handleCheckedRowKeysChange,
} = useSubscribeResource(subscribeUserId, message)

// 判断是否已加载全部资源
const isAllLoaded = ref(false)

// 判断是否已加载全部资源
const isAllLoadedComputed = computed(() => {
  return isAllLoaded.value || (resourceState.list.length > 0 && resourcePagination.itemCount > 0 && 
         resourceState.list.length >= resourcePagination.itemCount)
})

// 获取全部资源
const handleFetchAllResources = async () => {
  const success = await fetchAllResources()
  if (success) {
    isAllLoaded.value = true
  }
}

// 表格列定义
const resourceColumns: DataTableColumns<ShareResourceInfo> = [
  {
    type: 'selection',
    align: 'center',
  },
  {
    title: '序号',
    key: 'index',
    width: 80,
    align: 'center',
    ellipsis: { tooltip: true },
    render: (_, index) => isAllLoadedComputed ? index + 1 : (resourcePagination.page - 1) * resourcePagination.pageSize + index + 1,
  },
  {
    title: '资源名称',
    key: 'name',
    width: 300,
    align: 'left',
    ellipsis: { tooltip: true },
    render: (row) =>
      h('div', { style: 'display: flex; align-items: center; gap: 8px; justify-content: left' }, [
        h(
          NIcon,
          { size: 20, color: row.isFolder ? '#ff9800' : '#2196f3' },
          { default: () => (row.isFolder ? h(FolderOutline) : h(DocumentOutline)) }
        ),
        h('span', { style: 'word-break: break-all;' }, row.name),
      ]),
  },
  {
    title: '类型',
    key: 'type',
    width: 100,
    align: 'center',
    ellipsis: { tooltip: true },
    render: (row) => (row.isFolder ? '文件夹' : '单文件'),
  },
  {
    title: '分享时间',
    key: 'shareTime',
    width: 180,
    align: 'center',
    ellipsis: { tooltip: true },
    render: (row) => formatDateTime(row.shareTime),
  },
]

// 搜索用户资源
const handleSearchUser = async () => {
  if (!isValidSubscribeUserId.value) {
    message.warning('请输入订阅用户ID')
    return
  }
  searchLoading.value = true
  const success = await fetchResourceList(true)
  if (success) {
    currentStep.value = 2
  }
  searchLoading.value = false
}

// 返回上一步
const handleBackToStep1 = () => {
  currentStep.value = 1
  resourceState.selected = []
}

// 取消
const handleCancel = () => {
  emit('cancel')
}

// 确认挂载
const handleConfirm = () => {
  if (!hasSelectedResources.value) {
    message.warning('请选择要挂载的资源')
    return
  }
  const itemsToMount = resourceState.selected.map((resource) => ({
    name: resource.name,
    osType: OS_TYPES.SUBSCRIBE_SHARE_FOLDER,
    subscribeUser: subscribeUserId.value.trim(),
    shareCode: resource.accessCode,
    fileId: resource.id,
  }))

  mountPointBind.show(itemsToMount).then((payload) => {
    if (payload && payload.length > 0) {
      handleMountBindSuccess()
    }
  })
}

// 挂载成功回调
const handleMountBindSuccess = () => {
  emit('confirm', { success: true })
}

// 重置所有状态
const resetAllState = () => {
  currentStep.value = 1
  subscribeUserId.value = ''
  searchLoading.value = false
  resetResourceState()
}

// 组件挂载时重置状态
onMounted(() => {
  resetAllState()
})
</script>

<style scoped>
.subscribe-mount-container {
  min-width: 800px;
  width: 100%;
}

.subscribe-mount-content {
  padding: 16px 0;
}

.step-content {
  min-height: 300px;
}

.step-header {
  margin-bottom: 24px;
  text-align: center;
}

.step-header .n-text:first-child {
  display: block;
  margin-bottom: 8px;
  font-size: 18px;
}

.user-info-line {
  display: flex;
  align-items: baseline;
  justify-content: center;
  gap: 16px;
  flex-wrap: wrap;
  line-height: 1;
}

.selected-count {
  display: flex;
  align-items: center;
  font-size: 14px;
  line-height: 1;
}

.input-section {
  max-width: 400px;
  margin: 0 auto;
}

.batch-actions {
  margin-bottom: 16px;
  padding: 12px 16px;
  background: var(--n-card-color);
  border: 1px solid var(--n-border-color);
  border-radius: 6px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
}

.batch-select-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.search-section {
  display: flex;
  align-items: center;
  gap: 8px;
}

.table-container {
  max-height: 400px;
  overflow-y: auto;
  border: 1px solid var(--n-border-color);
  border-radius: 6px;
}

.resource-table {
  min-height: 200px;
}

:deep(.n-data-table-thead) {
  position: sticky;
  top: 0;
  z-index: 1;
  background: var(--n-th-color);
}

:deep(.n-data-table-th) {
  background: var(--n-th-color);
  font-weight: 600;
}

:deep(.n-data-table-td) {
  padding: 12px 8px;
}

.pagination-section {
  display: flex;
  justify-content: center;
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid var(--n-border-color);
}

.modal-actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
  padding-top: 16px;
  border-top: 1px solid var(--n-border-color);
}

@media (width <= 768px) {
  .search-section {
    flex-direction: column;
    gap: 8px;
  }

  .search-section .n-input {
    width: 100%;
  }

  .modal-actions {
    flex-direction: column;
  }
}
</style>
