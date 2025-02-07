package cli

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"webproxy/model"
)

// 启动 HTTP 服务
func startHTTPServer(httpPort int) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", reverseProxyHandler)

	// 启动 HTTP 服务
	return http.ListenAndServe(fmt.Sprintf(":%d", httpPort), mux)
}

// 启动 HTTPS 服务，动态加载证书
func startHTTPSServer(httpsPort int, certs map[string]*tls.Certificate) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", reverseProxyHandler)

	// 创建 TLS 配置并指定证书
	tlsConfig := &tls.Config{
		GetCertificate: func(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
			// 根据 Host 动态选择证书
			if cert, exists := certs[clientHello.ServerName]; exists {
				return cert, nil
			}
			return nil, fmt.Errorf("certificate not found for %s", clientHello.ServerName)
		},
	}

	// 启动 HTTPS 服务
	server := &http.Server{
		Addr:      fmt.Sprintf(":%d", httpsPort),
		Handler:   mux,
		TLSConfig: tlsConfig,
	}

	// 使用证书启动 HTTPS 服务器
	return server.ListenAndServeTLS("", "") // No default cert needed because it's dynamically loaded
}

var isDebug = false

// 启动 HTTP 和 HTTPS 服务
func start(args map[string]string) error {
	var _, err = initDatabase()
	if err != nil {
		return err
	}
	// 默认端口
	httpPort := 80
	httpsPort := 443
	isDebug = args["debug"] == "true"
	// 如果 args 中包含 http 和 https，进行转换
	if httpStr, exists := args["http"]; exists {
		var err error
		httpPort, err = strconv.Atoi(httpStr)
		if err != nil {
			return fmt.Errorf("invalid http port: %v", err)
		}
	}

	if httpsStr, exists := args["https"]; exists {
		var err error
		httpsPort, err = strconv.Atoi(httpsStr)
		if err != nil {
			return fmt.Errorf("invalid https port: %v", err)
		}
	}
	// 加载所有 SSL 网站的证书
	certs, err := loadSSLCertificates()
	if err != nil {
		return fmt.Errorf("failed to load SSL certificates: %v", err)
	}
	var wg sync.WaitGroup
	// 启动 HTTP 服务
	wg.Add(1)
	go func() {
		if err := startHTTPServer(httpPort); err != nil {
			fmt.Printf("Error starting HTTP server: %v\n", err)
		}
	}()

	// 启动 HTTPS 服务
	wg.Add(1)
	go func() {
		if err := startHTTPSServer(httpsPort, certs); err != nil {
			fmt.Printf("Error starting HTTPS server: %v\n", err)
		}
	}()
	wg.Wait()
	// 返回 nil 表示启动成功
	return nil
}

// 反向代理处理，根据域名处理 SSL
func reverseProxyHandler(w http.ResponseWriter, r *http.Request) {
	if isDebug {
		fmt.Printf("isDebuging\n")
		fmt.Println(r.Host)
	}
	// 获取请求的域名（Host）
	host := r.Host
	// 查询数据库中与请求域名匹配的 Website
	var website model.Website
	result := db.Where("domain = ?", host).First(&website)
	if result.Error != nil {
		http.Error(w, "Domain not found", http.StatusNotFound)
		return
	}

	// 创建反向代理
	proxyURL := website.ProxyUrl
	proxy := http.NewServeMux()
	proxy.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		target := proxyURL + r.RequestURI
		targetURL, err := url.Parse(target)
		fmt.Println(targetURL)
		resp, err := http.Get(target)
		if err != nil {
			http.Error(w, "Proxy request failed", http.StatusInternalServerError)
			return
		}
		headers := make(map[string]string)
		for key, values := range r.Header {
			headers[key] = values[0]
		}
		// 转发请求到目标服务器
		r.URL.Scheme = targetURL.Scheme
		r.URL.Host = targetURL.Host
		r.Host = targetURL.Host

		resp, err = http.DefaultTransport.RoundTrip(r)
		if err != nil {
			http.Error(w, fmt.Sprintf("proxy error: %v", err), http.StatusBadGateway)
			return
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				fmt.Printf("Error closing body: %v", err)
			}
		}(resp.Body)

		// 将目标服务器的响应写回客户端
		for key, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	})

	// 反向代理到目标 URL
	proxy.ServeHTTP(w, r)
}

// 加载所有 SSL 网站的证书
func loadSSLCertificates() (map[string]*tls.Certificate, error) {
	// 查询所有 SSL 配置为 true 的网站
	var websites []model.Website
	result := db.Where("ssl = ?", true).Find(&websites)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to query SSL websites: %v", result.Error)
	}

	certs := make(map[string]*tls.Certificate)

	// 遍历所有的 SSL 网站，加载证书
	for _, website := range websites {
		// 读取证书和私钥文件
		certPEM, err := os.ReadFile("./website/" + website.Website + "/cert.pem")
		if err != nil {
			return nil, fmt.Errorf("failed to read cert file for domain %s: %v", website.Domain, err)
		}

		keyPEM, err := os.ReadFile("./website/" + website.Website + "/key.pem")
		if err != nil {
			return nil, fmt.Errorf("failed to read key file for domain %s: %v", website.Domain, err)
		}

		// 加载证书和私钥
		cert, err := tls.X509KeyPair(certPEM, keyPEM)
		if err != nil {
			return nil, fmt.Errorf("failed to load certificate for domain %s: %v", website.Domain, err)
		}

		// 将证书存入 map，域名作为键
		certs[website.Domain] = &cert
	}

	return certs, nil
}
