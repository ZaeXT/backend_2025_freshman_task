export interface CategoryTreeNode {
  id: number
  name: string
  parent_id: number | null
  created_at: string
  updated_at: string
  children: CategoryTreeNode[]
}

export interface CreateCategoryPayload {
  name: string
  parent_id?: number | null
}

export interface UpdateCategoryPayload {
  name: string
  parent_id?: number | null
}
