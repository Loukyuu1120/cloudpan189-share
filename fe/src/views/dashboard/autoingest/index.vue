<template>
  <div class="autoingest-page">
    <!-- 顶部 Tabs -->
    <n-tabs type="line" v-model:value="activeTab">
      <n-tab name="plans">计划管理</n-tab>
      <n-tab name="logs">运行日志</n-tab>
    </n-tabs>

    <!-- 头部区域（Plans） -->
    <div v-if="activeTab === 'plans'" class="header">
      <div class="header-left">
        <n-input
          v-model:value="planQuery.name"
          placeholder="按名称搜索计划"
          clearable
          style="width: 240px; margin-right: 12px"
          @keyup.enter="handlePlanSearch"
        >
          <template #prefix>
            <n-icon :size="16" :depth="3">
              <SearchOutline />
            </n-icon>
          </template>
        </n-input>
        <n-button type="primary" @click="handlePlanSearch" style="margin-right: 8px">
          <template #icon>
            <n-icon>
              <SearchOutline />
            </n-icon>
          </template>
          搜索
        </n-button>
        <n-button @click="handlePlanReset">
          <template #icon>
            <n-icon>
              <RefreshOutline />
            </n-icon>
          </template>
          重置
        </n-button>
      </div>
      <div class="header-right">
        <n-button type="primary" @click="showCreateModal = true">
          <template #icon>
            <n-icon>
              <AddOutline />
            </n-icon>
          </template>
          新建入库计划
        </n-button>
      </div>
    </div>

    <!-- 头部区域（Logs） -->
    <div v-else class="header">
      <div class="header-left">
        <n-select
          v-model:value="logQuery.planId"
          :options="planOptions"
          placeholder="按计划筛选"
          clearable
          style="width: 220px; margin-right: 8px"
        />
        <n-select
          v-model:value="logQuery.level"
          :options="logLevelOptions"
          placeholder="日志级别"
          clearable
          style="width: 160px; margin-right: 8px"
        />
        <n-button type="primary" @click="handleLogFilter" style="margin-right: 8px">
          <template #icon>
            <n-icon>
              <SearchOutline />
            </n-icon>
          </template>
          筛选
        </n-button>
        <n-button @click="handleLogReset">
          <template #icon>
            <n-icon>
              <RefreshOutline />
            </n-icon>
          </template>
          重置
        </n-button>
        <n-button type="warning" @click="handleRetryAllFailed" style="margin-left: 8px">
          批量重试失败
        </n-button>
        <n-button type="error" @click="handleDeleteErrorLogs" style="margin-left: 8px">
          删除错误日志
        </n-button>
      </div>
      <div class="header-right">
        <n-text depth="3">最近刷新：{{ refreshTime.format('YYYY-MM-DD HH:mm:ss') }}</n-text>
      </div>
    </div>

    <!-- 计划管理表格 -->
    <n-data-table
      v-if="activeTab === 'plans'"
      :columns="planColumns"
      :data="planTable"
      :loading="planLoading"
      :pagination="planPagination"
      class="autoingest-table"
      remote
    />

    <!-- 运行日志表格 -->
    <n-data-table
      v-else
      :columns="logColumns"
      :data="logTable"
      :loading="logLoading"
      :pagination="logPagination"
      class="autoingest-table"
      remote
    />

    <!-- 新建计划弹窗（组件化） -->
    <CreatePlanModal
      v-model:show="showCreateModal"
      :cloud-token-options="cloudTokenOptions"
      @created="handlePlanCreated"
    />

    <!-- 修改计划弹窗（组件化） -->
    <EditPlanModal
      v-model:show="showEditModal"
      :plan="editingPlan"
      :cloud-token-options="cloudTokenOptions"
      @saved="fetchPlanList"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, h, onMounted } from 'vue'
import {
  NDataTable,
  NButton,
  NIcon,
  NInput,
  NText,
  NSelect,
  NSpace,
  NPopconfirm,
  NTabs,
  NTab,
  NTag,
  useMessage,
  type DataTableColumns,
  type PaginationProps,
} from 'naive-ui'
import {
  SearchOutline,
  RefreshOutline,
  AddOutline,
  TrashOutline,
  CheckmarkCircleOutline,
  CloseCircleOutline,
  CreateOutline,
} from '@vicons/ionicons5'
import {
  getAutoIngestPlanList,
  enableAutoIngestPlan,
  disableAutoIngestPlan,
  refreshAutoIngestPlan,
  deleteAutoIngestPlan,
  getAutoIngestLogList,
  retryFailedAutoIngest,
  deleteErrorLogs,
  type PlanLogResult,
} from '@/api/autoingest'
import { getCloudTokenList } from '@/api/cloudtoken'
import dayjs from 'dayjs'
import { AUTO_INGEST_SOURCE_TYPE_OPTIONS } from '@/constants/autoIngest'
import CreatePlanModal from '@/components/autoingest/CreatePlanModal.vue'
import EditPlanModal from '@/components/autoingest/EditPlanModal.vue'
import { type ApiResponse } from '@/utils/api'

const message = useMessage()

// Tabs
const activeTab = ref<'plans' | 'logs'>('plans')

// Refresh Time
const refreshTime = ref(dayjs())

// Cloud Token options
const cloudTokenOptions = ref<{ label: string; value: number }[]>([])
const loadCloudTokens = () => {
  getCloudTokenList({ noPaginate: true })
    .then((res: ApiResponse<Models.PaginationResponse<Models.CloudToken>>) => {
      if (res.code === 200 && res.data) {
        cloudTokenOptions.value = res.data.data.map((t) => ({
          label: t.name || `令牌${t.id}`,
          value: t.id,
        }))
      }
    })
    .catch((err: unknown) => {
      // 忽略错误，仅防止类型告警
      console.error('获取令牌列表失败:', err)
    })
}

// -------- Plans --------
const planLoading = ref(false)
const planTable = ref<Models.AutoIngestPlan[]>([])
const planQuery = reactive({
  name: '' as string | undefined,
})

// 分页（对齐用户组管理）
const planPagination = reactive<PaginationProps>({
  page: 1,
  pageSize: 10,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
  prefix: ({ itemCount }) => `共 ${itemCount} 条`,
  onChange: (page: number) => {
    planPagination.page = page
    fetchPlanList()
  },
  onUpdatePageSize: (ps: number) => {
    planPagination.pageSize = ps
    planPagination.page = 1
    fetchPlanList()
  },
})

const handlePlanSearch = () => {
  planPagination.page = 1
  fetchPlanList()
}

const handlePlanReset = () => {
  planQuery.name = ''
  planPagination.page = 1
  fetchPlanList()
}

const planColumns: DataTableColumns<Models.AutoIngestPlan> = [
  {
    title: '计划名称',
    key: 'name',
    width: 120,
    align: 'center',
    ellipsis: { tooltip: true },
  },
  {
    title: '来源类型',
    key: 'sourceType',
    width: 80,
    align: 'center',
    render: (row) => {
      const opt = AUTO_INGEST_SOURCE_TYPE_OPTIONS.find((o) => o.value === row.sourceType)
      return h(
        NTag,
        { type: 'primary', size: 'small', bordered: true },
        { default: () => opt?.label || '-' }
      )
    },
  },
  {
    title: '父目录',
    key: 'parentPath',
    width: 120,
    align: 'center',
    ellipsis: { tooltip: true },
    render: (row) =>
      h(
        'div',
        {
          style: 'max-width:240px;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;',
          title: row.parentPath || '-',
        },
        row.parentPath || '-'
      ),
  },
  {
    title: '间隔(分钟)',
    key: 'autoIngestInterval',
    width: 90,
    align: 'center',
  },
  {
    title: '启用',
    key: 'enabled',
    width: 90,
    align: 'center',
    render: (row) => {
      const type = row.enabled ? 'success' : 'default'
      const label = row.enabled ? '启用' : '停用'
      return h(NTag, { type, size: 'small', bordered: true }, { default: () => label })
    },
  },
  {
    title: '统计',
    key: 'stats',
    width: 100,
    align: 'center',
    render: (row) =>
      h(
        'div',
        {
          class: 'stats-cell',
          style: 'display:flex;justify-content:center;gap:6px;align-items:center;',
        },
        [
          h(
            NTag,
            { type: 'success', size: 'small', bordered: true },
            { default: () => String(row.addCount || 0) }
          ),
          h('span', null, '/'),
          h(
            NTag,
            { type: 'error', size: 'small', bordered: true },
            { default: () => String(row.failedCount || 0) }
          ),
        ]
      ),
  },
  {
    title: '时间',
    key: 'time',
    width: 220,
    align: 'center',
    render: (row) =>
      h('div', { class: 'time-cell' }, [
        h('div', { class: 'time-line' }, [
          h('span', { class: 'time-label' }, '创建: '),
          h('span', { class: 'time-value' }, formatDT(row.createdAt)),
        ]),
        h('div', { class: 'time-line' }, [
          h('span', { class: 'time-label' }, '更新: '),
          h('span', { class: 'time-value' }, formatDT(row.updatedAt)),
        ]),
      ]),
  },
  {
    title: '操作',
    key: 'actions',
    minWidth: 220,
    align: 'center',
    render: (row) =>
      h(
        NSpace,
        { size: 'small', justify: 'center', align: 'center' },
        {
          default: () => [
            row.enabled
              ? [
                  h(
                    NButton,
                    { size: 'tiny', type: 'info', secondary: true, onClick: () => onRefresh(row) },
                    {
                      icon: () => h(NIcon, { size: 12 }, { default: () => h(RefreshOutline) }),
                      default: () => '扫描',
                    }
                  ),
                  // 新增“修改”按钮（位于扫描右边）
                  h(
                    NButton,
                    { size: 'tiny', type: 'primary', secondary: true, onClick: () => onEdit(row) },
                    {
                      icon: () => h(NIcon, { size: 12 }, { default: () => h(CreateOutline) }),
                      default: () => '修改',
                    }
                  ),
                  h(
                    NButton,
                    {
                      size: 'tiny',
                      type: 'warning',
                      secondary: true,
                      onClick: () => onDisable(row),
                    },
                    {
                      icon: () => h(NIcon, { size: 12 }, { default: () => h(CloseCircleOutline) }),
                      default: () => '停用',
                    }
                  ),
                ]
              : h(
                  NButton,
                  { size: 'tiny', type: 'success', secondary: true, onClick: () => onEnable(row) },
                  {
                    icon: () =>
                      h(NIcon, { size: 12 }, { default: () => h(CheckmarkCircleOutline) }),
                    default: () => '启用',
                  }
                ),
            h(
              NPopconfirm,
              {
                onPositiveClick: () => onDelete(row),
                negativeText: '取消',
                positiveText: '确认删除',
              },
              {
                trigger: () =>
                  h(
                    NButton,
                    { size: 'tiny', type: 'error', secondary: true },
                    {
                      icon: () => h(NIcon, { size: 12 }, { default: () => h(TrashOutline) }),
                      default: () => '删除',
                    }
                  ),
                default: () => `确定要删除计划 "${row.name || '#' + row.id}" 吗？此操作不可撤销。`,
              }
            ),
          ],
        }
      ),
  },
]

const fetchPlanList = () => {
  planLoading.value = true
  getAutoIngestPlanList({
    currentPage: planPagination.page || 1,
    pageSize: planPagination.pageSize || 10,
    name: planQuery.name || undefined,
  })
    .then((res: ApiResponse<Models.PaginationResponse<Models.AutoIngestPlan>>) => {
      if (res.code === 200 && res.data) {
        planTable.value = res.data.data
        planPagination.itemCount = res.data.total
        // 更新 planOptions 供日志筛选使用
        planOptions.value = [
          { label: '全部计划', value: undefined },
          ...res.data.data.map((p: Models.AutoIngestPlan) => ({
            label: `${p.name || '#' + p.id}`,
            value: p.id,
          })),
        ]
      }
    })
    .catch((err: unknown) => {
      console.error('获取计划列表失败:', err)
    })
    .finally(() => {
      planLoading.value = false
      refreshTime.value = dayjs()
    })
}

const onEnable = (row: Models.AutoIngestPlan) => {
  enableAutoIngestPlan({ id: row.id })
    .then((res: ApiResponse) => {
      if (res.code === 200) {
        message.success('已启用')
        fetchPlanList()
      }
    })
    .catch((err: unknown) => {
      console.error('启用失败', err)
    })
}

const onDisable = (row: Models.AutoIngestPlan) => {
  disableAutoIngestPlan({ id: row.id })
    .then((res: ApiResponse) => {
      if (res.code === 200) {
        message.success('已停用')
        fetchPlanList()
      }
    })
    .catch((err: unknown) => {
      console.error('停用失败', err)
    })
}

const onRefresh = (row: Models.AutoIngestPlan) => {
  refreshAutoIngestPlan({ planId: row.id })
    .then(() => {
      message.success('已下发扫描任务')
    })
    .catch((err: unknown) => {
      console.error('扫描下发失败', err)
    })
}

const onDelete = (row: Models.AutoIngestPlan) => {
  deleteAutoIngestPlan({ id: row.id })
    .then(() => {
      message.success('删除成功')
      fetchPlanList()
    })
    .catch((err: unknown) => {
      console.error('删除失败', err)
    })
}

/** 由 CreatePlanModal 创建成功后刷新列表 */
const handlePlanCreated = () => {
  fetchPlanList()
}

const showCreateModal = ref(false)

// -------- 修改计划（弹窗：抽离为组件） --------
const showEditModal = ref(false)
const editingPlan = ref<Models.AutoIngestPlan | null>(null)
const onEdit = (row: Models.AutoIngestPlan) => {
  editingPlan.value = row
  showEditModal.value = true
}

// -------- Logs --------
const logLoading = ref(false)
const logTable = ref<PlanLogResult[]>([])
const logQuery = reactive<{
  planId?: number
  level?: 'info' | 'warn' | 'error'
}>({
  planId: undefined,
  level: undefined,
})
const planOptions = ref<{ label: string; value: number | undefined }[]>([
  { label: '全部计划', value: undefined },
])
const logLevelOptions = [
  { label: '全部级别', value: undefined },
  { label: 'info', value: 'info' },
  { label: 'warn', value: 'warn' },
  { label: 'error', value: 'error' },
]

// 分页（Logs 同步风格）
const logPagination = reactive<PaginationProps>({
  page: 1,
  pageSize: 10,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
  prefix: ({ itemCount }) => `共 ${itemCount} 条`,
  onChange: (page: number) => {
    logPagination.page = page
    fetchLogList()
  },
  onUpdatePageSize: (ps: number) => {
    logPagination.pageSize = ps
    logPagination.page = 1
    fetchLogList()
  },
})

const handleLogFilter = () => {
  logPagination.page = 1
  fetchLogList()
}
const handleLogReset = () => {
  logQuery.planId = undefined
  logQuery.level = undefined
  logPagination.page = 1
  fetchLogList()
}

// 批量重试所有失败任务
const handleRetryAllFailed = () => {
  const planId = logQuery.planId
  retryFailedAutoIngest({ planId })
    .then((res) => {
      if (res.code === 200 || res.code === 0) {
        message.success('已下发重试任务')
        fetchLogList()
      } else {
        message.error(res.msg || '重试失败')
      }
    })
    .catch((err: unknown) => {
      console.error('重试失败', err)
      message.error('重试失败')
    })
}

// 删除错误日志
const handleDeleteErrorLogs = () => {
  const planId = logQuery.planId
  deleteErrorLogs({ planId })
    .then((res) => {
      if (res.code === 200 || res.code === 0) {
        message.success(`已删除 ${res.data} 条错误日志`)
        fetchLogList()
      } else {
        message.error(res.msg || '删除失败')
      }
    })
    .catch((err: unknown) => {
      console.error('删除失败', err)
      message.error('删除失败')
    })
}

const logColumns: DataTableColumns<Models.AutoIngestLog> = [
  { title: 'ID', key: 'id', width: 90, align: 'center' },
  { title: '计划名称', key: 'planName', width: 100, align: 'center', ellipsis: { tooltip: true } },
  {
    title: '级别',
    key: 'level',
    width: 100,
    align: 'center',
    render: (row) => {
      const type = row.level === 'error' ? 'error' : row.level === 'warn' ? 'warning' : 'success'
      return h(NTag, { type, size: 'small', bordered: true }, { default: () => row.level })
    },
  },
  {
    title: '内容',
    key: 'content',
    minWidth: 360,
    align: 'center',
    render: (row) => h('div', { class: 'log-content' }, row.content),
  },
  {
    title: '时间',
    key: 'createdAt',
    width: 180,
    align: 'center',
    render: (row) => formatDT(row.createdAt),
  },
  {
    title: '操作',
    key: 'actions',
    width: 100,
    align: 'center',
    render: (row) => {
      if (row.level === 'error') {
        return h(
          NButton,
          {
            size: 'small',
            type: 'warning',
            onClick: () => onRetryFailed(row.planId),
          },
          { default: () => '重试' }
        )
      }
      return null
    },
  },
]

// 重试失败任务
const onRetryFailed = (planId?: number) => {
  retryFailedAutoIngest({ planId })
    .then((res) => {
      if (res.code === 200 || res.code === 0) {
        message.success('已下发重试任务')
        fetchLogList()
      } else {
        message.error(res.msg || '重试失败')
      }
    })
    .catch((err: unknown) => {
      console.error('重试失败', err)
      message.error('重试失败')
    })
}

const fetchLogList = () => {
  logLoading.value = true
  getAutoIngestLogList({
    currentPage: logPagination.page || 1,
    pageSize: logPagination.pageSize || 10,
    planId: logQuery.planId || undefined,
    level: logQuery.level || undefined,
  })
    .then((res: ApiResponse<Models.PaginationResponse<PlanLogResult>>) => {
      if (res.data) {
        logTable.value = res.data.data
        logPagination.itemCount = res.data.total
      }
    })
    .catch((err: unknown) => {
      console.error('获取日志失败:', err)
    })
    .finally(() => {
      logLoading.value = false
      refreshTime.value = dayjs()
    })
}

// Utils
const formatDT = (v?: string) => (v ? dayjs(v).format('YYYY-MM-DD HH:mm:ss') : '-')

// Init
onMounted(() => {
  loadCloudTokens()
  fetchPlanList()
  fetchLogList()
})
</script>

<style scoped>
.autoingest-page {
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

/* 头部样式参考用户组管理 */
.header {
  margin: 8px 0 4px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-left {
  display: flex;
  align-items: center;
}

.autoingest-table {
  background: var(--n-card-color);
  border-radius: 6px;
}

.autoingest-table :deep(.n-data-table-th) {
  text-align: center;
  font-weight: 600;
}

.autoingest-table :deep(.n-data-table-td) {
  text-align: center;
}

.cell-title {
  font-weight: 600;
}

.log-content {
  white-space: pre-wrap;
  word-break: break-all;
}

/* 时间列样式（合并为一列） */
.time-cell {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.time-line {
  display: flex;
  gap: 6px;
  align-items: center;
  font-size: 12px;
}

.time-label {
  color: var(--n-text-color-2);
}

.time-value {
  color: var(--n-text-color);
}
</style>
