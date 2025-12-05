package apistemplate

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestLoadAPITemplates(t *testing.T) {
	// 测试 LoadAPITemplates 函数
	templates, err := LoadAPITemplates("apis.json")
	if err != nil {
		t.Fatalf("LoadAPITemplates 失败: %v", err)
	}

	// 验证解析结果
	if len(templates) != 3 {
		t.Fatalf("期望解析出 3 个模板, 实际解析出 %d 个", len(templates))
	}

	callLogin, exists := templates["call_login"]
	if !exists {
		t.Fatalf("未找到模板 'call_login'")
	}
	fmt.Printf("call_login 模板: %+v\n", callLogin)

	if callLogin.Method != "post" {
		t.Errorf("期望 call_login 的 Method 为 post, 实际为 %s", callLogin.Method)
	}
}

// 定义用于解析测试用例 JSON 的辅助结构体
// 注意：这些结构体只在测试中使用，所以定义在 _test.go 文件里即可
type TestCaseConfig struct {
	Template string  `json:"template"`
	Params   []Param `json:"params"`
}

type TestCase struct {
	Name      string         `json:"name"`
	APIConfig TestCaseConfig `json:"api_config"`
}

func TestGenerateRequest_ParamReplacement(t *testing.T) {
	// 1. 准备数据：模拟加载进来的模板 (模拟 apis.json)
	// 这里我们模拟 "call_login" 模板
	templates := APITemplates{
		"call_login": APIRequest{
			URL:    "http://{api_host}:{api_port}/api/login",
			Method: "post",
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Data: map[string]interface{}{
				"username": "{username}", // 待替换
				"password": "{password}", // 待替换
			},
		},
	}

	// 2. 准备数据：你的测试用例 JSON
	jsonInput := `
	{
		"name": "Test Login API",
		"api_config": {
			"template": "call_login",
			"params": [{
				"key": "{username}",
				"value": "123"
			},{
				"key": "{password}",
				"value": "localhost"
			}]
		}
	}`

	// 3. 解析测试用例
	var tc TestCase
	if err := json.Unmarshal([]byte(jsonInput), &tc); err != nil {
		t.Fatalf("测试用例 JSON 解析失败: %v", err)
	}

	// 4. 获取模板
	tmpl, exists := templates[tc.APIConfig.Template]
	if !exists {
		t.Fatalf("模板 '%s' 不存在", tc.APIConfig.Template)
	}

	// 5. 执行核心逻辑：生成请求
	// 注意：你的 JSON 只提供了 {username}，没有提供 {password} 或 {api_host}
	// 所以预期结果是：username 被替换，其他保持原样
	gotReq, err := GenerateRequest(tmpl, tc.APIConfig.Params)
	if err != nil {
		t.Fatalf("GenerateRequest 执行失败: %v", err)
	}

	// 6. 验证结果 (断言)

	// 6.1 验证 Data 是否存在
	if gotReq.Data == nil {
		t.Fatal("生成的请求 Data 为空")
	}

	// 6.2 直接使用 gotReq.Data
	// 因为在结构体定义中，Data 已经是 map[string]interface{} 类型，无需断言
	dataMap := gotReq.Data

	if dataMap == nil {
		t.Fatal("生成的请求 Data 为空")
	}

	// 6.3 核心验证：检查 username 是否变成了 "123"
	expectedUsername := "123"
	if dataMap["username"] != expectedUsername {
		t.Errorf("替换失败: username 期望是 '%s', 实际是 '%v'", expectedUsername, dataMap["username"])
	}

	t.Logf("测试通过！生成的 Body: %v", dataMap)
}

func TestGenerateRequest_URLReplacement(t *testing.T) {
	// 1. 准备模板
	templates := APITemplates{
		"call_login": APIRequest{
			// 原始 URL 包含两个占位符
			URL:    "http://{api_host}:{api_port}/api/login",
			Method: "post",
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Data: map[string]interface{}{
				"username": "{username}", // 待替换
				"password": "{password}", // 待替换
			},
		},
	}

	// 2. 准备测试数据 (这次我们提供 host 和 port)
	jsonInput := `
	{
		"name": "Test URL Replace",
		"api_config": {
			"template": "call_login",
			"params": [
				{ "key": "{api_host}", "value": "192.168.1.100" },
				{ "key": "{api_port}", "value": "8080" }
			]
		}
	}`

	// 3. 解析测试配置
	var tc TestCase
	if err := json.Unmarshal([]byte(jsonInput), &tc); err != nil {
		t.Fatalf("JSON解析失败: %v", err)
	}

	// 4. 获取模板
	tmpl := templates[tc.APIConfig.Template]

	// 5. 执行生成逻辑
	gotReq, err := GenerateRequest(tmpl, tc.APIConfig.Params)
	if err != nil {
		t.Fatalf("GenerateRequest 出错: %v", err)
	}

	// 6. 验证 URL 替换结果
	expectedURL := "http://192.168.1.100:8080/api/login"

	if gotReq.URL != expectedURL {
		t.Errorf("URL 替换错误.\n期望: %s\n实际: %s", expectedURL, gotReq.URL)
	} else {
		t.Logf("URL 替换成功: %s", gotReq.URL)
	}
}
