export interface ApiResp<T> {
  code: number
  msg: string
  data: T
}

export interface PaginatedResult<T> {
  items: T[]
  total: number
  page: number
  page_size: number
}

export type ApiListResp<T> = ApiResp<T[]>

export interface ApiErrorPayload {
  code: number
  msg: string
  data?: unknown
}
