package broswerTemplate

import (
	"log"
	"time"

	"github.com/playwright-community/playwright-go"
)

var browser playwright.Browser
var context playwright.BrowserContext

// Start 启动浏览器（使用默认配置）
func Start() playwright.Page {
	return StartWithConfig(nil)
}

// StartWithConfig 使用指定配置启动浏览器
func StartWithConfig(cfg *Config) playwright.Page {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("启动 Playwright 失败: %v\n请确认已安装依赖环境，例如:\n  - 安装 Node.js\n  - 安装 Playwright 浏览器: npx playwright install", err)
	}

	// 根据配置选择浏览器
	var browserType playwright.BrowserType
	switch cfg.Browser {
	case "firefox":
		browserType = pw.Firefox
	case "webkit":
		browserType = pw.WebKit
	case "chromium":
		browserType = pw.Chromium
	default:
		// 未知浏览器类型，使用默认值 Chromium
		browserType = pw.Chromium
	}

	// 启动浏览器
	launchOpts := playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(cfg.Headless),
	}
	// 根据配置选择是否忽略 HTTPS 证书错误（通过启动参数实现）
	if cfg.IgnoreHTTPSErrors {
		launchOpts.Args = []string{"--ignore-certificate-errors"}
	}

	browser, err = browserType.Launch(launchOpts)
	if err != nil {
		log.Fatalf("启动浏览器失败 (%s): %v", cfg.Browser, err)
	}

	// 创建浏览器上下文
	videoDir := "assets/videos"
	context, err = browser.NewContext(playwright.BrowserNewContextOptions{
		RecordVideo: &playwright.RecordVideo{
			Dir: videoDir,
		},
	})
	if err != nil {
		log.Fatalf("创建浏览器上下文失败: %v", err)
	}

	// 创建页面
	page, err := context.NewPage()
	if err != nil {
		log.Fatalf("创建页面失败: %v", err)
	}

	// 设置默认超时时间
	page.SetDefaultTimeout(float64(cfg.Timeout))

	return page
}

func TakeErrorScreenshot(page playwright.Page) {
	timeStr := time.Now().Format("2006-01-02_15-04-05.000")
	file := "assets/errors/error_" + timeStr + ".png"
	page.Screenshot(playwright.PageScreenshotOptions{
		Path: playwright.String(file),
	})
}

func Stop() {
	context.Close()
	browser.Close()
}
