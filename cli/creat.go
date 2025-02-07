package cli

import (
	"fmt"
	"webproxy/model"
)

func creatWebsite(args map[string]string) error {
	website := args["website"]
	domain := args["domain"]
	proxyUrl := args["proxyUrl"]
	ssl := args["ssl"] == "true"

	return insertDB(website, domain, proxyUrl, ssl)
}

func insertDB(website, domain, proxyUrl string, ssl bool) error {
	// 初始化数据库
	db, err := initDatabase()
	if err != nil {
		return err
	}

	// 创建一个新的 Website 实例
	newWebsite := model.Website{
		Website:  website,
		Domain:   domain,
		ProxyUrl: proxyUrl,
		SSL:      ssl,
	}

	// 插入到数据库
	if err := db.Create(&newWebsite).Error; err != nil {
		return fmt.Errorf("插入失败: %v", err)
	}
	fmt.Println("成功创建一个新的网站：" + website + "\n" + "域名" + domain + "\n" + "后端地址：" + proxyUrl)
	return nil
}
