package utils

import (
	"fmt"
	"strings"

	"github.com/playwright-community/playwright-go"
)

// SelectorConfig 选择器配置
type SelectorConfig struct {
	Type  string `json:"type"`            // "text", "xpath", "css", "id" 等
	Value string `json:"value"`           // 选择器的值
	Scope string `json:"scope,omitempty"` // 作用域: "", "dialog"（当前弹窗内查找）, 未来可扩展 "new_window" 等
}

// LocateElement 基于选择器配置定位元素
// 支持多种定位方式，优先使用文本定位
func LocateElement(page playwright.Page, selector SelectorConfig) (playwright.ElementHandle, error) {
	switch selector.Type {
	case "text":
		return locateByText(page, selector.Value, selector.Scope)
	case "field":
		return locateField(page, selector.Value, selector.Scope)
	case "button":
		return locateButton(page, selector.Value, selector.Scope)
	case "xpath":
		return page.Locator(selector.Value).First().ElementHandle()
	case "css":
		return page.Locator(selector.Value).First().ElementHandle()
	case "id":
		return page.Locator("#" + selector.Value).First().ElementHandle()
	default:
		// 默认尝试文本定位
		return locateByText(page, selector.Value, selector.Scope)
	}
}

// locateField 基于文本内容定位“字段输入控件”
// 测试人员只需要写字段文字，例如: {type: "field", value: "策略名称"}
func locateField(page playwright.Page, text string, scope string) (playwright.ElementHandle, error) {
	// 根据 scope 决定查找范围（与 locateByText 一致）
	root := page.Locator("body")
	switch scope {
	case "dialog":
		dialogSelectors := []string{
			"[role='dialog']",
			".el-dialog__wrapper", // Element UI 外层
			".el-dialog",          // Element UI 内容层
			".ant-modal-root",     // Ant Design
			".ant-modal-content",
			".modal",
			".dialog",
		}

		foundVisibleRoot := false
		for _, ds := range dialogSelectors {
			loc := page.Locator(ds)
			count, err := loc.Count()
			if err == nil && count > 0 {
				// 遍历所有找到的 dialog 节点，只选可见的那个
				for i := range count {
					candidate := loc.Nth(i)
					// 检查该容器本身是否可见，或者其内部是否包含可见内容
					if visible, _ := candidate.IsVisible(); visible {
						root = candidate
						foundVisibleRoot = true
						break
					}
				}
			}
			if foundVisibleRoot {
				break
			}
		}
		// 如果指定了 scope 是 dialog 但没找到可见的 dialog，考虑回退到 body
		if !foundVisibleRoot {
			fmt.Println("[locateField] 警告: 未找到可见的 Dialog 容器，将尝试在全页面搜索...")
			root = page.Locator("body")
		}
	case "main":
		mainSelectors := []string{
			"main",
			".el-main",
			".el-container .el-main",
			".ant-layout-content",
			".layout-main",
			"#app main",
			"#app .main",
			"#app .content",
		}
		for _, ms := range mainSelectors {
			loc := page.Locator(ms)
			count, err := loc.Count()
			if err == nil && count > 0 {
				root = loc
				break
			}
		}
	}

	selectors := []string{
		// placeholder 方式
		fmt.Sprintf("input[placeholder*='%s']", text),
		fmt.Sprintf("input[placeholder='%s']", text),
		// label + input 经典布局
		fmt.Sprintf("label:has-text('%s') + input", text),
		fmt.Sprintf("label:has-text('%s') ~ input", text),
		fmt.Sprintf("//label[contains(text(), '%s')]/following-sibling::input[1]", text),
		fmt.Sprintf("//label[contains(text(), '%s')]/../input", text),
		// 通用：任意包含该文字的元素，后面紧跟的 input/textarea/contenteditable
		fmt.Sprintf("//*[contains(normalize-space(text()),'%s')]/following::input[1]", text),
		fmt.Sprintf("//*[contains(normalize-space(text()),'%s')]/following::textarea[1]", text),
		fmt.Sprintf("//*[contains(normalize-space(text()),'%s')]/following::*[@contenteditable='true'][1]", text),
	}

	var lastErr error
	for _, s := range selectors {
		locator := root.Locator(s)
		count, err := locator.Count()
		if err == nil && count > 0 {
			element, err := locator.First().ElementHandle()
			if err == nil && element != nil {
				return element, nil
			}
		}
		if err != nil {
			lastErr = err
		}
	}

	// 兜底策略（主要给弹窗使用）：如果根据字段文字找不到，
	// 且 scope 在 dialog 内，则返回弹窗中第一个可见的输入控件
	if scope == "dialog" {
		fallback := root.Locator("input, textarea, [contenteditable='true']")
		count, err := fallback.Count()
		if err == nil && count > 0 {
			for i := 0; i < count; i++ {
				h := fallback.Nth(i)
				visible, _ := h.IsVisible()
				if visible {
					element, err := h.ElementHandle()
					if err == nil && element != nil {
						return element, nil
					}
				}
			}
		}
	}

	return nil, fmt.Errorf("无法通过字段文本 '%s' 定位输入控件: %v", text, lastErr)
}

// locateButton 基于文本内容定位“可点击按钮”
// 测试人员只需要写按钮文字，例如: {type: "button", value: "新建"}
func locateButton(page playwright.Page, text string, scope string) (playwright.ElementHandle, error) {
	// 根据 scope 决定查找范围（与 locateByText 一致）
	root := page.Locator("body")
	switch scope {
	case "dialog":
		dialogSelectors := []string{
			"[role='dialog']",
			".el-dialog__wrapper", // Element UI 外层
			".el-dialog",          // Element UI 内容层
			".ant-modal-root",     // Ant Design
			".ant-modal-content",
			".modal",
			".dialog",
		}

		foundVisibleRoot := false
		for _, ds := range dialogSelectors {
			loc := page.Locator(ds)
			count, err := loc.Count()
			if err == nil && count > 0 {
				// 遍历所有找到的 dialog 节点，只选可见的那个
				for i := range count {
					candidate := loc.Nth(i)
					// 检查该容器本身是否可见，或者其内部是否包含可见内容
					if visible, _ := candidate.IsVisible(); visible {
						root = candidate
						foundVisibleRoot = true
						break
					}
				}
			}
			if foundVisibleRoot {
				break
			}
		}
		// 如果指定了 scope 是 dialog 但没找到可见的 dialog，考虑回退到 body
		if !foundVisibleRoot {
			fmt.Println("[locateButton] 警告: 未找到可见的 Dialog 容器，将尝试在全页面搜索...")
			root = page.Locator("body")
		}
	case "main":
		mainSelectors := []string{
			"main",
			".el-main",
			".el-container .el-main",
			".ant-layout-content",
			".layout-main",
			"#app main",
			"#app .main",
			"#app .content",
		}
		for _, ms := range mainSelectors {
			loc := page.Locator(ms)
			count, err := loc.Count()
			if err == nil && count > 0 {
				root = loc
				break
			}
		}
	}

	selectors := []string{
		// 原生 button
		fmt.Sprintf("button:has-text('%s')", text),
		// 常见 UI 框架按钮
		fmt.Sprintf(".el-button:has-text('%s')", text),
		fmt.Sprintf(".ant-btn:has-text('%s')", text),
		// 通过 role/button 属性
		fmt.Sprintf("[role='button']:has-text('%s')", text),
		// 文本在内部 span/div，向上找父级 button
		fmt.Sprintf("//button[.//*[normalize-space(text())='%s']]", text),
		fmt.Sprintf("//button[.//*[contains(normalize-space(text()),'%s')]]", text),
	}

	var lastErr error
	for _, s := range selectors {
		locator := root.Locator(s)
		count, err := locator.Count()
		if err == nil && count > 0 {
			// 优先返回可见的
			for i := 0; i < count; i++ {
				handle := locator.Nth(i)
				visible, _ := handle.IsVisible()
				if visible {
					element, err := handle.ElementHandle()
					if err == nil && element != nil {
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

	// 如果找不到真正的 button，回退到通用文本匹配
	// 找到包含文本的元素（可能是 <p>、<div>、<span>、<a> 等）
	// 很多前端框架使用这些标签作为按钮，直接返回找到的元素即可
	fallbackSelectors := []string{
		// 通用文本匹配
		fmt.Sprintf("text=%s", text),
		fmt.Sprintf("text=/^%s$/", escapeRegex(text)),
		fmt.Sprintf("text=/.*%s.*/", escapeRegex(text)),
		// 常见标签类型
		fmt.Sprintf("p:has-text('%s')", text),
		fmt.Sprintf("div:has-text('%s')", text),
		fmt.Sprintf("span:has-text('%s')", text),
		fmt.Sprintf("a:has-text('%s')", text),
		// 通过常见按钮相关的 class 或 id（如 signInWarp、login 等）
		fmt.Sprintf("[class*='signIn']:has-text('%s')", text),
		fmt.Sprintf("[class*='login']:has-text('%s')", text),
		fmt.Sprintf("[class*='button']:has-text('%s')", text),
		fmt.Sprintf("[class*='btn']:has-text('%s')", text),
		fmt.Sprintf("[id*='login']:has-text('%s')", text),
		fmt.Sprintf("[id*='signIn']:has-text('%s')", text),
	}

	for _, s := range fallbackSelectors {
		locator := root.Locator(s)
		count, err := locator.Count()
		if err == nil && count > 0 {
			// 优先返回可见的元素
			for i := 0; i < count; i++ {
				handle := locator.Nth(i)
				visible, _ := handle.IsVisible()
				if visible {
					element, err := handle.ElementHandle()
					if err == nil && element != nil {
						// 直接返回找到的元素（很多前端框架用 <p>、<div> 等作为按钮）
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

	return nil, fmt.Errorf("无法通过按钮文本 '%s' 定位按钮: %v", text, lastErr)
}

// locateByText 通过文本内容定位元素
// 尝试多种策略：
// 1. 直接文本匹配（按钮、链接等）
// 2. placeholder属性（输入框）
// 3. label标签关联
// 4. aria-label属性
// 5. title属性
// scope: "", "dialog" 等，默认为整页查找
func locateByText(page playwright.Page, text string, scope string) (playwright.ElementHandle, error) {
	// 为了让测试人员只写“可见文本”就能更稳定地定位到真正的输入框，
	// 这里优先尝试 placeholder / label 关联到 input 的策略，
	// 然后才回退到通用的 text= 文本匹配。

	// 根据 scope 决定查找范围
	// 默认在 body 下查找；
	// - scope == "dialog" : 在常见弹窗容器内查找
	// - scope == "main"   : 在主内容区域查找（排除左侧菜单等）
	root := page.Locator("body")
	switch scope {
	case "dialog":
		dialogSelectors := []string{
			"[role='dialog']",
			".el-dialog__wrapper",
			".el-dialog",
			".ant-modal-root",
			".ant-modal",
			".modal",
			".dialog",
		}
		for _, ds := range dialogSelectors {
			loc := page.Locator(ds)
			count, err := loc.Count()
			if err == nil && count > 0 {
				root = loc
				break
			}
		}
	case "main":
		mainSelectors := []string{
			"main",
			".el-main",
			".el-container .el-main",
			".ant-layout-content",
			".layout-main",
			"#app main",
			"#app .main",
			"#app .content",
		}
		for _, ms := range mainSelectors {
			loc := page.Locator(ms)
			count, err := loc.Count()
			if err == nil && count > 0 {
				root = loc
				break
			}
		}
	}

	selectors := []string{}

	// 策略1: 通过placeholder定位输入框
	selectors = append(selectors,
		fmt.Sprintf("input[placeholder*='%s']", text),
		fmt.Sprintf("input[placeholder='%s']", text),
	)

	// 策略2: 通过label定位（查找label文本，然后找关联的input、select、textarea等）
	selectors = append(selectors,
		fmt.Sprintf("label:has-text('%s') + input", text),
		fmt.Sprintf("label:has-text('%s') ~ input", text),
		fmt.Sprintf("label:has-text('%s') + select", text),
		fmt.Sprintf("label:has-text('%s') ~ select", text),
		fmt.Sprintf("label:has-text('%s') + textarea", text),
		fmt.Sprintf("label:has-text('%s') ~ textarea", text),
		fmt.Sprintf("//label[contains(text(), '%s')]/following-sibling::input[1]", text),
		fmt.Sprintf("//label[contains(text(), '%s')]/following-sibling::select[1]", text),
		fmt.Sprintf("//label[contains(text(), '%s')]/../input", text),
		fmt.Sprintf("//label[contains(text(), '%s')]/../select", text),
		fmt.Sprintf("//label[contains(text(), '%s')]/../textarea", text),
	)

	// 策略2.5: 通过label定位checkbox和radio
	selectors = append(selectors,
		fmt.Sprintf("label:has-text('%s') input[type='checkbox']", text),
		fmt.Sprintf("label:has-text('%s') input[type='radio']", text),
		fmt.Sprintf("//label[contains(text(), '%s')]//input[@type='checkbox']", text),
		fmt.Sprintf("//label[contains(text(), '%s')]//input[@type='radio']", text),
	)

	// 策略3: 通过aria-label定位
	selectors = append(selectors,
		fmt.Sprintf("[aria-label*='%s']", text),
		fmt.Sprintf("[aria-label='%s']", text),
	)

	// 策略4: 通过title属性定位
	selectors = append(selectors,
		fmt.Sprintf("[title*='%s']", text),
		fmt.Sprintf("[title='%s']", text),
	)

	// 策略5: 通过data-*属性定位（常见的前端框架）
	selectors = append(selectors,
		fmt.Sprintf("[data-label*='%s']", text),
		fmt.Sprintf("[data-placeholder*='%s']", text),
	)

	// 策略6: 定位select下拉框（通过选项文本定位select元素）
	selectors = append(selectors,
		fmt.Sprintf("select option:has-text('%s')", text),
		fmt.Sprintf("//select[.//option[contains(text(), '%s')]]", text),
	)

	// 策略7: 定位checkbox和radio（通过value或文本）
	selectors = append(selectors,
		fmt.Sprintf("input[type='checkbox'][value*='%s']", text),
		fmt.Sprintf("input[type='radio'][value*='%s']", text),
		fmt.Sprintf("input[type='checkbox'] + label:has-text('%s')", text),
		fmt.Sprintf("input[type='radio'] + label:has-text('%s')", text),
	)

	// 策略8: 按照“按钮优先”的方式，通过 text 定位可点击按钮
	selectors = append(selectors,
		fmt.Sprintf("button:has-text('%s')", text),
		fmt.Sprintf(".el-button:has-text('%s')", text),
		fmt.Sprintf(".ant-btn:has-text('%s')", text),
	)

	// 策略9: 最后才使用通用的文本匹配（按钮、链接、span等）
	selectors = append(selectors,
		fmt.Sprintf("text=%s", text),                    // 精确匹配
		fmt.Sprintf("text=/^%s$/", escapeRegex(text)),   // 正则精确匹配
		fmt.Sprintf("text=/.*%s.*/", escapeRegex(text)), // 正则包含匹配
	)

	// 尝试每个选择器
	var lastErr error
	for _, selector := range selectors {
		locator := root.Locator(selector)
		count, err := locator.Count()
		if err == nil && count > 0 {
			// 找到元素，返回第一个
			element, err := locator.First().ElementHandle()
			if err == nil && element != nil {
				return element, nil
			}
		}
		if err != nil {
			lastErr = err
		}
	}

	// 如果所有策略都失败，尝试更通用的方法
	// 查找包含该文本的所有元素，然后筛选可见和可交互的
	allLocators, err := root.Locator(fmt.Sprintf("text=/.*%s.*/", escapeRegex(text))).All()
	if err == nil && len(allLocators) > 0 {
		// 只返回可见的元素；如果都不可见，则视为未找到
		for _, locator := range allLocators {
			visible, _ := locator.IsVisible()
			if visible {
				element, err := locator.ElementHandle()
				if err == nil && element != nil {
					return element, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("无法通过文本 '%s' 定位元素: %v", text, lastErr)
}

// escapeRegex 转义正则表达式特殊字符
func escapeRegex(text string) string {
	specialChars := []string{"\\", "^", "$", ".", "|", "?", "*", "+", "(", ")", "[", "]", "{", "}"}
	result := text
	for _, char := range specialChars {
		result = strings.ReplaceAll(result, char, "\\"+char)
	}
	return result
}

// GetElementText 获取元素的文本内容
func GetElementText(element playwright.ElementHandle) (string, error) {
	return element.TextContent()
}

// GetElementValue 获取输入框的值
func GetElementValue(element playwright.ElementHandle) (string, error) {
	return element.InputValue()
}

// IsElementVisible 检查元素是否可见
func IsElementVisible(element playwright.ElementHandle) (bool, error) {
	return element.IsVisible()
}
