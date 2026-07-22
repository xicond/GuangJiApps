package domain

type Resource struct {
	ID        string `gorm:"column:id;primaryKey;size:50" json:"id"`
	Code      string `gorm:"column:code;size:80" json:"code"`
	Name      string `gorm:"column:name;size:200" json:"name"`
	Category  string `gorm:"column:category;size:120" json:"category"`
	Active    bool   `gorm:"column:active;default:true" json:"active"`
	CreatedBy string `gorm:"column:created_by;size:80" json:"createdBy,omitempty"`
	UpdatedBy string `gorm:"column:updated_by;size:80" json:"updatedBy,omitempty"`
	CreatedAt string `gorm:"column:created_at;size:40" json:"createdAt,omitempty"`
	UpdatedAt string `gorm:"column:updated_at;size:40" json:"updatedAt,omitempty"`
}

func (Resource) TableName() string { return "resources" }
