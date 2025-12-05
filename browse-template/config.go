package browseTemplate

import (
	"autotest/browse-template/utils"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// Config 配置结构
type Config struct {
	Browser           string `yaml:"browser"`             // chromium, firefox, webkit
	Headless          bool   `yaml:"headless"`            // 是否无头模式
	Timeout           int    `yaml:"timeout"`             // 超时时间（毫秒）
	RetryCaptcha      int    `yaml:"retry_captcha"`       // 验证码重试次数
	IgnoreHTTPSErrors bool   `yaml:"ignore_https_errors"` // 是否忽略 HTTPS 证书错误（仅测试环境建议开启）
	KeepBrowserOpen   bool   `yaml:"keep_browser_open"`   // 测试结束后是否保留浏览器（仅调试时建议开启）
}

// LoadConfig 从文件加载配置
func LoadConfig(configPath string) (*Config, error) {
	defaultConfig := DefaultConfig()
	// 如果文件不存在，使用默认配置
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return defaultConfig, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 如果某些字段未设置，使用默认值
	if config.Browser == "" {
		config.Browser = defaultConfig.Browser
	}
	if config.Timeout == 0 {
		config.Timeout = defaultConfig.Timeout
	}
	if config.RetryCaptcha == 0 {
		config.RetryCaptcha = defaultConfig.RetryCaptcha
	}
	// 默认不忽略 HTTPS 错误，除非配置中显式开启
	// 这里不强制设置，保持配置文件的布尔值即可

	return &config, nil
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Browser:           "chromium",
		Headless:          false,
		Timeout:           5000,
		RetryCaptcha:      3,
		IgnoreHTTPSErrors: false,
		KeepBrowserOpen:   false,
	}
}

// ExpectConfig 期望验证配置
type ExpectConfig struct {
	Type  string `json:"type"`           // "text", "xpath", "css", "id"
	Value string `json:"value"`          // 选择器的值
	Mode  string `json:"mode"`           // "value_equals", "text_equals", "text_contains", "visible"
	Text  string `json:"text,omitempty"` // 期望的文本内容
}

// TestStep 测试步骤
type TestStep struct {
	Action    string                 `json:"action"`              // "goto", "input", "click", "assert", "menu_click", "captcha_input", "select_option", "select_options", "checkbox_toggle", "checkbox_set", "checkboxes_set", "radio_select", "radios_select", "table_edit", "table_delete", "table_assert", "search"
	URL       string                 `json:"url,omitempty"`       // goto的URL
	Selector  *utils.SelectorConfig  `json:"selector,omitempty"`  // 元素选择器（单个）
	Selectors []utils.SelectorConfig `json:"selectors,omitempty"` // 元素选择器（多个，用于批量操作）
	Text      string                 `json:"text,omitempty"`      // input的文本内容，或select的选项值（单个）
	Options   []string               `json:"options,omitempty"`   // select的选项值（多个，用于多选）
	Expect    *ExpectConfig          `json:"expect,omitempty"`    // 期望验证配置
	MenuPath  string                 `json:"menu_path,omitempty"` // 菜单路径，格式: "系统管理 > 用户管理 > 新增用户"
	Captcha   *CaptchaConfig         `json:"captcha,omitempty"`   // 验证码配置
	Checked   *bool                  `json:"checked,omitempty"`   // checkbox_set时使用，true表示选中，false表示取消选中
	Table     *TableConfig           `json:"table,omitempty"`     // 表格配置
	Search    *SearchConfig          `json:"search,omitempty"`    // 查询配置
}

// TableConfig 表格配置
type TableConfig struct {
	Selector utils.SelectorConfig `json:"selector"`         // 表格选择器
	Row      *TableRowConfig      `json:"row,omitempty"`    // 行定位配置
	Column   *TableColumnConfig   `json:"column,omitempty"` // 列定位配置
	Action   string               `json:"action,omitempty"` // 操作类型: "edit", "delete"
	Value    string               `json:"value,omitempty"`  // 断言期望值
	Mode     string               `json:"mode,omitempty"`   // 断言模式: "equals", "contains", "not_equals", "not_contains"
}

// TableRowConfig 表格行配置
type TableRowConfig struct {
	Type  string `json:"type"`  // "index"（索引）, "text"（文本匹配）, "contains"（包含文本）
	Value string `json:"value"` // 行定位值
}

// TableColumnConfig 表格列配置
type TableColumnConfig struct {
	Type  string `json:"type"`  // "index"（索引）, "header"（表头文本）
	Value string `json:"value"` // 列定位值
}

// SearchConfig 查询配置
type SearchConfig struct {
	Inputs []SearchInput         `json:"inputs,omitempty"` // 查询输入框配置
	Button *utils.SelectorConfig `json:"button,omitempty"` // 查询按钮选择器
}

// SearchInput 查询输入配置
type SearchInput struct {
	Selector *utils.SelectorConfig `json:"selector"` // 输入框选择器
	Text     string                `json:"text"`     // 输入文本
}

// CaptchaConfig 验证码配置
type CaptchaConfig struct {
	ImageSelector *utils.SelectorConfig `json:"image_selector,omitempty"` // 验证码图片选择器
	InputSelector *utils.SelectorConfig `json:"input_selector,omitempty"` // 验证码输入框选择器
	Auto          bool                  `json:"auto,omitempty"`           // 是否自动识别（自动查找验证码图片和输入框）
}
