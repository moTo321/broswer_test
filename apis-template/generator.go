package apistemplate

import (
	"encoding/json"
	"strings"
)

// Param 定义用于替换的键值对
// 例如: {Key: "{username}", Value: "admin"}
type Param struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// GenerateRequest 根据 API 模板和参数列表，生成最终用于发送的 APIRequest 对象
func GenerateRequest(template APIRequest, params []Param) (APIRequest, error) {
	// 1. 深拷贝 (Deep Copy)
	var newReq APIRequest
	bytes, err := json.Marshal(template)
	if err != nil {
		return newReq, err
	}
	if err := json.Unmarshal(bytes, &newReq); err != nil {
		return newReq, err
	}

	// 2. 替换 URL 中的占位符
	newReq.URL = replaceString(newReq.URL, params)

	// 3. 替换 Data (Body) 中的占位符
	// Data 是 map[string]interface{}，可能包含嵌套结构，需要递归处理
	if newReq.Data != nil {
		processedData := resolveData(newReq.Data, params)
		// 断言回 map[string]interface{}
		if m, ok := processedData.(map[string]interface{}); ok {
			newReq.Data = m
		}
	}

	return newReq, nil
}

// =======================================================
// 私有辅助函数 (Helper Functions)
// =======================================================

// replaceString 简单的字符串替换
func replaceString(str string, params []Param) string {
	if str == "" {
		return ""
	}
	result := str
	for _, p := range params {
		// 将所有出现的 Key (如 "{username}") 替换为 Value
		result = strings.ReplaceAll(result, p.Key, p.Value)
	}
	return result
}

// resolveData 递归遍历任意类型的数据，查找并替换字符串中的占位符
func resolveData(input interface{}, params []Param) interface{} {
	switch v := input.(type) {
	case string:
		// 如果是字符串，直接尝试替换
		return replaceString(v, params)

	case map[string]interface{}:
		// 如果是 Map (JSON Object)，递归处理每一个 Value
		newMap := make(map[string]interface{})
		for k, val := range v {
			newMap[k] = resolveData(val, params)
		}
		return newMap

	case []interface{}:
		// 如果是 Slice (JSON Array)，递归处理每一个 Element
		newSlice := make([]interface{}, len(v))
		for i, val := range v {
			newSlice[i] = resolveData(val, params)
		}
		return newSlice

	default:
		// 其他类型（如 float64, bool, nil）保持原样
		return v
	}
}
