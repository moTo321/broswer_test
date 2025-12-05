package apistemplate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// APIResponse 封装执行后的结果
type APIResponse struct {
	StatusCode int                    // HTTP 状态码
	Body       map[string]interface{} // 解析后的 JSON 响应体
	RawBody    string                 // 原始响应文本 (用于调试)
	Header     http.Header            // 响应头
}

// ExecuteRequest 发送 HTTP 请求
func ExecuteRequest(req APIRequest) (*APIResponse, error) {
	// 1. 准备请求体 (Body)
	var bodyReader io.Reader
	if len(req.Data) > 0 {
		jsonBytes, err := json.Marshal(req.Data)
		if err != nil {
			return nil, fmt.Errorf("请求体序列化失败: %v", err)
		}
		bodyReader = bytes.NewBuffer(jsonBytes)
	}

	// 2. 创建 HTTP Request 对象
	httpReq, err := http.NewRequest(strings.ToUpper(req.Method), req.URL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 3. 设置请求头 (Headers)
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	// 4. 发起网络请求
	client := &http.Client{
		Timeout: 10 * time.Second, // 设置超时防止卡死
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("网络请求发送失败: %v", err)
	}
	defer resp.Body.Close()

	// 5. 读取响应体
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 6. 封装结果
	result := &APIResponse{
		StatusCode: resp.StatusCode,
		Header:     resp.Header,
		RawBody:    string(respBytes),
		Body:       make(map[string]interface{}),
	}

	// 尝试将响应解析为 JSON Map
	if len(respBytes) > 0 {
		_ = json.Unmarshal(respBytes, &result.Body)
	}

	return result, nil
}
