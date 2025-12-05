//go:build nocaptcha
// +build nocaptcha

package utils

import (
	"fmt"
)

// SolveCaptcha 验证码识别（未启用 OCR 功能）
// 当使用 nocaptcha build tag 编译时，此函数会返回错误
func SolveCaptcha(imagePath string) (string, error) {
	return "", fmt.Errorf("验证码识别功能未启用: 请安装 Tesseract OCR 或使用正常编译方式")
}
