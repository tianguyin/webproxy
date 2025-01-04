package cli

import (
	"bytes"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
	"os"
	"regexp"
)

// WAFRules 定义 WAF 规则结构
type WAFRules struct {
	Low  WAFRuleSet `yaml:"low"`
	High WAFRuleSet `yaml:"high"`
}

// WAFRuleSet 定义单个规则集
type WAFRuleSet struct {
	Allow    WAFRule `yaml:"allow"`
	Disallow WAFRule `yaml:"disallow"`
}

// WAFRule 定义规则内容
type WAFRule struct {
	Agent []string `yaml:"agent"`
	Body  []string `yaml:"body"`
	URL   []string `yaml:"url"`
}

// waf 函数加载规则并检查请求
func waf(r *http.Request, rulesFile string) error {
	// 加载规则文件
	data, err := os.ReadFile(rulesFile)
	if err != nil {
		return fmt.Errorf("failed to read rules file: %v", err)
	}

	// 解析 YAML 文件
	var wafRules WAFRules
	if err := yaml.Unmarshal(data, &wafRules); err != nil {
		return fmt.Errorf("failed to parse rules file: %v", err)
	}

	// 1. 检查 URL
	urlPath := r.URL.Path
	if err := checkRuleWithHigh(urlPath, wafRules.Low.Allow.URL, wafRules.Low.Disallow.URL, wafRules.High.Allow.URL); err != nil {
		return err
	}

	// 2. 检查 User-Agent
	userAgent := r.Header.Get("User-Agent")
	if err := checkRuleWithHigh(userAgent, wafRules.Low.Allow.Agent, wafRules.Low.Disallow.Agent, wafRules.High.Allow.Agent); err != nil {
		return err
	}

	// 3. 检查 Body 内容
	var bodyData string
	if r.Body != nil {
		bodyBytes, _ := io.ReadAll(r.Body)
		bodyData = string(bodyBytes)
		r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}
	if err := checkRuleWithHigh(bodyData, wafRules.Low.Allow.Body, wafRules.Low.Disallow.Body, wafRules.High.Allow.Body); err != nil {
		return err
	}

	return nil
}

// checkRuleWithHigh 检查规则，并考虑 high 规则的覆盖
func checkRuleWithHigh(content string, allowRules []string, disallowRules []string, highAllowRules []string) error {
	// 如果禁止规则匹配，直接返回错误
	for _, rule := range disallowRules {
		if match, _ := regexp.MatchString(rule, content); match {
			// 如果 low 规则拒绝了，但 high 规则允许某些内容，继续检查 high 规则
			for _, highAllowRule := range highAllowRules {
				if match, _ := regexp.MatchString(highAllowRule, content); match {
					// 如果 high 规则允许更大的词语，放行请求
					return nil
				}
			}
			return fmt.Errorf("content matches disallowed rule: %s", rule)
		}
	}

	// 如果允许规则不为空且未匹配，返回错误
	if len(allowRules) > 0 {
		matched := false
		for _, rule := range allowRules {
			if match, _ := regexp.MatchString(rule, content); match {
				matched = true
				break
			}
		}
		if !matched {
			return fmt.Errorf("content does not match any allowed rules")
		}
	}

	return nil
}
