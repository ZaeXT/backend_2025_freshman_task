package request

type CreateCategory struct {
	Name     string `json:"name" binding:"required,max=100"`
	ParentID *uint  `json:"parent_id"`
}

type UpdateCategory struct {
	Name     string `json:"name" binding:"required,max=100"`
	ParentID *uint  `json:"parent_id"`
}
