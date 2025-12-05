package utils

import (
	"fmt"
	"time"

	"github.com/playwright-community/playwright-go"
)

// SelectOption 选择下拉框选项（单选）
// 支持通过文本、值选择
func SelectOption(page playwright.Page, selectSelector SelectorConfig, optionValue string) error {
	return SelectOptions(page, selectSelector, []string{optionValue})
}

// SelectOptions 选择下拉框选项（支持多选）
// 支持通过文本、值选择多个选项
func SelectOptions(page playwright.Page, selectSelector SelectorConfig, optionValues []string) error {
	if len(optionValues) == 0 {
		return fmt.Errorf("选项列表不能为空")
	}
	// 构建select元素的定位器
	var selectLocator playwright.Locator

	switch selectSelector.Type {
	case "text":
		// 通过文本定位select，尝试多种方式
		selectors := []string{
			fmt.Sprintf("label:has-text('%s') + select", selectSelector.Value),
			fmt.Sprintf("label:has-text('%s') ~ select", selectSelector.Value),
			fmt.Sprintf("//label[contains(text(), '%s')]/following-sibling::select[1]", selectSelector.Value),
			fmt.Sprintf("//label[contains(text(), '%s')]/../select", selectSelector.Value),
			fmt.Sprintf("select:has(option:has-text('%s'))", selectSelector.Value),
		}

		var found bool
		for _, sel := range selectors {
			loc := page.Locator(sel)
			count, err := loc.Count()
			if err == nil && count > 0 {
				selectLocator = loc
				found = true
				break
			}
		}

		if !found {
			// 如果找不到，尝试通用文本定位
			_, err := LocateElement(page, selectSelector)
			if err != nil {
				return fmt.Errorf("定位下拉框失败: %v", err)
			}
			// 获取元素的定位器（通过xpath）
			selectLocator = page.Locator(fmt.Sprintf("//select[.//option[contains(text(), '%s')]]", selectSelector.Value))
		}
	case "xpath", "css", "id":
		selectLocator = page.Locator(selectSelector.Value)
	default:
		_, err := LocateElement(page, selectSelector)
		if err != nil {
			return fmt.Errorf("定位下拉框失败: %v", err)
		}
		// 对于其他类型，直接使用选择器值
		selectLocator = page.Locator(selectSelector.Value)
	}

	// 尝试通过label（文本）选择
	_, err := selectLocator.SelectOption(playwright.SelectOptionValues{Labels: &optionValues})
	if err != nil {
		// 如果失败，尝试通过value选择
		_, err = selectLocator.SelectOption(playwright.SelectOptionValues{Values: &optionValues})
		if err != nil {
			// 如果还是失败，可能是自定义下拉框，尝试点击展开并逐个选择
			selectElement, err2 := selectLocator.First().ElementHandle()
			if err2 == nil && selectElement != nil {
				// 检查是否是multiple select
				isMultipleResult, _ := selectElement.Evaluate("el => el.multiple || false")
				isMultiple := false
				if boolResult, ok := isMultipleResult.(bool); ok {
					isMultiple = boolResult
				}

				// 对于多选，需要先点击展开（如果还没展开）
				optionsToSelect := optionValues
				if !isMultiple && len(optionValues) > 1 {
					// 如果不是multiple，只能选第一个
					optionsToSelect = []string{optionValues[0]}
				}

				err2 = selectElement.Click()
				if err2 == nil {
					time.Sleep(300 * time.Millisecond)

					// 逐个选择选项
					for _, optionValue := range optionsToSelect {
						optionSelectors := []string{
							fmt.Sprintf("text=%s", optionValue),
							fmt.Sprintf("//option[contains(text(), '%s')]", optionValue),
							fmt.Sprintf("//li[contains(text(), '%s')]", optionValue),
							fmt.Sprintf("[role='option']:has-text('%s')", optionValue),
						}

						var optionFound bool
						for _, sel := range optionSelectors {
							optionLocator := page.Locator(sel)
							count, err3 := optionLocator.Count()
							if err3 == nil && count > 0 {
								optionElement, err4 := optionLocator.First().ElementHandle()
								if err4 == nil && optionElement != nil {
									// 对于多选，可能需要按住Ctrl键
									if isMultiple {
										// 使用键盘按下 Ctrl 键
										page.Keyboard().Down("Control")
										err3 = optionElement.Click()
										page.Keyboard().Up("Control")
									} else {
										err3 = optionElement.Click()
									}
									if err3 == nil {
										optionFound = true
										time.Sleep(100 * time.Millisecond)
										break
									}
								}
							}
						}

						if !optionFound {
							return fmt.Errorf("未找到选项: %s", optionValue)
						}
					}

					time.Sleep(200 * time.Millisecond)
					return nil
				}
			}
			return fmt.Errorf("选择选项失败: %v", err)
		}
	}

	time.Sleep(200 * time.Millisecond)
	return nil
}

// ToggleCheckbox 切换复选框状态
// 如果未选中则选中，如果已选中则取消选中
func ToggleCheckbox(page playwright.Page, checkboxSelector SelectorConfig) error {
	checkboxElement, err := LocateElement(page, checkboxSelector)
	if err != nil {
		return fmt.Errorf("定位复选框失败: %v", err)
	}

	// 检查当前状态
	checked, err := checkboxElement.IsChecked()
	if err != nil {
		return fmt.Errorf("检查复选框状态失败: %v", err)
	}

	// 切换状态
	if checked {
		err = checkboxElement.Uncheck()
	} else {
		err = checkboxElement.Check()
	}

	if err != nil {
		return fmt.Errorf("切换复选框状态失败: %v", err)
	}

	time.Sleep(200 * time.Millisecond)
	return nil
}

// SetCheckbox 设置复选框状态
// checked: true表示选中，false表示取消选中
func SetCheckbox(page playwright.Page, checkboxSelector SelectorConfig, checked bool) error {
	checkboxElement, err := LocateElement(page, checkboxSelector)
	if err != nil {
		return fmt.Errorf("定位复选框失败: %v", err)
	}

	if checked {
		err = checkboxElement.Check()
	} else {
		err = checkboxElement.Uncheck()
	}

	if err != nil {
		return fmt.Errorf("设置复选框状态失败: %v", err)
	}

	time.Sleep(200 * time.Millisecond)
	return nil
}

// SelectRadio 选择单选按钮（单选）
func SelectRadio(page playwright.Page, radioSelector SelectorConfig) error {
	return SelectRadios(page, []SelectorConfig{radioSelector})
}

// SelectRadios 选择多个单选按钮（不同组）
// 注意：同一组内的单选按钮只能选一个，但可以选择不同组的单选按钮
func SelectRadios(page playwright.Page, radioSelectors []SelectorConfig) error {
	if len(radioSelectors) == 0 {
		return fmt.Errorf("单选按钮列表不能为空")
	}

	for _, radioSelector := range radioSelectors {
		radioElement, err := LocateElement(page, radioSelector)
		if err != nil {
			return fmt.Errorf("定位单选按钮失败: %v", err)
		}

		err = radioElement.Check()
		if err != nil {
			return fmt.Errorf("选择单选按钮失败: %v", err)
		}

		time.Sleep(100 * time.Millisecond)
	}

	time.Sleep(200 * time.Millisecond)
	return nil
}

// SetCheckboxes 批量设置多个复选框
func SetCheckboxes(page playwright.Page, checkboxSelectors []SelectorConfig, checked bool) error {
	if len(checkboxSelectors) == 0 {
		return fmt.Errorf("复选框列表不能为空")
	}

	for _, checkboxSelector := range checkboxSelectors {
		err := SetCheckbox(page, checkboxSelector, checked)
		if err != nil {
			return fmt.Errorf("设置复选框失败: %v", err)
		}
	}

	return nil
}

// ToggleCheckboxes 批量切换多个复选框
func ToggleCheckboxes(page playwright.Page, checkboxSelectors []SelectorConfig) error {
	if len(checkboxSelectors) == 0 {
		return fmt.Errorf("复选框列表不能为空")
	}

	for _, checkboxSelector := range checkboxSelectors {
		err := ToggleCheckbox(page, checkboxSelector)
		if err != nil {
			return fmt.Errorf("切换复选框失败: %v", err)
		}
	}

	return nil
}

// GetSelectValue 获取下拉框的当前选中值
func GetSelectValue(page playwright.Page, selectSelector SelectorConfig) (string, error) {
	selectElement, err := LocateElement(page, selectSelector)
	if err != nil {
		return "", fmt.Errorf("定位下拉框失败: %v", err)
	}

	// 获取选中的option
	value, err := selectElement.Evaluate("el => { const sel = el.tagName.toLowerCase() === 'select' ? el : el.querySelector('select'); return sel ? sel.value : ''; }")
	if err != nil {
		return "", fmt.Errorf("获取下拉框值失败: %v", err)
	}

	if strValue, ok := value.(string); ok {
		return strValue, nil
	}

	return "", fmt.Errorf("下拉框值为空或格式错误")
}

// GetCheckboxState 获取复选框的选中状态
func GetCheckboxState(page playwright.Page, checkboxSelector SelectorConfig) (bool, error) {
	checkboxElement, err := LocateElement(page, checkboxSelector)
	if err != nil {
		return false, fmt.Errorf("定位复选框失败: %v", err)
	}

	return checkboxElement.IsChecked()
}

// GetRadioState 获取单选按钮的选中状态
func GetRadioState(page playwright.Page, radioSelector SelectorConfig) (bool, error) {
	radioElement, err := LocateElement(page, radioSelector)
	if err != nil {
		return false, fmt.Errorf("定位单选按钮失败: %v", err)
	}

	return radioElement.IsChecked()
}

// LocateSelectByLabel 通过label文本定位select元素
func LocateSelectByLabel(page playwright.Page, labelText string) (playwright.ElementHandle, error) {
	selectors := []string{
		fmt.Sprintf("label:has-text('%s') + select", labelText),
		fmt.Sprintf("label:has-text('%s') ~ select", labelText),
		fmt.Sprintf("//label[contains(text(), '%s')]/following-sibling::select[1]", labelText),
		fmt.Sprintf("//label[contains(text(), '%s')]/../select", labelText),
		fmt.Sprintf("//label[contains(text(), '%s')]/../div/select", labelText),
	}

	for _, selector := range selectors {
		locator := page.Locator(selector)
		count, err := locator.Count()
		if err == nil && count > 0 {
			element, err := locator.First().ElementHandle()
			if err == nil && element != nil {
				visible, _ := element.IsVisible()
				if visible {
					return element, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("无法通过label '%s' 定位select元素", labelText)
}

// LocateCheckboxByLabel 通过label文本定位checkbox元素
func LocateCheckboxByLabel(page playwright.Page, labelText string) (playwright.ElementHandle, error) {
	selectors := []string{
		fmt.Sprintf("label:has-text('%s') input[type='checkbox']", labelText),
		fmt.Sprintf("//label[contains(text(), '%s')]//input[@type='checkbox']", labelText),
		fmt.Sprintf("input[type='checkbox'] + label:has-text('%s')", labelText),
		fmt.Sprintf("//input[@type='checkbox']/following-sibling::label[contains(text(), '%s')]", labelText),
	}

	for _, selector := range selectors {
		locator := page.Locator(selector)
		count, err := locator.Count()
		if err == nil && count > 0 {
			element, err := locator.First().ElementHandle()
			if err == nil && element != nil {
				visible, _ := element.IsVisible()
				if visible {
					return element, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("无法通过label '%s' 定位checkbox元素", labelText)
}

// LocateRadioByLabel 通过label文本定位radio元素
func LocateRadioByLabel(page playwright.Page, labelText string) (playwright.ElementHandle, error) {
	selectors := []string{
		fmt.Sprintf("label:has-text('%s') input[type='radio']", labelText),
		fmt.Sprintf("//label[contains(text(), '%s')]//input[@type='radio']", labelText),
		fmt.Sprintf("input[type='radio'] + label:has-text('%s')", labelText),
		fmt.Sprintf("//input[@type='radio']/following-sibling::label[contains(text(), '%s')]", labelText),
	}

	for _, selector := range selectors {
		locator := page.Locator(selector)
		count, err := locator.Count()
		if err == nil && count > 0 {
			element, err := locator.First().ElementHandle()
			if err == nil && element != nil {
				visible, _ := element.IsVisible()
				if visible {
					return element, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("无法通过label '%s' 定位radio元素", labelText)
}

// GetSelectOptions 获取下拉框的所有选项
func GetSelectOptions(page playwright.Page, selectSelector SelectorConfig) ([]string, error) {
	selectElement, err := LocateElement(page, selectSelector)
	if err != nil {
		return nil, fmt.Errorf("定位下拉框失败: %v", err)
	}

	// 获取所有option的文本
	options, err := selectElement.Evaluate("el => { const sel = el.tagName.toLowerCase() === 'select' ? el : el.querySelector('select'); if (!sel) return []; return Array.from(sel.options).map(opt => opt.text); }")
	if err != nil {
		return nil, fmt.Errorf("获取选项列表失败: %v", err)
	}

	if optionsList, ok := options.([]interface{}); ok {
		result := make([]string, 0, len(optionsList))
		for _, opt := range optionsList {
			if str, ok := opt.(string); ok {
				result = append(result, str)
			}
		}
		return result, nil
	}

	return nil, fmt.Errorf("选项列表格式错误")
}
