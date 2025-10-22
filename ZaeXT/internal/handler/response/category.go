package response

type CategoryInfo struct {
	ID       uint            `json:"id"`
	Name     string          `json:"name"`
	ParentID *uint           `json:"parent_id,omitempty"`
	Children []*CategoryInfo `json:"children,omitempty"`
}
