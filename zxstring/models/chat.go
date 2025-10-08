package models

// ChatRecord 聊天记录模型
type ChatRecord struct {
	ID       uint   `gorm:"primaryKey;autoIncrement" json:"id"`       // 主键ID
	Username string `gorm:"column:username;not null" json:"username"` // 用户名
	Model    string `gorm:"column:model;not null" json:"model"`       // 模型名称
	Role     string `gorm:"column:rol;not null" json:"role"`          // 角色(user/assistant)
	Time     int64  `gorm:"column:time;not null" json:"time"`         // 时间戳(毫秒)
	Content  string `gorm:"column:content;not null" json:"content"`   // 聊天内容
	Cid      string `gorm:"column:cid;not null" json:"cid"`           // 对话ID
}
