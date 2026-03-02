import { defineStore } from 'pinia'
import { reactive, watch } from 'vue'
import { localStg } from '@/utils/storage'
import { type StorageType } from '@/types/global.d'

export const useSharedStore = defineStore('shared', () => {
  const load = (): StorageType.StorageSetting => {
    const _storageSetting = localStg.get('storageSetting')

    return _storageSetting || { pathPrefix: '/', selectedToken: 0 }
  }

  const store = (data: StorageType.StorageSetting) => {
    localStg.set('storageSetting', data)
    Object.assign(storageSetting, data)
  }

  const storageSetting = reactive<StorageType.StorageSetting>(load())

  // 重置路径前缀（不保存到localStorage，用于订阅号挂载）
  const resetPathPrefix = (prefix: string) => {
    storageSetting.pathPrefix = prefix
  }

  // 监听变化保存到localStorage
  watch(
    () => storageSetting,
    (state) => {
      store(state)
    },
    { deep: true }
  )

  return { storageSetting, resetPathPrefix }
})
