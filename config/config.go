package config

import (
	"fmt"
	"io/ioutil"
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
	// 如果文件不存在，使用默认配置
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	data, err := ioutil.ReadFile(configPath)
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
		config.Browser = "chromium"
	}
	if config.Timeout == 0 {
		config.Timeout = 5000
	}
	if config.RetryCaptcha == 0 {
		config.RetryCaptcha = 3
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

