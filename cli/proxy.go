package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

// LogEntry 定义日志结构
type LogEntry struct {
	Timestamp string            `json:"timestamp"`
	Method    string            `json:"method"`
	URL       string            `json:"url"`
	Headers   map[string]string `json:"headers"`
	Body      string            `json:"body,omitempty"`
}

func runProxy(args map[string]string) error {
	// 获取 server_port 和 proxy_port 参数
	serverPort := args["server_port"]
	if serverPort == "" {
		fmt.Println("server_port is required, server failed to start")
		return nil
	}

	proxyPort := args["proxy_port"]
	if proxyPort == "" {
		fmt.Println("proxy_port is required, proxy failed to start")
		return nil
	}

	proxyIP := args["proxy_ip"]
	if proxyIP == "" {
		proxyIP = "127.0.0.1" // 默认代理 IP
	}
	rulesFile := args["waf_rules"] // 获取 WAF 规则文件路径

	// 获取 log_mode 和 log_path 参数
	logMode := args["log_mode"]
	logPath := args["log_path"]
	if logMode == "" {
		logMode = "cli"
	}
	// 启动代理服务器
	fmt.Printf("Starting HTTP proxy server on port %s, forwarding to %s:%s...\n", serverPort, proxyIP, proxyPort)
	return server(serverPort, proxyIP, proxyPort, logMode, logPath, rulesFile)
}

func server(serverPort, proxyIP, proxyPort, logMode, logPath, rulesFile string) error {
	// 构造目标地址
	proxyURL := fmt.Sprintf("http://%s:%s", proxyIP, proxyPort)
	targetURL, err := url.Parse(proxyURL)
	if err != nil {
		return fmt.Errorf("failed to parse proxy URL: %v", err)
	}

	// 创建日志处理器
	var logWriter io.Writer
	if logMode == "cli" {
		logWriter = os.Stdout
	} else if logMode == "save" {
		if logPath == "" {
			return fmt.Errorf("log_path is required in 'save' mode")
		}
		logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file: %v", err)
		}
		defer logFile.Close()
		logWriter = logFile
	} else {
		logWriter = nil // 不记录日志
	}

	// 创建反向代理处理器
	proxy := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 如果指定了 WAF 规则文件，进行 WAF 检查
		if rulesFile != "" {
			if err := waf(r, rulesFile); err != nil {
				http.Error(w, fmt.Sprintf("WAF validation failed: %v", err), http.StatusForbidden)
				return
			}
		}

		// 记录请求头
		headers := make(map[string]string)
		for key, values := range r.Header {
			headers[key] = values[0]
		}

		// 保存请求体（用于 POST/PUT 等方法记录 data）
		var bodyData string
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
			bodyBytes, err := io.ReadAll(r.Body)
			if err == nil {
				bodyData = string(bodyBytes)
				r.Body = io.NopCloser(bytes.NewReader(bodyBytes)) // 恢复请求体供后续使用
			} else {
				bodyData = fmt.Sprintf("failed to read body: %v", err)
			}
		}

		// 构造日志条目
		logEntry := LogEntry{
			Timestamp: time.Now().Format(time.RFC3339),
			Method:    r.Method,
			URL:       r.URL.String(),
			Headers:   headers,
			Body:      bodyData,
		}

		// 输出日志
		if logWriter != nil {
			logData, err := json.Marshal(logEntry)
			if err == nil {
				logData = append(logData, '\n') // 每行一个 JSON 对象
				_, _ = logWriter.Write(logData)
			}
		}

		// 转发请求到目标服务器
		r.URL.Scheme = targetURL.Scheme
		r.URL.Host = targetURL.Host
		r.Host = targetURL.Host

		resp, err := http.DefaultTransport.RoundTrip(r)
		if err != nil {
			http.Error(w, fmt.Sprintf("proxy error: %v", err), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// 将目标服务器的响应写回客户端
		for key, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	})

	// 启动 HTTP 服务器
	serverAddr := fmt.Sprintf(":%s", serverPort)
	fmt.Printf("Proxy server is running at %s\n", serverAddr)
	return http.ListenAndServe(serverAddr, proxy)
}
