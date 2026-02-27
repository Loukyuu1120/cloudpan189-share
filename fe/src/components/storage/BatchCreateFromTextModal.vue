<template>
  <div class="batch-text-container">
    <div class="batch-text-content">
      <div class="header-section">
        <n-alert type="info" show-icon title="使用说明" class="mb-4">
          <template #default>
            <div class="usage-guide">
              <p>输入分享链接、分享码、文件夹ID或订阅号链接，系统将自动识别名称。</p>
              <p>支持格式：</p>
              <p>• 分享链接：https://cloud.189.cn/t/xxxxx</p>
              <p>• 文件夹ID：123456789</p>
              <p>• 订阅号链接：https://content.21cn.com/h5/subscrip/... ?uuid=xxx</p>
              <p>点击"下一步"后，您可以批量设置挂载路径前缀。</p>
            </div>
          </template>
        </n-alert>
      </div>

      <n-form
          ref="formRef"
          :model="formModel"
          :rules="rules"
          label-placement="left"
          label-width="100px"
      >
        <!-- 云盘账号选择 -->
        <n-form-item label="云盘账号" path="cloudToken">
          <n-select
              v-model:value="formModel.cloudToken"
              :options="cloudTokenOptions"
              :loading="state.loadingTokens"
              placeholder="请选择用于解析的账号"
              clearable
          />
        </n-form-item>

        <!-- 文本内容 -->
        <n-form-item label="资源列表" path="content">
          <n-input
              v-model:value="formModel.content"
              type="textarea"
              placeholder="示例：&#10;https://cloud.189.cn/t/code&#10;123456789 (文件夹ID)"
              :autosize="{ minRows: 8, maxRows: 15 }"
              class="resource-textarea"
          />
        </n-form-item>
      </n-form>
    </div>

    <!-- 底部操作栏 -->
    <div class="modal-actions">
      <n-button @click="handleCancel">取消</n-button>
      <n-button
          type="primary"
          :loading="state.submitting"
          @click="handleNext"
      >
        下一步：解析并设置路径
      </n-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, computed, onMounted } from 'vue'
import {
  NForm, NFormItem, NInput, NSelect, NButton, NAlert, useMessage,
  type FormRules
} from 'naive-ui'
import { getCloudTokenList } from '@/api/cloudtoken'
import { batchParseStorageText, type BatchParseItem } from '@/api/storage'

interface Emits {
  (e: 'parsed', payload: { items: BatchParseItem[], token: number }): void
  (e: 'cancel'): void
}
const emit = defineEmits<Emits>()

const message = useMessage()
const formRef = ref(null)

const state = reactive({
  loadingTokens: false,
  submitting: false,
  cloudTokens: [] as any[]
})

const formModel = reactive({
  cloudToken: null as number | null,
  content: ''
})

// 修改点：显式指定类型 : FormRules
const rules: FormRules = {
  cloudToken: [{
    type: 'number',
    required: true,
    message: '请选择云盘账号',
    trigger: ['blur', 'change']
  }],
  content: [{
    required: true,
    message: '请输入内容',
    trigger: 'blur'
  }]
}

const cloudTokenOptions = computed(() => {
  return state.cloudTokens.map((token) => ({
    label: token.name,
    value: token.id
  }))
})

const fetchCloudTokens = async () => {
  state.loadingTokens = true
  try {
    const res = await getCloudTokenList({ noPaginate: true })
    if (res.code === 200 || res.code === 0) {
      const rawData = res.data
      let list: any[] = []
      if (Array.isArray(rawData)) list = rawData
      else if (rawData && typeof rawData === 'object') list = (rawData as any).data || []
      state.cloudTokens = list
      if (state.cloudTokens.length === 1) formModel.cloudToken = state.cloudTokens[0].id
    }
  } finally {
    state.loadingTokens = false
  }
}

const handleNext = () => {
  (formRef.value as any)?.validate((errors: any) => {
    if (errors || !formModel.cloudToken) return

    state.submitting = true

    batchParseStorageText({
      content: formModel.content,
      cloudToken: formModel.cloudToken
    })
        .then(res => {
          if (res.code === 200 && res.data && res.data.length > 0) {
            message.success(`成功解析 ${res.data.length} 个资源`)
            emit('parsed', {
              items: res.data,
              token: formModel.cloudToken as number
            })
          } else {
            message.warning('未能解析出有效资源，请检查格式或账号权限')
          }
        })
        .catch(err => {
          message.error(err.message || '解析失败')
        })
        .finally(() => {
          state.submitting = false
        })
  })
}

const handleCancel = () => emit('cancel')

onMounted(() => fetchCloudTokens())
</script>

<style scoped>
.batch-text-container { width: 100%; }
.batch-text-content { padding: 0 4px; }
.header-section { margin-bottom: 20px; }
.usage-guide { font-size: 13px; line-height: 1.6; }
.resource-textarea { font-family: monospace; }
.modal-actions { display: flex; gap: 12px; justify-content: flex-end; padding-top: 20px; margin-top: 10px; border-top: 1px solid var(--n-border-color); }
</style>
