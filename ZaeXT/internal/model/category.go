package model

type Category struct {
	BaseModel
	UserID   uint   `gorm:"not null;index"`
	Name     string `gorm:"size:100;not null"`
	ParentID *uint  `gorm:"index"`

	User     User        `gorm:"foreignKey:UserID"`
	Children []*Category `gorm:"foreignKey:ParentID;constraint:OnDelete:CASCADE"`
}
