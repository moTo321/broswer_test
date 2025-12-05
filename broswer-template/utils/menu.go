package utils

import (
	"fmt"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

// ClickMenu 点击多级菜单
// menuPath 格式: "系统管理 > 用户管理 > 新增用户"
func ClickMenu(page playwright.Page, menuPath string) error {
	// 分割菜单路径
	menuItems := strings.Split(menuPath, ">")
	if len(menuItems) == 0 {
		return fmt.Errorf("菜单路径不能为空")
	}

	// 清理每个菜单项的空白字符
	for i := range menuItems {
		menuItems[i] = strings.TrimSpace(menuItems[i])
		if menuItems[i] == "" {
			return fmt.Errorf("菜单路径包含空项")
		}
	}

	// 逐级点击菜单
	for i, menuText := range menuItems {
		fmt.Printf("    点击菜单项 [%d/%d]: %s\n", i+1, len(menuItems), menuText)

		// 定位菜单项
		element, err := locateMenuElement(page, menuText)
		if err != nil {
			return fmt.Errorf("定位菜单项 '%s' 失败: %v", menuText, err)
		}

		// 等待元素可见
		visible, err := element.IsVisible()
		if err != nil || !visible {
			// 尝试悬停以展开菜单
			err = element.Hover()
			if err != nil {
				return fmt.Errorf("悬停菜单项 '%s' 失败: %v", menuText, err)
			}
			time.Sleep(200 * time.Millisecond)
		}

		// 点击菜单项
		err = element.Click()
		if err != nil {
			return fmt.Errorf("点击菜单项 '%s' 失败: %v", menuText, err)
		}

		// 等待菜单展开/页面响应
		time.Sleep(300 * time.Millisecond)

		// 如果不是最后一项，等待子菜单出现
		if i < len(menuItems)-1 {
			time.Sleep(200 * time.Millisecond)
		}
	}

	return nil
}

// locateMenuElement 定位菜单元素
// 尝试多种策略来定位菜单项
func locateMenuElement(page playwright.Page, menuText string) (playwright.ElementHandle, error) {
	// 策略1: 直接文本匹配（最常见的菜单项）
	selectors := []string{
		fmt.Sprintf("text=%s", menuText),
		fmt.Sprintf("text=/^%s$/", escapeRegex(menuText)),
		fmt.Sprintf("text=/.*%s.*/", escapeRegex(menuText)),
	}

	// 策略2: 通过常见的菜单元素定位
	// li, a, span, div 等可能包含菜单文本的元素
	selectors = append(selectors,
		fmt.Sprintf("li:has-text('%s')", menuText),
		fmt.Sprintf("a:has-text('%s')", menuText),
		fmt.Sprintf("span:has-text('%s')", menuText),
		fmt.Sprintf("div:has-text('%s')", menuText),
		fmt.Sprintf("[role='menuitem']:has-text('%s')", menuText),
		fmt.Sprintf("[role='button']:has-text('%s')", menuText),
	)

	// 策略3: 通过aria-label定位
	selectors = append(selectors,
		fmt.Sprintf("[aria-label*='%s']", menuText),
		fmt.Sprintf("[aria-label='%s']", menuText),
	)

	// 策略4: 通过title属性定位
	selectors = append(selectors,
		fmt.Sprintf("[title*='%s']", menuText),
		fmt.Sprintf("[title='%s']", menuText),
	)

	// 策略5: 通过data-*属性定位
	selectors = append(selectors,
		fmt.Sprintf("[data-menu*='%s']", menuText),
		fmt.Sprintf("[data-title*='%s']", menuText),
	)

	// 尝试每个选择器
	var lastErr error
	for _, selector := range selectors {
		locator := page.Locator(selector)
		count, err := locator.Count()
		if err == nil && count > 0 {
			// 找到元素，返回第一个可见的
			for i := 0; i < count; i++ {
				elem := locator.Nth(i)
				element, err := elem.ElementHandle()
				if err == nil && element != nil {
					visible, _ := element.IsVisible()
					if visible {
						return element, nil
					}
				}
			}
			// 如果没有可见的，返回第一个
			element, err := locator.First().ElementHandle()
			if err == nil && element != nil {
				return element, nil
			}
		}
		if err != nil {
			lastErr = err
		}
	}

	return nil, fmt.Errorf("无法定位菜单项 '%s': %v", menuText, lastErr)
}
