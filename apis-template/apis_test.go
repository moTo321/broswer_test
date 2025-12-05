package apisTemplate

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

func TestGenerateRequest_ParamReplacement(t *testing.T) {
	// 1. 准备数据：模拟加载进来的模板
	templates := APITemplates{
		"call_login": APIRequest{
			URL:    "http://{api_host}:{api_port}/api/login",
			Method: "post",
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Data: map[string]interface{}{
				"username": "{username}",
				"password": "{password}",
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
	gotReq, err := GenerateRequest(tmpl, tc.APIConfig.Params)
	if err != nil {
		t.Fatalf("GenerateRequest 执行失败: %v", err)
	}

	// 6. 验证结果
	if gotReq.Data == nil {
		t.Fatal("生成的请求 Data 为空")
	}

	dataMap := gotReq.Data
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
			URL:    "http://{api_host}:{api_port}/api/login",
			Method: "post",
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Data: map[string]interface{}{
				"username": "{username}",
				"password": "{password}",
			},
		},
	}

	// 2. 准备测试数据
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

	var tc TestCase
	if err := json.Unmarshal([]byte(jsonInput), &tc); err != nil {
		t.Fatalf("JSON解析失败: %v", err)
	}

	tmpl := templates[tc.APIConfig.Template]

	gotReq, err := GenerateRequest(tmpl, tc.APIConfig.Params)
	if err != nil {
		t.Fatalf("GenerateRequest 出错: %v", err)
	}

	expectedURL := "http://192.168.1.100:8080/api/login"

	if gotReq.URL != expectedURL {
		t.Errorf("URL 替换错误.\n期望: %s\n实际: %s", expectedURL, gotReq.URL)
	} else {
		t.Logf("URL 替换成功: %s", gotReq.URL)
	}
}

func TestRunTestCases_RealServer(t *testing.T) {
	// 1. 加载 API 模板 (从 apis.json 文件)
	// 确保 apis.json 在当前测试目录下，或者使用绝对路径
	templates, err := LoadAPITemplates("apis.json")
	if err != nil {
		t.Fatalf("加载 apis.json 模板失败: %v", err)
	}

	// 2. 您的测试用例数据 (JSON Array)
	jsonInput := `
	[
		{
			"name": "Test Login API",
			"api_config": {
				"template": "call_login",
				"params": [
					{ "key": "{username}", "value": "123" },
					{ "key": "{password}", "value": "123" },
					{ "key": "{api_host}", "value": "192.168.0.108" },
					{ "key": "{api_port}", "value": "8080" }
				]
			},
			"expect": {
				"status": 200,
				"body": {
					"code": 1001,
					"message": "用户名或密码错误"
				}
			}
		}
	]`

	// 3. 解析测试用例
	var testCases []TestCase
	if err := json.Unmarshal([]byte(jsonInput), &testCases); err != nil {
		t.Fatalf("JSON 解析失败: %v", err)
	}

	// 4. 循环执行测试
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			// A. 获取模板
			tmpl, exists := templates[tc.APIConfig.Template]
			if !exists {
				t.Fatalf("模板 '%s' 不存在", tc.APIConfig.Template)
			}

			// B. 生成请求 (替换 host, port, username 等参数)
			req, err := GenerateRequest(tmpl, tc.APIConfig.Params)
			if err != nil {
				t.Fatalf("请求生成失败: %v", err)
			}

			// 打印一下最终 URL 方便调试
			t.Logf("正在请求 URL: %s", req.URL)

			// C. 执行请求 (发送到真实服务器)
			resp, err := ExecuteRequest(req)
			if err != nil {
				t.Fatalf("请求执行失败: %v", err)
			}
			t.Logf("收到响应\n Status: %d,\n RawBody: %s,\n Header: %s,\n Body: %v\n", resp.StatusCode, resp.RawBody, resp.Header, resp.Body)

			// ==========================================
			// D. 验证阶段
			// ==========================================

			// D-1: 校验 HTTP 状态码 (Status)
			if resp.StatusCode != tc.Expect.Status {
				t.Errorf("HTTP状态码错误! 期望: %d, 实际: %d", tc.Expect.Status, resp.StatusCode)
			}

			// D-2: 校验 Body 内容 (Code, Message 等)
			for key, expectedVal := range tc.Expect.Body {
				actualVal, exists := resp.Body[key]
				if !exists {
					t.Errorf("响应 Body 缺少字段: %s", key)
					continue
				}

				// 使用辅助函数比较 (解决 1001 vs 1001.0 问题)
				if !compareValues(expectedVal, actualVal) {
					t.Errorf("字段 '%s' 校验失败.\n期望: %v (类型 %T)\n实际: %v (类型 %T)",
						key, expectedVal, expectedVal, actualVal, actualVal)
				}
			}
		})
	}
}
