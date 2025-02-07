package cli

import (
	"fmt"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"sync"
	"webproxy/model"
)

// 全局数据库连接
var (
	db   *gorm.DB
	once sync.Once
)

// 初始化数据库
func initDatabase() (*gorm.DB, error) {
	var err error
	once.Do(func() {
		db, err = gorm.Open(sqlite.Open("server.db"), &gorm.Config{})
		if err != nil {
			fmt.Println("Failed to connect to database:", err)
			return
		}
		// 自动迁移
		err = db.AutoMigrate(&model.Key{}, &model.Website{})
		if err != nil {
			fmt.Println("Failed to migrate database:", err)
			return
		}
	})
	return db, err
}
