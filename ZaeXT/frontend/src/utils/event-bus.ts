import mitt from 'mitt'

export type AppEventMap = {
  'auth:unauthorized': void
}

export const appEventBus = mitt<AppEventMap>()
