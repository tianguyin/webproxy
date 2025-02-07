package model

type Key struct {
	ID  uint   `gorm:"primaryKey"`
	Key string `gorm:"size:255"`
}

type Website struct {
	ID       uint   `gorm:"primaryKey"`
	Website  string `gorm:"size:255"`
	Domain   string `gorm:"size:100"`
	ProxyUrl string `gorm:"size:100"`
	SSL      bool   `gorm:"default:false"`
}
