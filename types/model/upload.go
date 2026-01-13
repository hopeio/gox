package model

type Upload struct {
	ID           uint64 `gorm:"primary_key" json:"id"`
	Name         string `gorm:"type:varchar(100);not null" json:"name"`
	OriginalName string `gorm:"type:varchar(100);not null" json:"original_name"`
	URL          string `json:"url"`
	MD5          string `gorm:"type:varchar(32)" json:"md5"`
	Mime         string `json:"mime"`
	Size         uint64 `json:"size"`
}
