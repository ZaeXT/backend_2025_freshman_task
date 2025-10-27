import type { ModelInfo } from '@/types'

import { http } from './http'

export const fetchModels = () => http.get<ModelInfo[], ModelInfo[]>('/models')
