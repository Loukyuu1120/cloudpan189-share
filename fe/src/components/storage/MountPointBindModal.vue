<template>
  <div class="mount-bind-content">
    <!-- 批量操作区域 -->
    <div class="batch-actions">
      <n-card size="small" title="批量设置">
        <div class="batch-grid">
          <div class="batch-item">
            <n-text class="batch-label">路径前缀</n-text>
            <n-input
              v-model:value="storageSetting.pathPrefix"
              placeholder="必须以 / 开头"
              style="flex-grow: 1"
            />
            <n-button
              type="primary"
              size="small"
              :disabled="!hasValidPathPrefix"
              @click="handleBatchApplyPathPrefix"
            >
              应用
            </n-button>
          </div>

          <div class="batch-item" v-if="!allTokenSwitchDisabled">
            <n-text class="batch-label">云盘令牌</n-text>
            <n-select
              v-model:value="storageSetting.selectedToken"
              :options="cloudTokenOptions"
              placeholder="选择令牌"
              clearable
              style="flex-grow: 1"
            />
            <n-button
              type="primary"
              size="small"
              :disabled="storageSetting.selectedToken === undefined"
              @click="handleBatchApplyToken"
            >
              应用
            </n-button>
          </div>

          <div class="batch-item">
            <n-text class="batch-label">自动刷新</n-text>
            <n-switch v-model:value="storageSetting.enableAutoRefresh" />
            <n-button
              v-if="storageSetting.enableAutoRefresh"
              type="primary"
              size="small"
              @click="handleEditAutoRefreshConfig"
            >
              编辑
            </n-button>
          </div>
        </div>
        <template #footer>
          <n-text depth="3">
            提示：批量设置将应用到下方所有可编辑的行。路径前缀会与识别出的名称组合成完整挂载路径。
          </n-text>
        </template>
      </n-card>
    </div>

    <div class="table-container">
      <n-data-table
        :columns="columns"
        :data="tableData"
        :pagination="false"
        :bordered="false"
        size="small"
        class="mount-table"
      />
    </div>

    <!-- 自动刷新配置弹窗 -->
    <n-modal v-model:show="showAutoRefreshModal" preset="dialog" title="自动刷新配置">
      <div class="auto-refresh-config">
        <n-form
          ref="autoRefreshFormRef"
          :model="autoRefreshForm"
          :rules="autoRefreshRules"
          label-placement="left"
          label-width="120px"
        >
          <n-form-item label="刷新间隔(分钟)" path="refreshInterval">
            <n-input-number
              v-model:value="autoRefreshForm.refreshInterval"
              :min="30"
              :max="1440"
              placeholder="30-1440分钟"
              style="width: 100%"
            />
          </n-form-item>

          <n-form-item label="持续天数" path="autoRefreshDays">
            <n-input-number
              v-model:value="autoRefreshForm.autoRefreshDays"
              :min="1"
              :max="365"
              placeholder="1-365天"
              style="width: 100%"
            />
          </n-form-item>

          <n-form-item label="深度刷新" path="enableDeepRefresh">
            <n-switch v-model:value="autoRefreshForm.enableDeepRefresh" />
          </n-form-item>
        </n-form>
      </div>

      <template #action>
        <n-button @click="showAutoRefreshModal = false">取消</n-button>
        <n-button type="primary" @click="handleAutoRefreshConfirm"> 确认 </n-button>
      </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { reactive, computed, h, onMounted, ref } from 'vue'
import { storeToRefs } from 'pinia'
import {
  NButton,
  NDataTable,
  NInput,
  NSelect,
  NText,
  useMessage,
  type DataTableColumns,
  NCard,
  NSwitch,
  NModal,
  NForm,
  NFormItem,
  NInputNumber,
} from 'naive-ui'
import { batchAddStorage, type AddStorageRequest, type BatchAddStorageResponse } from '@/api/storage'
import type { ApiResponse } from '@/utils/api'
import { getCloudTokenList } from '@/api/cloudtoken'
import { getOsTypeDisplayName, getOsTypeColor } from '@/utils/osType'
import { useSharedStore } from '@/stores/modules/shared'
import type { FormRules } from 'naive-ui'

export interface MountItem {
  name: string
  osType: string
  subscribeUser?: string
  shareCode?: string
  shareAccessCode?: string
  cloudToken?: number
  disableSwitchCloudToken?: boolean
  fileId?: string
  familyId?: string
}

interface Props {
  items: MountItem[]
  defaultCloudToken?: number
}

interface Emits {
  (e: 'confirm', payload: { id: number; path: string }[]): void
  (e: 'cancel'): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

// 消息提示
const message = useMessage()

// 状态管理
const state = reactive({
  submitLoading: false,
  cloudTokens: [] as Models.CloudToken[],
})

// 共享存储
const sharedStore = useSharedStore()
const { storageSetting } = storeToRefs(sharedStore)

// 表格数据
interface TableRow extends MountItem {
  id: string
  localPath: string
  selectedCloudToken?: number
}

const tableData = reactive<TableRow[]>([])

// 自动刷新配置相关
const showAutoRefreshModal = ref(false)
const autoRefreshFormRef = ref<InstanceType<typeof NForm>>()

const autoRefreshForm = reactive({
  refreshInterval: 60,
  autoRefreshDays: 7,
  enableDeepRefresh: false,
})

const autoRefreshRules: FormRules = {
  refreshInterval: [
    {
      type: 'number' as const,
      min: 30,
      max: 1440,
      message: '刷新间隔必须在30-1440分钟之间',
      trigger: 'blur' as const,
    },
  ],
  autoRefreshDays: [
    {
      type: 'number' as const,
      min: 1,
      max: 365,
      message: '持续天数必须在1-365天之间',
      trigger: 'blur' as const,
    },
  ],
}

// 计算属性
const cloudTokenOptions = computed(() => [
  { label: '不绑定', value: 0 },
  ...state.cloudTokens.map((token) => ({
    label: token.name,
    value: token.id,
  })),
])

const hasValidPathPrefix = computed(
  () => storageSetting.value.pathPrefix && storageSetting.value.pathPrefix.startsWith('/')
)

const hasInvalidRows = computed(() => tableData.some((row) => !row.localPath.trim()))

// 检查是否所有项目都禁用令牌切换
const allTokenSwitchDisabled = computed(
  () => tableData.length > 0 && tableData.every((row) => row.disableSwitchCloudToken)
)

// 表格列定义
const columns: DataTableColumns<TableRow> = [
  {
    title: '序号',
    key: 'index',
    width: 80,
    render: (_, index) => index + 1,
    align: 'center',
    ellipsis: { tooltip: true },
  },
  {
    title: '识别出的名称',
    key: 'name',
    width: 200,
    align: 'center',
    ellipsis: {
      tooltip: true,
    },
  },
  {
    title: '挂载类型',
    key: 'osType',
    width: 150,
    align: 'center',
    ellipsis: { tooltip: true },
    render: (row) => {
      const displayName = getOsTypeDisplayName(row.osType)
      const colorInfo = getOsTypeColor(row.osType)
      return h('span', { style: { color: colorInfo.textColor } }, displayName)
    },
  },
  {
    title: '挂载路径',
    key: 'localPath',
    width: 250,
    align: 'left',
    titleAlign: 'center',
    render: (row, index) => {
      return h(NInput, {
        value: row.localPath,
        placeholder: '请输入挂载路径',
        onUpdateValue: (value: string) => {
          tableData[index].localPath = value
        },
      })
    },
  },
  {
    title: '绑定令牌',
    key: 'selectedCloudToken',
    width: 200,
    align: 'center',
    render: (row, index) => {
      return h(NSelect, {
        value: row.selectedCloudToken,
        options: cloudTokenOptions.value,
        placeholder: '选择云盘令牌',
        clearable: true,
        disabled: row.disableSwitchCloudToken,
        onUpdateValue: (value: number | undefined) => {
          tableData[index].selectedCloudToken = value
        },
      })
    },
  },
]

// 初始化表格数据
const initTableData = () => {
  const newItems = props.items.map(
    (item, index) =>
      ({
        ...item,
        id: `item_${index}`,
        localPath: `${storageSetting.value.pathPrefix || '/'}${item.name}`,
        selectedCloudToken: item.disableSwitchCloudToken
          ? item.cloudToken
          : storageSetting.value.selectedToken,
      }) as TableRow
  )
  tableData.length = 0
  tableData.push(...newItems)
}

// 获取云盘令牌列表
const fetchCloudTokens = () => {
  return getCloudTokenList({ noPaginate: true })
    .then((res) => {
      if (res.data) {
        state.cloudTokens = res.data.data
      }
    })
    .catch((error) => {
      message.error(error?.message || '获取云盘令牌列表失败')
    })
}

// 批量应用令牌
const handleBatchApplyToken = () => {
  if (allTokenSwitchDisabled.value) {
    message.warning('所有项目都禁止修改令牌')
    return
  }

  // 将选中的令牌应用到所有未禁用的行
  let appliedCount = 0
  let skippedCount = 0

  tableData.forEach((row) => {
    if (!row.disableSwitchCloudToken) {
      row.selectedCloudToken = storageSetting.value.selectedToken
      appliedCount++
    } else {
      skippedCount++
    }
  })

  if (skippedCount > 0) {
    message.success(`已批量应用令牌设置到 ${appliedCount} 个项目，跳过 ${skippedCount} 个禁用项目`)
  } else {
    message.success('已批量应用令牌设置')
  }
}

// 批量应用路径前缀
const handleBatchApplyPathPrefix = () => {
  if (!hasValidPathPrefix.value) {
    message.warning('路径前缀必须以 / 开头')
    return
  }

  // 确保前缀以 / 结尾（如果不是单独的 /）
  let prefix = storageSetting.value.pathPrefix
  if (prefix !== '/' && !prefix.endsWith('/')) {
    prefix += '/'
  }

  // 将路径前缀应用到所有行
  tableData.forEach((row) => {
    row.localPath = `${prefix}${row.name}`
  })

  message.success('已批量应用路径前缀设置')
}

// 取消
const handleCancel = () => {
  emit('cancel')
}

// 编辑自动刷新配置
const handleEditAutoRefreshConfig = () => {
  // 从 store 同步当前配置到表单
  autoRefreshForm.refreshInterval = storageSetting.value.refreshInterval || 60
  autoRefreshForm.autoRefreshDays = storageSetting.value.autoRefreshDays || 7
  autoRefreshForm.enableDeepRefresh = storageSetting.value.enableDeepRefresh || false

  showAutoRefreshModal.value = true
}

// 确认自动刷新配置
const handleAutoRefreshConfirm = () => {
  autoRefreshFormRef.value?.validate((errors: unknown) => {
    if (errors) {
      message.error('请检查表单输入')
      return
    }

    // 保存配置到共享存储
    storageSetting.value.autoRefreshDays = autoRefreshForm.autoRefreshDays
    storageSetting.value.refreshInterval = autoRefreshForm.refreshInterval
    storageSetting.value.enableDeepRefresh = autoRefreshForm.enableDeepRefresh

    showAutoRefreshModal.value = false
    message.success('自动刷新配置已保存')
  })
}

// 构建请求数据
const buildRequests = (): AddStorageRequest[] => {
  return tableData.map((row) => ({
    localPath: row.localPath.trim(),
    osType: row.osType as AddStorageRequest['osType'],
    cloudToken: row.selectedCloudToken,
    subscribeUser: row.subscribeUser,
    shareCode: row.shareCode,
    shareAccessCode: row.shareAccessCode,
    fileId: row.fileId,
    familyId: row.familyId,
    enableAutoRefresh: storageSetting.value.enableAutoRefresh,
    autoRefreshDays: storageSetting.value.enableAutoRefresh
      ? storageSetting.value.autoRefreshDays
      : undefined,
    refreshInterval: storageSetting.value.enableAutoRefresh
      ? storageSetting.value.refreshInterval
      : undefined,
    enableDeepRefresh: storageSetting.value.enableAutoRefresh
      ? storageSetting.value.enableDeepRefresh
      : undefined,
  }))
}

// 处理批量挂载结果
const handleMountResults = (response: ApiResponse<BatchAddStorageResponse>) => {
  if (response.code !== 200) {
    message.error(response.msg || '批量挂载失败')
    return
  }

  const { successCount, failCount, results } = response.data

  const successItems = results.filter((r) => r.success).map((r) => ({
    id: r.id,
    path: r.localPath,
  }))

  if (successCount > 0) {
    message.success(`成功挂载 ${successCount} 个存储点${failCount > 0 ? `，失败 ${failCount} 个` : ''}`)
    emit('confirm', successItems)
  } else {
    const firstError = results.find((r) => !r.success)?.error
    message.error(`挂载全部失败${firstError ? `: ${firstError}` : ''}`)
  }
}

// 确认挂载
const handleConfirm = () => {
  // 验证数据
  if (hasInvalidRows.value) {
    message.warning('请填写所有挂载路径')
    return
  }

  state.submitLoading = true

  // 构建请求数据
  const requests = buildRequests()

  // 批量添加存储挂载（使用后台任务方式）
  return batchAddStorage({ items: requests })
    .then((response) => {
      handleMountResults(response)
    })
    .catch((error) => {
      console.error('批量挂载失败:', error)
      message.error('批量挂载失败')
    })
    .finally(() => {
      state.submitLoading = false
    })
}

// 组件挂载时初始化数据
onMounted(() => {
  // 如果有默认令牌，并且用户没有自定义设置，则使用默认令牌
  if (props.defaultCloudToken && storageSetting.value.selectedToken === 0) {
    storageSetting.value.selectedToken = props.defaultCloudToken
  }
  initTableData()
  fetchCloudTokens()
})

defineExpose({
  handleConfirm,
  handleCancel,
  state,
})
</script>

<style scoped>
.mount-bind-content {
  padding: 16px 0;
}

.batch-actions {
  margin-bottom: 16px;
}

.batch-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 16px;
}

.batch-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.batch-label {
  font-weight: 600;
  white-space: nowrap;
}

.table-container {
  max-height: 400px;
  overflow-y: auto;
  border: 1px solid var(--n-border-color);
  border-radius: 6px;
}

.mount-table {
  min-height: 200px;
}

/* 固定表格标题 */
:deep(.n-data-table-thead) {
  position: sticky;
  top: 0;
  z-index: 1;
  background: var(--n-th-color);
}

.modal-actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}

/* 表格样式优化 */
:deep(.n-data-table-th) {
  background: var(--n-th-color);
  font-weight: 600;
}

:deep(.n-data-table-td) {
  padding: 12px 8px;
}

/* 响应式设计 */
@media (width <= 768px) {
  .modal-actions {
    flex-direction: column;
  }
}
</style>
