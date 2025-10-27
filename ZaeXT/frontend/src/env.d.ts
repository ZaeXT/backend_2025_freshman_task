/// <reference types="vite/client" />

declare module '*.vue'

declare interface ImportMetaEnv {
  readonly VITE_API_BASE: string
}

declare interface ImportMeta {
  readonly env: ImportMetaEnv
}
