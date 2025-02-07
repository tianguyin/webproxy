package cli

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"webproxy/model"
)

// 启动 Web 服务
func runWeb(args map[string]string) error {
	fmt.Println("Starting webserver...")

	// 初始化数据库
	_, err := initDatabase()
	if err != nil {
		return err
	}

	// 解析 webView 参数
	webView := args["webView"] == "true"
	webPort := 7788
	if portStr, exists := args["webPort"]; exists {
		if port, err := strconv.Atoi(portStr); err == nil {
			webPort = port
		} else {
			return fmt.Errorf("invalid webPort: %v", err)
		}
	}

	// 启动 Web 服务器
	return webserver(webView, webPort)
}

func webserver(webView bool, webPort int) error {
	if webView {
		// 创建新的 HTTP 多路复用器
		mux := http.NewServeMux()

		// 注册所有 API 路由
		defineAPIRoutes(mux)

		// 启动 API 服务器
		return http.ListenAndServe(fmt.Sprintf(":%d", webPort), mux)
	}
	return nil
}

// APIRoute API 路由定义结构体
type APIRoute struct {
	Path         string
	HandlerFunc  http.HandlerFunc
	AuthRequired bool
}

// 定义所有的 API 路由
func defineAPIRoutes(mux *http.ServeMux) {
	// API 路由列表
	routes := []APIRoute{
		{
			Path:         "/api/creatWebsite",
			HandlerFunc:  creatWebsiteHandler,
			AuthRequired: true,
		},
		// 添加更多路由
		// {
		//     Path:        "/api/other",
		//     HandlerFunc: otherHandler,
		//     AuthRequired: true,
		// },
	}

	// 遍历路由列表并注册
	for _, route := range routes {
		// 根据是否需要认证来决定是否使用 Bearer Token 验证中间件
		if route.AuthRequired {
			mux.HandleFunc(route.Path, withBearerAuth(route.HandlerFunc))
		} else {
			mux.HandleFunc(route.Path, route.HandlerFunc)
		}
	}
}

// Bearer Token 验证中间件
func withBearerAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 验证 Bearer Token
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Authorization header is missing Bearer token", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if !isValidToken(token) {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// 调用下一个处理程序
		next(w, r)
	}
}

// 验证 Bearer Token 是否有效
func isValidToken(token string) bool {
	var key model.Key
	// 查询数据库中的 key 是否匹配
	result := db.Where("key = ?", token).First(&key)
	return result.Error == nil
}

// 处理受保护的 API 路由
func creatWebsiteHandler(w http.ResponseWriter, r *http.Request) {
	website := r.FormValue("website")
	domain := r.FormValue("domain")
	proxyUrl := r.FormValue("proxyUrl")
	ssl := r.FormValue("ssl") == "true" // 如果 ssl 字段值为 "true" 则设为 true
	i := insertDB(website, domain, proxyUrl, ssl)
	if i != nil {
		return
	}
	// 处理请求
	w.WriteHeader(http.StatusCreated)
	_, err := fmt.Fprintf(w, "Authenticated request")
	if err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

// 示例：处理其他 API 路由
func otherHandler(w http.ResponseWriter, r *http.Request) {
	// 处理请求
	w.WriteHeader(http.StatusOK)
	_, err := fmt.Fprintf(w, "This is another API route")
	if err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}
