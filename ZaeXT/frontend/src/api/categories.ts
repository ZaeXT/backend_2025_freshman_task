import type { CategoryTreeNode, CreateCategoryPayload, UpdateCategoryPayload } from '@/types'

import { http } from './http'

export const fetchCategories = () => http.get<CategoryTreeNode[], CategoryTreeNode[]>('/categories')

export const createCategory = (payload: CreateCategoryPayload) =>
  http.post<CategoryTreeNode, CategoryTreeNode>('/categories', payload)

export const updateCategory = (id: number, payload: UpdateCategoryPayload) =>
  http.put<CategoryTreeNode, CategoryTreeNode>(`/categories/${id}`, payload)

export const deleteCategory = (id: number) => http.delete<void, void>(`/categories/${id}`)
