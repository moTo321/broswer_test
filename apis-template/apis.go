package apisTemplate

import (
	"encoding/json"
	"fmt"
	"os"
)

// APIRequest 定义单个 API 请求的结构
type APIRequest struct {
	URL          string                 `json:"url"`                     // API 请求的 URL
	Method       string                 `json:"method"`                  // HTTP 方法 (GET, POST, etc.)
	Data         map[string]interface{} `json:"data,omitempty"`          // 请求体数据
	Headers      map[string]string      `json:"headers,omitempty"`       // 请求头
	SaveResponse map[string]string      `json:"save_response,omitempty"` // 保存响应中的字段
}

// APITemplates 定义所有 API 请求模板的集合
type APITemplates map[string]APIRequest

// LoadAPITemplates 从指定的 JSON 文件加载所有 API 请求模板
func LoadAPITemplates(filePath string) (APITemplates, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("没有API模板文件")
	}

	// 读取文件内容
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取API模板文件失败: %v", err)
	}

	// 解析 JSON 数据
	var templates APITemplates
	if err := json.Unmarshal(data, &templates); err != nil {
		return nil, fmt.Errorf("解析API模板文件失败: %v", err)
	}

	return templates, nil
}

// 定义用于解析测试用例 JSON 的辅助结构体
type TestCaseConfig struct {
	Template string  `json:"template"`
	Params   []Param `json:"params"`
}

// TestCase 对应 JSON Array 中的单个对象
type TestCase struct {
	Name      string         `json:"name"`
	APIConfig TestCaseConfig `json:"api_config"`
	Expect    ExpectConfig   `json:"expect"`
}

// ExpectConfig 定义期望结果
type ExpectConfig struct {
	Status int                    `json:"status"`
	Body   map[string]interface{} `json:"body"`
}

// ValidateResponse 验证响应是否符合期望
func ValidateResponse(resp *APIResponse, expect ExpectConfig) error {
	// 1. 校验状态码
	if resp.StatusCode != expect.Status {
		return fmt.Errorf("HTTP状态码不匹配: 期望 %d, 实际 %d", expect.Status, resp.StatusCode)
	}

	// 2. 校验 Body
	for key, expectedVal := range expect.Body {
		actualVal, exists := resp.Body[key]
		if !exists {
			return fmt.Errorf("响应Body缺少字段: %s", key)
		}

		if !compareValues(expectedVal, actualVal) {
			return fmt.Errorf("字段 '%s' 校验失败: 期望 %v, 实际 %v", key, expectedVal, actualVal)
		}
	}
	return nil
}

// compareValues 内部辅助函数 (处理 int vs float64)
func compareValues(expected, actual interface{}) bool {
	if expected == actual {
		return true
	}
	expFloat, ok1 := toFloat64(expected)
	actFloat, ok2 := toFloat64(actual)
	if ok1 && ok2 {
		return expFloat == actFloat
	}
	// 兜底：转字符串比较
	return fmt.Sprintf("%v", expected) == fmt.Sprintf("%v", actual)
}

func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case float64:
		return val, true
	case float32:
		return float64(val), true
	default:
		return 0, false
	}
}
