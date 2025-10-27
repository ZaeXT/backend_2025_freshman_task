import type { RecycleBinItem } from '@/types'

import { http } from './http'

export const fetchRecycleBinItems = () =>
  http.get<RecycleBinItem[], RecycleBinItem[]>('/recycle-bin')

export const restoreConversation = (id: number) =>
  http.post<void, void>(`/recycle-bin/restore/${id}`)

export const deleteConversationPermanently = (id: number) =>
  http.delete<void, void>(`/recycle-bin/permanent/${id}`)
