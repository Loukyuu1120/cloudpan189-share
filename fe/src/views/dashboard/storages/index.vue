<template>
  <div class="storages-page">
    <!-- 头部区域 -->
    <div class="header">
      <div class="header-left">
        <n-input
          v-model:value="searchKeyword"
          placeholder="请输入路径搜索"
          clearable
          style="width: 200px; margin-right: 12px"
          @keyup.enter="handleSearch"
        >
          <template #prefix>
            <n-icon :size="16" :depth="3">
              <SearchOutline />
            </n-icon>
          </template>
        </n-input>
        <n-select
          v-model:value="selectedTaskLogStatus"
          placeholder="扫描状态"
          clearable
          style="width: 120px; margin-right: 12px"
          :options="taskLogStatusOptions"
          @update:value="handleSearch"
        />
        <n-button type="primary" @click="handleSearch" style="margin-right: 8px">
          <template #icon>
            <n-icon>
              <SearchOutline />
            </n-icon>
          </template>
          搜索
        </n-button>
        <n-button @click="handleReset">
          <template #icon>
            <n-icon>
              <RefreshOutline />
            </n-icon>
          </template>
          重置
        </n-button>
        <n-text v-if="pageAutoRefreshStore.autoRefreshEnabled">
          上次刷新时间：{{ refreshTime.format('YYYY-MM-DD HH:mm:ss') }}
        </n-text>
      </div>
      <div class="header-right batch-mode">
        <template v-if="isBatchMode">
          <n-button
            :type="isAllSelected ? 'warning' : 'default'"
            @click="toggleSelectAll"
            class="batch-btn"
          >
            {{ isAllSelected ? '取消当页' : '全选当页' }}
          </n-button>
          <n-button
            :type="isAllSelectedAllPages ? 'warning' : 'info'"
            @click="selectAllPages"
            :loading="isSelectingAllPages"
            class="batch-btn"
          >
            {{ isAllSelectedAllPages ? '取消全选' : '全选所有' }}
          </n-button>
          <n-button
            type="info"
            @click="handleBatchRefresh(false)"
            class="batch-btn"
            :disabled="selectedIds.length === 0"
          >
            <template #icon><n-icon><RefreshOutline /></n-icon></template>
            刷新选中 ({{ selectedIds.length }})
          </n-button>
          <n-button
            type="warning"
            @click="handleBatchRefresh(true)"
            class="batch-btn"
            :disabled="selectedIds.length === 0"
          >
            <template #icon><n-icon><RefreshOutline /></n-icon></template>
            深度刷新 ({{ selectedIds.length }})
          </n-button>
          <n-button
            type="primary"
            @click="showBatchModifyTokenModal = true"
            class="batch-btn"
            :disabled="selectedIds.length === 0"
          >
            <template #icon><n-icon><KeyOutline /></n-icon></template>
            修改令牌 ({{ selectedIds.length }})
          </n-button>
          <n-button
            type="error"
            @click="handleBatchDelete"
            class="batch-btn"
            :disabled="selectedIds.length === 0"
          >
            <template #icon><n-icon><TrashOutline /></n-icon></template>
            删除选中 ({{ selectedIds.length }})
          </n-button>
          <n-button @click="exitBatchMode" class="batch-btn">取消</n-button>
        </template>
        <template v-else>
          <n-button @click="enterBatchMode" style="margin-right: 12px">批量管理</n-button>
        </template>
        <n-tooltip trigger="hover">
          <template #trigger>
            <n-button text @click="handlePageSettings" style="margin-right: 8px; font-size: 16px">
              <template #icon>
                <n-icon>
                  <SettingsOutline />
                </n-icon>
              </template>
            </n-button>
          </template>
          页面设置
        </n-tooltip>
        <!-- 下拉菜单 -->
        <n-dropdown trigger="click" :options="addMountOptions" @select="handleSelectMountType">
          <n-button type="primary">
            <template #icon>
              <n-icon>
                <AddOutline />
              </n-icon>
            </template>
            新增挂载
          </n-button>
        </n-dropdown>
      </div>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading" class="loading-container">
      <n-spin size="large">
        <template #description>
          <n-text depth="2">正在加载存储数据...</n-text>
        </template>
      </n-spin>
    </div>

    <!-- 存储卡片列表 -->
    <div class="storage-cards">
      <n-card
        v-for="storage in tableData"
        :key="storage.id"
        class="storage-card"
        :class="{ 'is-selected': selectedIds.includes(storage.id) }"
        hoverable
        :bordered="false"
        @click="handleCardClick(storage.id)"
      >
        <!-- 选择遮罩 -->
        <div v-if="isBatchMode" class="selection-overlay">
          <n-checkbox
            :checked="selectedIds.includes(storage.id)"
            class="selection-checkbox"
            size="large"
            @click.stop="toggleSelection(storage.id)"
          />
        </div>
        <!-- 存储卡片内容 -->
        <template #header>
          <div class="card-header">
            <div class="storage-info">
              <div class="storage-title">
                <n-text strong class="storage-name">
                  {{ storage.name || '未命名存储' }}
                </n-text>
              </div>
              <n-ellipsis class="storage-path" :tooltip="{ placement: 'top' }">
                {{ storage.fullPath || '-' }}
              </n-ellipsis>
            </div>

            <!-- 非批量模式才显示操作按钮 -->
            <div class="storage-actions" v-if="!isBatchMode">
              <n-button size="small" quaternary circle @click="handleModifyToken(storage)">
                <template #icon>
                  <n-icon :size="16">
                    <KeyOutline />
                  </n-icon>
                </template>
              </n-button>

              <n-button size="small" quaternary circle @click="handleDelete(storage)">
                <template #icon>
                  <n-icon :size="16">
                    <TrashOutline />
                  </n-icon>
                </template>
              </n-button>

              <n-dropdown
                :options="getRefreshOptions(storage.id)"
                @select="handleRefreshSelect"
                trigger="click"
              >
                <n-button size="small" quaternary circle>
                  <template #icon>
                    <n-icon :size="16">
                      <RefreshOutline />
                    </n-icon>
                  </template>
                </n-button>
              </n-dropdown>
            </div>
          </div>
        </template>

        <div class="card-content">
          <!-- 基础信息区域 - 使用flex容器包裹前三个字段 -->
          <div class="basic-info-container">
            <div class="info-item">
              <div class="info-label">
                <n-icon :size="14" class="info-icon">
                  <FolderOutline />
                </n-icon>
                <span>存储类型</span>
              </div>
              <n-tag :color="getOsTypeColor(storage.osType)" size="small" class="info-tag">
                {{ getOsTypeDisplayName(storage.osType) }}
              </n-tag>
            </div>

            <div class="info-item">
              <div class="info-label">
                <n-icon :size="14" class="info-icon">
                  <KeyOutline />
                </n-icon>
                <span>绑定令牌</span>
              </div>
              <n-text class="info-value">{{ storage.tokenName || '未绑定' }}</n-text>
            </div>

            <div class="info-item">
              <div class="info-label">
                <n-icon :size="14" class="info-icon">
                  <DocumentsOutline />
                </n-icon>
                <span>文件数量</span>
              </div>
              <n-text class="info-value">{{ storage.fileCount || 0 }}</n-text>
            </div>
          </div>

          <div class="additional-info">
            <div class="info-item">
              <div class="refresh-header">
                <div class="refresh-title-section">
                  <div class="info-label">
                    <n-icon :size="14" class="info-icon">
                      <RefreshOutline />
                    </n-icon>
                    <span>自动刷新</span>
                  </div>
                  <n-tag
                    v-if="storage.enableAutoRefresh"
                    :type="storage.isInAutoRefreshPeriod ? 'success' : 'warning'"
                    size="small"
                  >
                    {{ computedRefreshStatusText(storage) }}
                  </n-tag>
                  <n-tag v-else type="default" size="small">未启用</n-tag>
                </div>
                <n-button size="tiny" type="primary" @click="handleEditAutoRefresh(storage)">
                  编辑
                </n-button>
              </div>
              <div v-if="storage.enableAutoRefresh" class="refresh-details">
                <div class="refresh-status">
                  <n-text depth="3" class="refresh-detail"> {{ storage.refreshInterval }}m </n-text>
                  <n-text depth="3" class="refresh-detail">
                    {{ storage.enableDeepRefresh ? '深度刷新' : '普通刷新' }}
                  </n-text>
                </div>
                <n-text depth="3" class="refresh-period">
                  {{ formatRefreshPeriod(storage) }}
                </n-text>
              </div>
            </div>

            <!-- 最近一次运行日志 -->
            <div v-if="storage.taskLogs && storage.taskLogs.length > 0" class="info-item">
              <div class="task-log-header">
                <div class="task-log-first-row">
                  <n-popover trigger="hover">
                    <template #trigger>
                      <div class="info-label">
                        <n-icon :size="14" class="info-icon">
                          <component :is="getTaskStatusInfo(storage.taskLogs[0].status).icon" />
                        </n-icon>
                        <span>最近运行</span>
                        <n-text depth="3" class="task-log-time">
                          {{ formatTaskLogTime(storage.taskLogs[0]) }}
                        </n-text>
                      </div>
                    </template>
                    <n-text depth="1"> {{ storage.taskLogs[0].title }}; </n-text>
                    <n-text depth="2">
                      {{ storage.taskLogs[0].desc }}
                    </n-text>
                  </n-popover>
                  <n-popover trigger="hover" :disabled="storage.taskLogs[0].result ? false : true">
                    <template #trigger>
                      <n-tag :type="getTaskStatusInfo(storage.taskLogs[0].status).type" size="small">
                        {{ getTaskStatusInfo(storage.taskLogs[0].status).text }}
                      </n-tag>
                    </template>
                    {{ storage.taskLogs[0].result }}
                  </n-popover>
                </div>
                <div v-if="storage.enableAutoRefresh" class="next-run-time">
                  <n-text depth="3">下次 {{ formatNextRunTime(storage) }}</n-text>
                </div>
              </div>
              <div class="task-log-content"></div>
            </div>
          </div>
        </div>

        <template #footer>
          <div class="card-footer">
            <div class="footer-time">
              <n-icon :size="12" class="footer-icon">
                <TimeOutline />
              </n-icon>
              <n-text depth="3" class="footer-text">
                创建时间：{{ formatDateTime(storage.createdAt) }}
              </n-text>
            </div>
            <div class="footer-time">
              <n-icon :size="12" class="footer-icon">
                <TimeOutline />
              </n-icon>
              <n-text depth="3" class="footer-text">
                更新时间：{{ formatDateTime(storage.updatedAt) }}
              </n-text>
            </div>
          </div>
        </template>
      </n-card>

      <!-- 空状态 -->
      <div v-if="tableData.length === 0" class="empty-state">
        <n-empty description="暂无存储数据" size="large">
          <template #icon>
            <n-icon size="64" :depth="3">
              <ServerOutline />
            </n-icon>
          </template>
          <template #extra>
            <n-text depth="3">您还没有配置任何存储挂载点</n-text>
          </template>
        </n-empty>
      </div>
    </div>

    <!-- 分页 -->
    <div v-if="!loading && tableData.length > 0" class="pagination-container">
      <n-pagination
        v-model:page="paginationReactive.page"
        v-model:page-size="paginationReactive.pageSize"
        :item-count="paginationReactive.itemCount"
        :page-sizes="paginationReactive.pageSizes"
        show-size-picker
        @update:page="handlePageChange"
        @update:page-size="handlePageSizeChange"
      >
        <template #prefix="{ itemCount }"> 共 {{ itemCount }} 项 </template>
      </n-pagination>
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
          <n-form-item label="启用自动刷新" path="enableAutoRefresh">
            <n-switch v-model:value="autoRefreshForm.enableAutoRefresh" />
          </n-form-item>

          <template v-if="autoRefreshForm.enableAutoRefresh">
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

            <n-form-item label="开始日期" path="refreshBeginAt">
              <n-date-picker
                v-model:value="autoRefreshForm.refreshBeginAt"
                type="date"
                placeholder="选择开始日期"
                style="width: 100%"
              />
            </n-form-item>

            <n-form-item label="深度刷新" path="enableDeepRefresh">
              <n-switch v-model:value="autoRefreshForm.enableDeepRefresh" />
            </n-form-item>
          </template>
        </n-form>
      </div>

      <template #action>
        <n-button @click="showAutoRefreshModal = false">取消</n-button>
        <n-button type="primary" @click="handleAutoRefreshConfirm" :loading="autoRefreshSubmitting">
          确认
        </n-button>
      </template>
    </n-modal>

    <!-- 页面设置弹窗 -->
    <n-modal
      v-model:show="showPageSettingsModal"
      preset="dialog"
      title="页面设置"
      style="width: 420px"
    >
      <div class="page-settings-config">
        <div class="settings-section">
          <div class="setting-item">
            <div class="setting-label">启用自动刷新</div>
            <n-switch
              v-model:value="pageSettingsForm.autoRefreshEnabled"
              @update:value="handlePageAutoRefreshToggle"
              size="medium"
            >
              <template #checked>已开启</template>
              <template #unchecked>已关闭</template>
            </n-switch>
          </div>

          <div v-if="pageSettingsForm.autoRefreshEnabled" class="setting-item">
            <div class="setting-label">刷新间隔</div>
            <n-select
              v-model:value="pageSettingsForm.refreshInterval"
              :options="refreshIntervalOptions"
              style="width: 160px"
              @update:value="handlePageRefreshIntervalChange"
              size="small"
            />
          </div>

          <div v-if="pageSettingsForm.autoRefreshEnabled" class="setting-item">
            <div class="setting-label">下次刷新</div>
            <n-text depth="3">{{ nextRefreshTime.format('HH:mm:ss') }}</n-text>
          </div>
        </div>
      </div>

      <template #action>
        <n-button @click="showPageSettingsModal = false">关闭</n-button>
      </template>
    </n-modal>

    <!-- 修改令牌弹窗 -->
    <n-modal v-model:show="showModifyTokenModal" preset="dialog" title="修改绑定令牌">
      <div class="modify-token-config">
        <n-form label-placement="left" label-width="100px">
          <n-form-item label="存储名称">
            <n-text>{{ currentModifyStorage?.name || '未命名存储' }}</n-text>
          </n-form-item>

          <n-form-item label="当前令牌">
            <n-text depth="3">{{ currentModifyStorage?.tokenName || '未绑定' }}</n-text>
          </n-form-item>

          <n-form-item label="选择令牌">
            <n-select
              v-model:value="selectedTokenId"
              :options="cloudTokenOptions"
              placeholder="请选择要绑定的令牌"
              clearable
              style="width: 100%"
            />
          </n-form-item>
        </n-form>
      </div>

      <template #action>
        <n-button @click="showModifyTokenModal = false">取消</n-button>
        <n-button type="primary" @click="handleModifyTokenConfirm" :loading="modifyTokenSubmitting">
          确认修改
        </n-button>
      </template>
    </n-modal>

    <!-- 批量修改令牌弹窗 -->
    <n-modal v-model:show="showBatchModifyTokenModal" preset="dialog" title="批量修改令牌">
      <div class="modify-token-config">
        <n-form label-placement="left" label-width="100px">
          <n-form-item label="已选中">
            <n-text>{{ selectedIds.length }} 个挂载点</n-text>
          </n-form-item>

          <n-form-item label="选择令牌">
            <n-select
              v-model:value="batchModifyTokenId"
              :options="cloudTokenOptions"
              placeholder="请选择要绑定的令牌"
              clearable
              style="width: 100%"
            />
          </n-form-item>
        </n-form>
      </div>

      <template #action>
        <n-button @click="showBatchModifyTokenModal = false">取消</n-button>
        <n-button type="primary" @click="handleBatchModifyTokenConfirm" :loading="modifyTokenSubmitting">
          确认修改
        </n-button>
      </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted, computed, h } from 'vue' // 引入 h
import {
  NCheckbox,
  NInput,
  NButton,
  NCard,
  NText,
  NTag,
  NSpin,
  NEmpty,
  NPagination,
  NEllipsis,
  NIcon,
  NModal,
  NDropdown,
  NForm,
  NFormItem,
  NSwitch,
  NInputNumber,
  NDatePicker,
  NSelect,
  NTooltip,
  useMessage,
  useDialog,
  type PaginationProps,
  type DropdownOption,
  type FormInst,
  type FormRules,
} from 'naive-ui'
import {
  SearchOutline,
  RefreshOutline,
  FolderOutline,
  KeyOutline,
  TimeOutline,
  ServerOutline,
  AddOutline,
  TrashOutline,
  DocumentsOutline,
  SettingsOutline,
  ClipboardOutline, // 引入剪贴板图标用于批量文本导入
} from '@vicons/ionicons5'
import {
  getStorageList,
  refreshStorage,
  deleteStorage,
  toggleAutoRefresh,
  modifyToken,
  batchDeleteStorage,
  batchRefreshStorage,
  batchModifyToken,
} from '@/api/storage'
import type { StorageInfo } from '@/api/storage'
import { getCloudTokenList } from '@/api/cloudtoken'
import { formatDateTime } from '@/utils/time'
import { getOsTypeDisplayName, getOsTypeColor, mountTypeConfigs } from '@/utils/osType'
import { getTaskStatusInfo } from '@/utils/taskStatus'
import { useSubscribeMount } from '@/composables/useSubscribeMount'
import { useShareMount } from '@/composables/useShareMount'
import { usePersonMount } from '@/composables/usePersonMount'
import { useFamilyMount } from '@/composables/useFamilyMount'
import { useBatchCreateFromTextMount } from '@/composables/useBatchCreateFromTextMount' // 引入新组件的hook
import { usePageAutoRefreshStore } from '@/stores/modules/pageAutoRefresh'
import dayjs from 'dayjs'

// 表格数据
const tableData = reactive<StorageInfo[]>([])
const loading = ref(false)
const searchKeyword = ref('')
const selectedTaskLogStatus = ref<string>('')

// 弹窗控制
const showAutoRefreshModal = ref(false)
const showPageSettingsModal = ref(false)
const showModifyTokenModal = ref(false)
const showBatchModifyTokenModal = ref(false)
const batchModifyTokenId = ref<number | null>(null)

const subscribeMount = useSubscribeMount()
const shareMount = useShareMount()
const personMount = usePersonMount()
const familyMount = useFamilyMount()
const batchTextMount = useBatchCreateFromTextMount() // 初始化批量文本导入

// 自动刷新配置表单
const autoRefreshFormRef = ref<FormInst>()
const autoRefreshSubmitting = ref(false)
const currentEditStorage = ref<StorageInfo | null>(null)

const autoRefreshForm = ref({
  enableAutoRefresh: false,
  refreshInterval: 60,
  autoRefreshDays: 7,
  refreshBeginAt: null as number | null,
  enableDeepRefresh: false,
})

const autoRefreshRules: FormRules = {
  refreshInterval: [
    {
      type: 'number',
      min: 30,
      max: 1440,
      message: '刷新间隔必须在30-1440分钟之间',
      trigger: 'blur',
    },
  ],
  autoRefreshDays: [
    {
      type: 'number',
      min: 1,
      max: 365,
      message: '持续天数必须在1-365天之间',
      trigger: 'blur',
    },
  ],
  refreshBeginAt: [],
}

// 消息提示和对话框
const message = useMessage()
const dialog = useDialog()

// 分页配置
const paginationReactive = reactive<PaginationProps>({
  page: 1,
  pageSize: 12,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [12, 24, 48, 96],
})

// 分页处理函数
const handlePageChange = (page: number) => {
  paginationReactive.page = page
  fetchStorageList()
}

const handlePageSizeChange = (pageSize: number) => {
  paginationReactive.pageSize = pageSize
  paginationReactive.page = 1
  fetchStorageList()
}

const refreshTime = ref(dayjs())

// 计算下一次运行时间
const getNextRunTime = (storage: StorageInfo) => {
  if (!storage.enableAutoRefresh || !storage.refreshInterval) return null
  const lastRun = storage.updatedAt ? dayjs(storage.updatedAt) : null
  if (!lastRun) return null
  return lastRun.add(storage.refreshInterval, 'minute')
}

// 格式化下一次运行时间
const formatNextRunTime = (storage: StorageInfo) => {
  const nextRun = getNextRunTime(storage)
  if (!nextRun) return ''
  const now = dayjs()
  if (nextRun.isBefore(now)) {
    return '即将运行'
  }
  return nextRun.format('MM-DD HH:mm')
}

// 使用页面自动刷新store
const pageAutoRefreshStore = usePageAutoRefreshStore()

// 自动刷新状态管理
const nextRefreshTime = ref(dayjs().add(pageAutoRefreshStore.refreshInterval, 'second'))
const intervalTimer = ref<NodeJS.Timeout | null>(null)

// 页面设置表单
const pageSettingsForm = ref({
  autoRefreshEnabled: pageAutoRefreshStore.autoRefreshEnabled,
  refreshInterval: pageAutoRefreshStore.refreshInterval,
})

// 刷新间隔选项（秒）
const refreshIntervalOptions = [
  { label: '30秒', value: 30 },
  { label: '1分钟', value: 60 },
  { label: '3分钟', value: 180 },
  { label: '5分钟', value: 300 },
  { label: '10分钟', value: 600 },
]

// 启动自动刷新定时器
const startAutoRefresh = () => {
  if (intervalTimer.value) {
    clearInterval(intervalTimer.value)
  }

  if (pageAutoRefreshStore.autoRefreshEnabled) {
    updateNextRefreshTime()
    intervalTimer.value = setInterval(() => {
      fetchStorageList()
      updateNextRefreshTime()
    }, pageAutoRefreshStore.refreshInterval * 1000)
  }
}

// 停止自动刷新定时器
const stopAutoRefresh = () => {
  if (intervalTimer.value) {
    clearInterval(intervalTimer.value)
    intervalTimer.value = null
  }
}

// 更新下次刷新时间
const updateNextRefreshTime = () => {
  nextRefreshTime.value = dayjs().add(pageAutoRefreshStore.refreshInterval, 'second')
}

// 打开页面设置
const handlePageSettings = () => {
  pageSettingsForm.value = {
    autoRefreshEnabled: pageAutoRefreshStore.autoRefreshEnabled,
    refreshInterval: pageAutoRefreshStore.refreshInterval,
  }
  showPageSettingsModal.value = true
}

// 处理页面设置中的自动刷新开关切换
const handlePageAutoRefreshToggle = (enabled: boolean) => {
  pageSettingsForm.value.autoRefreshEnabled = enabled
  pageAutoRefreshStore.updateSettings({ autoRefreshEnabled: enabled })

  if (enabled) {
    startAutoRefresh()
    message.success('自动刷新已开启')
  } else {
    stopAutoRefresh()
    message.info('自动刷新已关闭')
  }
}

// 处理页面设置中的刷新间隔变更
const handlePageRefreshIntervalChange = (interval: number) => {
  pageSettingsForm.value.refreshInterval = interval
  pageAutoRefreshStore.updateSettings({ refreshInterval: interval })

  if (pageAutoRefreshStore.autoRefreshEnabled) {
    startAutoRefresh()
    message.success(`刷新间隔已更改为 ${Math.floor(interval / 60)} 分钟`)
  }
}

// 获取存储列表
const fetchStorageList = () => {
  loading.value = true

  const params = {
    currentPage: paginationReactive.page || 1,
    pageSize: paginationReactive.pageSize || 10,
    path: searchKeyword.value || undefined,
    taskLogStatus: selectedTaskLogStatus.value || undefined,
  }

  getStorageList(params)
    .then((response) => {
      if (response.data) {
        tableData.splice(0, tableData.length, ...response.data.data)
        paginationReactive.itemCount = response.data.total
      }
    })
    .catch((error) => {
      console.error('获取存储列表失败:', error)
      message.error(error?.message || '获取存储列表失败')
    })
    .finally(() => {
      loading.value = false
      refreshTime.value = dayjs()
    })
}

// 任务日志状态筛选选项
const taskLogStatusOptions = [
  { label: '全部', value: '' },
  { label: '失败', value: 'failed' },
  { label: '成功', value: 'completed' },
  // { label: '进行中', value: 'running' },
  // { label: '等待中', value: 'pending' },
]

// 搜索
const handleSearch = () => {
  paginationReactive.page = 1
  fetchStorageList()
}

// 重置
const handleReset = () => {
  searchKeyword.value = ''
  selectedTaskLogStatus.value = ''
  paginationReactive.page = 1
  fetchStorageList()
}

// 挂载类型配置
const mountTypes = mountTypeConfigs

// 批量文本导入
const addMountOptions = computed(() => {
  const options: DropdownOption[] = mountTypes.map((type) => ({
    label: type.label,
    key: type.value,
  }))

  // 添加分割线
  options.push({
    type: 'divider',
    key: 'divider-batch',
  })

  // 添加批量文本导入选项
  options.push({
    label: '批量文本导入',
    key: 'batch_text_import',
    icon: () => h(NIcon, null, { default: () => h(ClipboardOutline) }),
  })

  return options
})

// 选择挂载类型
const handleSelectMountType = (mountType: string) => {
  console.log('选择的挂载类型:', mountType)

  if (mountType === 'subscribe') {
    subscribeMount.show().then(addNewStorageCallback)
  } else if (mountType === 'share_folder') {
    shareMount.show().then(addNewStorageCallback)
  } else if (mountType === 'person_folder') {
    personMount.show().then(addNewStorageCallback)
  } else if (mountType === 'family_folder') {
    familyMount.show().then(addNewStorageCallback)
  } else if (mountType === 'batch_text_import') {
    // 处理批量文本导入
    batchTextMount.show().then(addNewStorageCallback)
  } else {
    message.info(
      `您选择了：${mountTypes.find((t) => t.value === mountType)?.label}，该功能正在开发中`
    )
  }
}

const addNewStorageCallback = (data: { success: boolean }) => {
  if (data.success) {
    message.success('操作成功')
  }
  // 无论成功与否，如果是成功的话通常需要刷新列表
  // 这里可以根据 data.success 判断，或者简单地都刷新
  fetchStorageList()
}

// 获取刷新选项
const getRefreshOptions = (storageId: number): DropdownOption[] => {
  return [
    {
      label: '普通刷新',
      key: `normal-${storageId}`,
      props: {
        onClick: () => handleRefresh(storageId, false),
      },
    },
    {
      label: '深度刷新',
      key: `deep-${storageId}`,
      props: {
        onClick: () => handleRefresh(storageId, true),
      },
    },
  ]
}

// 处理刷新选择
const handleRefreshSelect = (key: string) => {
  console.log('刷新选择:', key)
}

// 处理刷新
const handleRefresh = (storageId: number, deep: boolean) => {
  const storage = tableData.find((s) => s.id === storageId)
  const refreshType = deep ? '深度刷新' : '普通刷新'

  message.loading(`正在执行${refreshType}...`)

  refreshStorage({ id: storageId, deep })
    .then(() => {
      message.success(`${storage?.name || '存储'} ${refreshType}成功`)
      fetchStorageList()
    })
    .catch((error) => {
      console.error('刷新存储失败:', error)
      message.error(error?.message || '刷新失败')
    })
}

// 处理删除
const handleDelete = (storage: StorageInfo) => {
  dialog.warning({
    title: '确认删除',
    content: `确定要删除存储挂载点 "${storage.name || '未命名存储'}" 吗？此操作不可撤销。`,
    positiveText: '确认删除',
    negativeText: '取消',
    onPositiveClick: () => {
      message.loading(`正在删除 ${storage.name || '存储'}...`)

      deleteStorage({ id: storage.id })
        .then(() => {
          message.success(`${storage.name || '存储'} 删除成功`)
          fetchStorageList()
        })
        .catch((error) => {
          console.error('删除存储失败:', error)
          message.error(error?.message || '删除失败')
        })
    },
  })
}

// 计算属性：当前页面展示的所有 ID
const currentViewIds = computed(() => tableData.map((item) => item.id))

// 计算属性：是否已全选当前页
const isAllSelected = computed(() => {
  if (tableData.length === 0) return false
  return currentViewIds.value.every((id) => selectedIds.value.includes(id))
})

// 处理全选/取消全选
const toggleSelectAll = () => {
  if (isAllSelected.value) {
    selectedIds.value = selectedIds.value.filter((id) => !currentViewIds.value.includes(id))
  } else {
    const newIds = currentViewIds.value.filter((id) => !selectedIds.value.includes(id))
    selectedIds.value.push(...newIds)
  }
}

// 全选所有页
const isSelectingAllPages = ref(false)
const isAllSelectedAllPages = computed(() => {
  const total = paginationReactive.itemCount || 0
  return total > 0 && selectedIds.value.length === total
})
const selectAllPages = () => {
  const total = paginationReactive.itemCount || 0
  if (selectedIds.value.length === total) {
    selectedIds.value = []
    return
  }
  isSelectingAllPages.value = true
  // 保存当前分页设置
  const originalPageSize = paginationReactive.pageSize
  // 获取所有数据
  getStorageList({ currentPage: 1, pageSize: total })
    .then((res) => {
      if (res.data?.data) {
        const allIds = res.data.data.map((item: StorageInfo) => item.id)
        selectedIds.value = [...allIds]
      }
    })
    .catch((error) => {
      console.error('获取全量数据失败:', error)
      message.error('获取全量数据失败')
    })
    .finally(() => {
      isSelectingAllPages.value = false
      // 恢复分页设置并刷新列表
      paginationReactive.pageSize = originalPageSize
      fetchStorageList()
    })
}

// 批量操作状态
const isBatchMode = ref(false)
const selectedIds = ref<number[]>([])

// 进入批量模式
const enterBatchMode = () => {
  isBatchMode.value = true
  selectedIds.value = []
}

// 退出批量模式
const exitBatchMode = () => {
  isBatchMode.value = false
  selectedIds.value = []
}

// 切换选中状态
const toggleSelection = (id: number) => {
  if (selectedIds.value.includes(id)) {
    selectedIds.value = selectedIds.value.filter((item) => item !== id)
  } else {
    selectedIds.value.push(id)
  }
}

// 处理卡片点击
const handleCardClick = (id: number) => {
  if (isBatchMode.value) {
    toggleSelection(id)
  }
}

// 处理批量删除
const handleBatchDelete = () => {
  if (selectedIds.value.length === 0) return

  dialog.warning({
    title: '批量删除',
    content: `确定要删除选中的 ${selectedIds.value.length} 个挂载点吗？此操作不可撤销。`,
    positiveText: '确认删除',
    negativeText: '取消',
    onPositiveClick: () => {
      message.loading('正在批量删除...')

      batchDeleteStorage({ ids: selectedIds.value })
        .then(() => {
          message.success('批量删除成功')
          exitBatchMode()
          fetchStorageList()
        })
        .catch((error) => {
          message.error(error?.message || '批量删除失败')
        })
    },
  })
}

// 处理批量刷新
const handleBatchRefresh = (deep: boolean) => {
  if (selectedIds.value.length === 0) return

  const refreshType = deep ? '深度刷新' : '普通刷新'
  
  dialog.warning({
    title: `批量${refreshType}`,
    content: `确定要${refreshType}选中的 ${selectedIds.value.length} 个挂载点吗？`,
    positiveText: '确认刷新',
    negativeText: '取消',
    onPositiveClick: () => {
      message.loading(`正在批量${refreshType}...`)

      batchRefreshStorage({ ids: selectedIds.value, deep })
        .then((res) => {
          message.success(res?.msg || `批量${refreshType}任务已提交`)
          exitBatchMode()
        })
        .catch((error) => {
          message.error(error?.message || `批量${refreshType}失败`)
        })
    },
  })
}

// 确认批量修改令牌
const handleBatchModifyTokenConfirm = () => {
  if (selectedIds.value.length === 0 || !batchModifyTokenId.value) return
  
  modifyTokenSubmitting.value = true
  
  batchModifyToken({
    ids: selectedIds.value,
    tokenId: batchModifyTokenId.value,
  })
    .then((res) => {
      message.success(res?.msg || '批量修改令牌成功')
      showBatchModifyTokenModal.value = false
      exitBatchMode()
      fetchStorageList()
    })
    .catch((error) => {
      console.error('批量修改令牌失败:', error)
      message.error(error?.message || '批量修改令牌失败')
    })
    .finally(() => {
      modifyTokenSubmitting.value = false
    })
}

// 处理编辑自动刷新
const handleEditAutoRefresh = (storage: StorageInfo) => {
  currentEditStorage.value = storage
  autoRefreshForm.value = {
    enableAutoRefresh: storage.enableAutoRefresh || false,
    refreshInterval: storage.refreshInterval || 60,
    autoRefreshDays: storage.autoRefreshDays || 7,
    refreshBeginAt: storage.autoRefreshBeginAt
      ? new Date(storage.autoRefreshBeginAt).getTime()
      : Date.now(),
    enableDeepRefresh: storage.enableDeepRefresh || false,
  }
  showAutoRefreshModal.value = true
}

// 处理自动刷新配置确认
const handleAutoRefreshConfirm = () => {
  if (!currentEditStorage.value) return

  autoRefreshFormRef.value?.validate((errors) => {
    if (errors) {
      message.error('请检查表单输入')
      return
    }

    autoRefreshSubmitting.value = true

    const refreshBeginAt = autoRefreshForm.value.refreshBeginAt
      ? dayjs(autoRefreshForm.value.refreshBeginAt).format('YYYY-MM-DD')
      : dayjs().format('YYYY-MM-DD')

    toggleAutoRefresh({
      id: currentEditStorage.value!.id,
      enableAutoRefresh: autoRefreshForm.value.enableAutoRefresh,
      refreshInterval: autoRefreshForm.value.enableAutoRefresh
        ? autoRefreshForm.value.refreshInterval
        : undefined,
      autoRefreshDays: autoRefreshForm.value.enableAutoRefresh
        ? autoRefreshForm.value.autoRefreshDays
        : undefined,
      refreshBeginAt: autoRefreshForm.value.enableAutoRefresh ? refreshBeginAt : undefined,
      enableDeepRefresh: autoRefreshForm.value.enableAutoRefresh
        ? autoRefreshForm.value.enableDeepRefresh
        : undefined,
    })
      .then(() => {
        message.success('自动刷新配置更新成功')
        showAutoRefreshModal.value = false
        fetchStorageList()
      })
      .catch((error) => {
        console.error('更新自动刷新配置失败:', error)
        message.error(error?.message || '配置更新失败')
      })
      .finally(() => {
        autoRefreshSubmitting.value = false
      })
  })
}

// 格式化刷新周期显示
const formatRefreshPeriod = (storage: StorageInfo) => {
  if (!storage.autoRefreshBeginAt || !storage.autoRefreshDays) {
    return '未设置刷新周期'
  }
  const beginDate = new Date(storage.autoRefreshBeginAt)
  const endDate = new Date(beginDate)
  endDate.setDate(beginDate.getDate() + storage.autoRefreshDays)

  const formatDate = (date: Date) => {
    return date.toLocaleDateString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
    })
  }
  return `${formatDate(beginDate)} ~ ${formatDate(endDate)} (${storage.autoRefreshDays}天)`
}

const computedRefreshStatusText = (storage: StorageInfo): string => {
  if (storage.isInAutoRefreshPeriod) {
    return '待执行'
  }
  let refreshBeginAt = dayjs(storage.autoRefreshBeginAt)
  if (refreshBeginAt.isAfter(dayjs())) {
    return '未到开始时间'
  } else {
    return '已失效'
  }
}

// 修改令牌相关变量
const modifyTokenSubmitting = ref(false)
const currentModifyStorage = ref<StorageInfo | null>(null)
const cloudTokenOptions = ref<{ label: string; value: number }[]>([])
const selectedTokenId = ref<number | null>(null)

// 处理修改令牌
const handleModifyToken = (storage: StorageInfo) => {
  currentModifyStorage.value = storage
  selectedTokenId.value = storage.tokenId || null

  getCloudTokenList({ noPaginate: true })
    .then((response) => {
      if (response.data) {
        cloudTokenOptions.value = response.data.data.map((token) => ({
          label: token.name,
          value: token.id,
        }))
        cloudTokenOptions.value.unshift({
          label: '解绑令牌',
          value: 0,
        })
        showModifyTokenModal.value = true
      }
    })
    .catch((error) => {
      console.error('获取云盘令牌列表失败:', error)
      message.error(error?.message || '获取令牌列表失败')
    })
}

// 确认修改令牌
const handleModifyTokenConfirm = () => {
  if (!currentModifyStorage.value) return
  modifyTokenSubmitting.value = true
  const tokenId = selectedTokenId.value === 0 ? 0 : selectedTokenId.value || 0

  modifyToken({
    id: currentModifyStorage.value.id,
    tokenId,
  })
    .then(() => {
      const actionText = tokenId === 0 ? '解绑' : '修改绑定'
      message.success(`令牌${actionText}成功`)
      showModifyTokenModal.value = false
      fetchStorageList()
    })
    .catch((error) => {
      console.error('修改令牌失败:', error)
      message.error(error?.message || '令牌修改失败')
    })
    .finally(() => {
      modifyTokenSubmitting.value = false
    })
}

// 格式化任务日志时间
const formatTaskLogTime = (taskLog: Models.FileTaskLog) => {
  if (!taskLog.beginAt) return '未知时间'
  const beginTime = dayjs(taskLog.beginAt)
  const startTime = beginTime.format('MM-DD HH:mm')

  if ((taskLog.status === 'completed' || taskLog.status === 'failed') && taskLog.duration) {
    const duration = taskLog.duration
    let durationText = ''
    if (duration < 1000) {
      durationText = `${duration}ms`
    } else if (duration < 60000) {
      durationText = `${Math.round(duration / 1000)}s`
    } else {
      const minutes = Math.floor(duration / 60000)
      const seconds = Math.round((duration % 60000) / 1000)
      durationText = `${minutes}m${seconds}s`
    }
    return `${startTime} (用时 ${durationText})`
  }
  return startTime
}

// 初始化
onMounted(() => {
  console.log('页面挂载，开始获取数据')
  fetchStorageList()
  if (pageAutoRefreshStore.autoRefreshEnabled) {
    startAutoRefresh()
  }
})

onUnmounted(() => {
  console.log('页面卸载，清除定时器')
  stopAutoRefresh()
})
</script>

<style scoped>
/* 样式部分保持不变 */

/* 页面整体样式 */
.storages-page {
  padding: 24px;
  background: var(--n-color-target);
  flex: 1;
}

/* 头部搜索区域 */
.header {
  margin-bottom: 24px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  background: var(--n-card-color);
  border-radius: 12px;
  box-shadow: 0 2px 8px rgb(0 0 0 / 6%);
  border: 1px solid var(--n-border-color);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

/* 页面设置样式 */
.page-settings-config {
  padding: 8px 0;
}

.settings-section {
  padding: 16px 0;
}

.setting-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
}

.setting-item:last-child {
  margin-bottom: 0;
}

.setting-label {
  font-size: 14px;
  font-weight: 500;
  color: var(--n-text-color);
  min-width: 80px;
}

.header-right {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}

.header-right.batch-mode {
  flex-wrap: wrap;
}

.header-right .batch-btn {
  margin-bottom: 4px;
}

.header-right .n-button:hover {
  background-color: var(--n-color-hover);
  border-radius: 6px;
}

/* 加载状态 */
.loading-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 400px;
  background: var(--n-card-color);
  border-radius: 12px;
  border: 1px solid var(--n-border-color);
  box-shadow: 0 2px 8px rgb(0 0 0 / 6%);
}

/* 存储卡片网格布局 */
.storage-cards {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 20px;
  margin-bottom: 24px;
}

/* 单个存储卡片样式 */
.storage-card {
  background: var(--n-card-color);
  border-radius: 12px;
  border: 1px solid var(--n-border-color);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  overflow: hidden;
  box-shadow: 0 2px 8px rgb(0 0 0 / 4%);
  position: relative;
}

.storage-card.is-selected {
  border: 1px solid var(--n-primary-color);
  background-color: var(--n-color-hover);
}

.selection-overlay {
  position: absolute;
  inset: 0;
  z-index: 10;
  cursor: pointer;

  /* 半透明背景，让用户知道处于选择模式 */
  background-color: rgb(0 0 0 / 2%);
}

.selection-checkbox {
  position: absolute;
  top: 12px;
  right: 12px;
  z-index: 11;
}

.storage-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 32px rgb(0 0 0 / 15%);
  border-color: var(--n-primary-color);
}

/* 卡片头部 */
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: 0;
  margin-bottom: 0;
}

.storage-info {
  flex: 1;
  min-width: 0;

  /* 确保flex子项可以收缩 */
}

.storage-title {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.storage-name {
  font-size: 18px;
  font-weight: 600;
  color: var(--n-text-color);
  line-height: 1.4;

  /* 单行显示，超出省略号 */
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;

  /* 确保为操作按钮留出空间 */
  max-width: calc(100% - 80px);
}

.storage-path {
  font-size: 13px;
  color: var(--n-text-color-2);
  line-height: 1.4;

  /* 最多两行显示，超出省略号 */
  display: -webkit-box;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
  word-break: break-all;
}

.storage-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
  align-items: flex-start;
}

.storage-actions .n-button {
  background-color: var(--n-color-target);
  border: 1px solid var(--n-border-color);
  transition: all 0.3s ease;
  width: 32px;
  height: 32px;
}

.storage-actions .n-button:hover {
  border-color: var(--n-primary-color);
  transform: scale(1.1);
}

.storage-actions .n-button:hover .n-icon {
  color: var(--n-primary-color);
}

/* 卡片内容 */
.card-content {
  padding: 0;
}

/* 基础信息容器 - 使用flex布局实现三个字段的对齐 */
.basic-info-container {
  display: flex;
  gap: 10px;
  margin-bottom: 10px;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.basic-info-container .info-item {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  min-height: 60px;
  gap: 8px;
  text-align: center;
}

.additional-info {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.info-label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  font-weight: 500;
  color: var(--n-text-color-2);
}

.info-icon {
  flex-shrink: 0;
}

.info-value {
  font-size: 14px;
  color: var(--n-text-color);
  font-weight: 500;
}

/* 时间行样式 - 一行显示，两边对齐 */
.time-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}

.time-row .info-label {
  flex-shrink: 0;
}

.time-row .info-value {
  text-align: right;
  flex-shrink: 0;
}

.info-tag {
  font-weight: 500;
}

/* 刷新信息样式 */
.refresh-info {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.refresh-detail {
  font-size: 12px;
  color: var(--n-text-color-2);
  background: var(--n-card-color);
  padding: 4px 8px;
  border-radius: 4px;
  font-weight: 500;
  border: 1px solid var(--n-border-color);
}

/* 自动刷新头部样式 */
.refresh-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}

.refresh-title-section {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
}

.refresh-title-section .info-label {
  flex-shrink: 0;
}

/* 刷新详情样式 - 复用 time-row 的 flex 布局 */
.refresh-details {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}

/* 刷新状态样式 - 左侧内容 */
.refresh-status {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
  flex-shrink: 0;
}

/* 刷新周期样式 - 右侧内容，使用灰色调避免与编辑按钮冲突 */
.refresh-period {
  font-size: 12px;
  color: var(--n-text-color);
  padding: 4px 8px;
  background: var(--n-color-hover);
  border-radius: 4px;
  border: 1px solid var(--n-border-color);
  font-weight: 500;
  flex-shrink: 0;
  text-align: right;
}

/* 任务日志样式 */
.task-log-header {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-bottom: 6px;
}

.task-log-first-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.task-log-left {
  display: flex;
  flex-direction: row;
  align-items: center;
  flex: 1;
  gap: 5px;
}

.task-log-left .info-label {
  flex-shrink: 0;
}

.task-log-time {
  font-size: 12px;
  color: var(--n-text-color-3);
}

.next-run-time {
  font-size: 12px;
  color: var(--n-text-color-3);
  margin-top: 4px;
}

.task-log-content {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 0 12px;
  border-radius: 6px;
  margin-top: 2px;
}

.task-log-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--n-text-color);
  margin-bottom: 2px;
}

.task-log-desc {
  font-size: 13px;
  color: var(--n-text-color-2);
  line-height: 1.4;
  word-break: break-all;

  /* 确保省略号正确显示 */
  display: -webkit-box;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
}

.task-log-error {
  margin-top: 6px;
  padding: 8px 10px;
  background: rgb(245 108 108 / 8%);
  border: 1px solid rgb(245 108 108 / 20%);
  border-radius: 4px;
  border-left: 3px solid #f56c6c;
}

.task-log-error .n-text {
  font-size: 12px;
  line-height: 1.4;
  font-weight: 500;
}

/* 卡片底部样式 */
.card-footer {
  padding: 0 0 16px;
}

/* 空状态 */
.empty-state {
  grid-column: 1 / -1;
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 400px;
  background: var(--n-card-color);
  border-radius: 12px;
  border: 2px dashed var(--n-border-color);
}

/* 分页容器 */
.pagination-container {
  display: flex;
  justify-content: center;
  padding: 20px;
  background: var(--n-card-color);
  border-radius: 12px;
  box-shadow: 0 2px 8px rgb(0 0 0 / 6%);
  border: 1px solid var(--n-border-color);
}

/* 响应式设计 */
@media (width <=1400px) {
  .storage-cards {
    grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
    gap: 18px;
  }
}

@media (width <=1200px) {
  .storage-cards {
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: 16px;
  }
}

@media (width <=768px) {
  .storages-page {
    padding: 16px;
  }

  .storage-cards {
    grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
    gap: 14px;
  }

  .header {
    padding: 16px;
    flex-direction: column;
    gap: 16px;
    align-items: stretch;
  }

  .header-left {
    justify-content: center;
    flex-direction: column;
    align-items: stretch;
    gap: 16px;
  }

  .header-right {
    align-self: center;
    display: flex;
    justify-content: flex-end;
    width: 100%;
  }

  .storage-name {
    max-width: calc(100% - 70px);
  }
}

@media (width <=480px) {
  .storages-page {
    padding: 12px;
  }

  .storage-cards {
    grid-template-columns: 1fr;
    gap: 12px;
  }

  .header {
    padding: 12px;
  }

  .pagination-container {
    padding: 12px;
  }

  .storage-name {
    max-width: calc(100% - 60px);
  }

  .storage-actions {
    gap: 4px;
  }

  .storage-actions .n-button {
    width: 28px;
    height: 28px;
  }
}
</style>
