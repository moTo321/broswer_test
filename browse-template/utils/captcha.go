package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/playwright-community/playwright-go"
)

// SolveAndInputCaptcha 识别验证码并输入
// 返回识别出的验证码文本
func SolveAndInputCaptcha(page playwright.Page, captchaSelector, inputSelector SelectorConfig) (string, error) {
	// 1. 定位验证码图片
	captchaElement, err := LocateElement(page, captchaSelector)
	if err != nil {
		return "", fmt.Errorf("定位验证码图片失败: %v", err)
	}

	// 2. 确保目录存在
	captchaDir := filepath.Join("assets", "captcha")
	err = os.MkdirAll(captchaDir, 0755)
	if err != nil {
		return "", fmt.Errorf("创建验证码目录失败: %v", err)
	}

	// 3. 截图验证码
	timestamp := time.Now().Unix()
	captchaImagePath := filepath.Join(captchaDir, fmt.Sprintf("captcha_%d.png", timestamp))
	_, err = captchaElement.Screenshot(playwright.ElementHandleScreenshotOptions{
		Path: playwright.String(captchaImagePath),
	})
	if err != nil {
		return "", fmt.Errorf("验证码截图失败: %v", err)
	}

	// 4. OCR识别验证码
	captchaText, err := SolveCaptcha(captchaImagePath)
	if err != nil {
		return "", fmt.Errorf("验证码识别失败: %v", err)
	}

	if captchaText == "" {
		return "", fmt.Errorf("验证码识别结果为空")
	}

	fmt.Printf("    识别验证码: %s\n", captchaText)

	// 5. 定位输入框并输入
	inputElement, err := LocateElement(page, inputSelector)
	if err != nil {
		return "", fmt.Errorf("定位验证码输入框失败: %v", err)
	}

	// 清空并输入验证码
	err = inputElement.Fill(captchaText)
	if err != nil {
		return "", fmt.Errorf("输入验证码失败: %v", err)
	}

	// 等待一下，确保输入完成
	time.Sleep(200 * time.Millisecond)

	return captchaText, nil
}

// AutoSolveCaptcha 自动识别并输入验证码（使用默认选择器）
// 自动查找常见的验证码图片和输入框
func AutoSolveCaptcha(page playwright.Page) (string, error) {
	// 尝试常见的验证码图片选择器
	captchaSelectors := []SelectorConfig{
		{Type: "css", Value: "img[src*='captcha']"},
		{Type: "css", Value: "img[src*='verify']"},
		{Type: "css", Value: "img[alt*='验证码']"},
		{Type: "css", Value: "img[alt*='captcha']"},
		{Type: "css", Value: ".captcha img"},
		{Type: "css", Value: ".verify-code img"},
		{Type: "xpath", Value: "//img[contains(@src, 'captcha')]"},
		{Type: "xpath", Value: "//img[contains(@src, 'verify')]"},
	}

	// 尝试常见的验证码输入框选择器
	inputSelectors := []SelectorConfig{
		{Type: "text", Value: "验证码"},
		{Type: "text", Value: "请输入验证码"},
		{Type: "css", Value: "input[name*='captcha']"},
		{Type: "css", Value: "input[name*='verify']"},
		{Type: "css", Value: "input[placeholder*='验证码']"},
		{Type: "css", Value: "input[placeholder*='captcha']"},
		{Type: "xpath", Value: "//input[contains(@placeholder, '验证码')]"},
		{Type: "xpath", Value: "//input[contains(@name, 'captcha')]"},
	}

	// 尝试找到验证码图片
	var captchaSelector SelectorConfig
	var captchaFound bool
	for _, selector := range captchaSelectors {
		element, err := LocateElement(page, selector)
		if err == nil && element != nil {
			visible, _ := element.IsVisible()
			if visible {
				captchaSelector = selector
				captchaFound = true
				break
			}
		}
	}

	if !captchaFound {
		return "", fmt.Errorf("未找到验证码图片")
	}

	// 尝试找到输入框
	var inputSelector SelectorConfig
	var inputFound bool
	for _, selector := range inputSelectors {
		element, err := LocateElement(page, selector)
		if err == nil && element != nil {
			visible, _ := element.IsVisible()
			if visible {
				inputSelector = selector
				inputFound = true
				break
			}
		}
	}

	if !inputFound {
		return "", fmt.Errorf("未找到验证码输入框")
	}

	// 识别并输入验证码
	return SolveAndInputCaptcha(page, captchaSelector, inputSelector)
}
