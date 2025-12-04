// +build !nocaptcha

package utils

import (
	"strings"

	"github.com/otiai10/gosseract/v2"
)

func SolveCaptcha(imagePath string) (string, error) {
	client := gosseract.NewClient()
	defer client.Close()

	client.SetWhitelist("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	client.SetImage(imagePath)
	text, err := client.Text()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(text), nil
}
