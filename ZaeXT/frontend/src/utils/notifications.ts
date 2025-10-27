import mitt from 'mitt'

export type ToastLevel = 'info' | 'success' | 'warning' | 'error'

export interface ToastPayload {
  id?: string
  type?: ToastLevel
  title?: string
  message: string
  duration?: number
}

type NotificationEvents = {
  toast: ToastPayload
}

const emitter = mitt<NotificationEvents>()

export const emitToast = (payload: ToastPayload) => {
  emitter.emit('toast', {
    type: 'info',
    duration: 3000,
    ...payload,
  })
}

export const onToast = (handler: (payload: ToastPayload) => void) => {
  emitter.on('toast', handler)
  return () => emitter.off('toast', handler)
}

export const notificationBus = emitter
