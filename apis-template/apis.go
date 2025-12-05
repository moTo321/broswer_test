package apistemplate

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
