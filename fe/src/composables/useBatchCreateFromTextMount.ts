import { h } from 'vue'
import { useModal } from 'naive-ui'
import { BatchCreateFromTextModal } from '@/components/storage'
import { useMountPointBind } from '@/composables/useMountPointBind'

export function useBatchCreateFromTextMount() {
    const modal = useModal()
    const mountPointBind = useMountPointBind()

    const show = (): Promise<{ success: boolean }> => {
        return new Promise((resolve) => {
            const modalInstance = modal.create({
                title: '批量文本导入',
                preset: 'dialog',
                style: { width: '700px' },
                content: () =>
                    h(BatchCreateFromTextModal, {
                        onParsed: async (payload: { items: any[], token: number }) => {
                            modalInstance.destroy()

                            const mountItems = payload.items.map(item => ({
                                name: item.name,
                                osType: item.osType,
                                subscribeUser: item.subscribeUser,
                                shareCode: item.shareCode,
                                shareAccessCode: item.shareAccessCode,
                                fileId: item.fileId,
                                cloudToken: payload.token,
                                disableSwitchCloudToken: false
                            }))

                            const result = await mountPointBind.show(mountItems, { defaultCloudToken: payload.token })

                            if (result && result.length > 0) {
                                resolve({ success: true })
                            } else {
                                resolve({ success: false })
                            }
                        },
                        onCancel: () => {
                            resolve({ success: false })
                            modalInstance.destroy()
                        },
                    }),
                closable: true,
                maskClosable: false,
            })
        })
    }

    return { show }
}
