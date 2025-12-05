package utils

import "github.com/playwright-community/playwright-go"

func ScreenshotElement(el playwright.ElementHandle, savePath string) error {
	_, err := el.Screenshot(playwright.ElementHandleScreenshotOptions{
		Path: playwright.String(savePath),
	})
	return err
}
