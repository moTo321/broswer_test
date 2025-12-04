package utils

import (
	"fmt"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

// TableConfig 表格配置
type TableConfig struct {
	Selector SelectorConfig     `json:"selector"`         // 表格选择器
	Row      *TableRowConfig    `json:"row,omitempty"`    // 行定位配置
	Column   *TableColumnConfig `json:"column,omitempty"` // 列定位配置
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

// FindTableRow 查找表格行
// 根据条件查找表格中的行，返回行元素
// 如果 tableSelector 为空，则在当前页面查找所有表格
func FindTableRow(page playwright.Page, tableSelector SelectorConfig, rowConfig TableRowConfig) (playwright.ElementHandle, error) {
	// 获取表格的定位器
	var tableLocator playwright.Locator

	// 如果未指定表格选择器，默认查找页面中的第一个表格
	if tableSelector.Type == "" || tableSelector.Value == "" {
		// 查找页面中的第一个表格
		tableLocator = page.Locator("table").First()
		count, err := tableLocator.Count()
		if err != nil || count == 0 {
			return nil, fmt.Errorf("当前页面未找到表格")
		}
	} else {
		// 根据选择器类型定位表格
		switch tableSelector.Type {
		case "text":
			// 通过文本定位表格（通常是表格标题或label）
			tableLocator = page.Locator(fmt.Sprintf("//table[.//th[contains(text(), '%s')]]", tableSelector.Value))
		case "css", "id":
			tableLocator = page.Locator(tableSelector.Value)
		case "xpath":
			tableLocator = page.Locator(tableSelector.Value)
		default:
			tableLocator = page.Locator(tableSelector.Value)
		}

		// 验证表格是否存在
		count, err := tableLocator.Count()
		if err != nil || count == 0 {
			return nil, fmt.Errorf("定位表格失败: 未找到匹配的表格")
		}
	}

	// 根据行配置查找行
	switch rowConfig.Type {
	case "index":
		// 通过索引查找（从1开始）
		rows := tableLocator.Locator("tbody tr, tr")
		count, err := rows.Count()
		if err != nil {
			return nil, fmt.Errorf("获取表格行数失败: %v", err)
		}

		rowIndex := 0
		fmt.Sscanf(rowConfig.Value, "%d", &rowIndex)
		if rowIndex < 1 || rowIndex > count {
			return nil, fmt.Errorf("行索引超出范围: %d (总共 %d 行)", rowIndex, count)
		}

		rowElement, err := rows.Nth(rowIndex - 1).ElementHandle()
		if err != nil {
			return nil, fmt.Errorf("获取行元素失败: %v", err)
		}
		return rowElement, nil

	case "text", "contains":
		// 通过文本匹配查找行
		rows := tableLocator.Locator("tbody tr, tr")
		count, err := rows.Count()
		if err != nil {
			return nil, fmt.Errorf("获取表格行数失败: %v", err)
		}

		for i := 0; i < count; i++ {
			row := rows.Nth(i)
			text, err := row.TextContent()
			if err == nil {
				if rowConfig.Type == "text" {
					if strings.TrimSpace(text) == rowConfig.Value {
						rowElement, err := row.ElementHandle()
						if err == nil {
							return rowElement, nil
						}
					}
				} else if rowConfig.Type == "contains" {
					if strings.Contains(text, rowConfig.Value) {
						rowElement, err := row.ElementHandle()
						if err == nil {
							return rowElement, nil
						}
					}
				}
			}
		}

		return nil, fmt.Errorf("未找到匹配的行: %s", rowConfig.Value)

	default:
		return nil, fmt.Errorf("不支持的行定位类型: %s", rowConfig.Type)
	}
}

// FindTableCell 查找表格单元格
func FindTableCell(page playwright.Page, tableSelector SelectorConfig, rowConfig TableRowConfig, columnConfig TableColumnConfig) (playwright.ElementHandle, error) {
	// 先找到行
	rowElement, err := FindTableRow(page, tableSelector, rowConfig)
	if err != nil {
		return nil, err
	}

	// 获取列索引
	columnIndex := -1
	if columnConfig.Type == "index" {
		fmt.Sscanf(columnConfig.Value, "%d", &columnIndex)
		if columnIndex < 1 {
			return nil, fmt.Errorf("列索引必须大于0: %d", columnIndex)
		}
	} else if columnConfig.Type == "header" {
		// 通过表头文本查找列索引
		var tableLocator playwright.Locator
		if tableSelector.Type == "" || tableSelector.Value == "" {
			// 如果未指定表格选择器，使用页面中的第一个表格
			tableLocator = page.Locator("table").First()
		} else {
			tableLocator = page.Locator(tableSelector.Value)
		}
		headers := tableLocator.Locator("thead th, th")
		count, err := headers.Count()
		if err != nil {
			return nil, fmt.Errorf("获取表头数量失败: %v", err)
		}

		for i := 0; i < count; i++ {
			header := headers.Nth(i)
			text, err := header.TextContent()
			if err == nil && strings.Contains(text, columnConfig.Value) {
				columnIndex = i + 1
				break
			}
		}

		if columnIndex == -1 {
			return nil, fmt.Errorf("未找到表头: %s", columnConfig.Value)
		}
	} else {
		return nil, fmt.Errorf("不支持的列定位类型: %s", columnConfig.Type)
	}

	// 获取单元格（td或th）
	// 由于 ElementHandle 没有 Locator 方法，需要通过 page 重新定位行
	rowText, _ := rowElement.TextContent()

	// 重新定位表格和行
	var tableLocator playwright.Locator
	switch tableSelector.Type {
	case "css", "id":
		tableLocator = page.Locator(tableSelector.Value)
	case "xpath":
		tableLocator = page.Locator(tableSelector.Value)
	default:
		if tableSelector.Value == "" {
			tableLocator = page.Locator("table").First()
		} else {
			tableLocator = page.Locator(tableSelector.Value)
		}
	}

	rows := tableLocator.Locator("tbody tr, tr")
	count, _ := rows.Count()

	// 找到对应的行
	for i := 0; i < count; i++ {
		row := rows.Nth(i)
		text, _ := row.TextContent()
		if strings.Contains(text, rowText) || text == rowText {
			// 获取该行的单元格
			cells := row.Locator("td, th")
			cellElement, err := cells.Nth(columnIndex - 1).ElementHandle()
			if err != nil {
				return nil, fmt.Errorf("获取单元格失败: %v", err)
			}
			return cellElement, nil
		}
	}

	return nil, fmt.Errorf("无法重新定位行以获取单元格")
}

// ClickTableAction 点击表格中的操作按钮（编辑、删除等）
func ClickTableAction(page playwright.Page, tableSelector SelectorConfig, rowConfig TableRowConfig, actionText string) error {
	// 找到行
	rowElement, err := FindTableRow(page, tableSelector, rowConfig)
	if err != nil {
		return err
	}

	// 获取行的文本内容用于重新定位
	rowText, _ := rowElement.TextContent()

	// 重新通过 page 定位行，然后查找操作按钮
	var tableLocator playwright.Locator
	switch tableSelector.Type {
	case "css", "id":
		tableLocator = page.Locator(tableSelector.Value)
	case "xpath":
		tableLocator = page.Locator(tableSelector.Value)
	default:
		tableLocator = page.Locator(tableSelector.Value)
	}

	rows := tableLocator.Locator("tbody tr, tr")
	count, _ := rows.Count()

	// 找到对应的行
	for i := 0; i < count; i++ {
		row := rows.Nth(i)
		text, _ := row.TextContent()
		if strings.Contains(text, rowText) || text == rowText {
			// 在行中查找操作按钮
			actionSelectors := []string{
				fmt.Sprintf("text=%s", actionText),
				fmt.Sprintf("text=/.*%s.*/", escapeRegexForTable(actionText)),
				fmt.Sprintf("button:has-text('%s')", actionText),
				fmt.Sprintf("a:has-text('%s')", actionText),
				fmt.Sprintf("//button[contains(text(), '%s')]", actionText),
				fmt.Sprintf("//a[contains(text(), '%s')]", actionText),
			}

			for _, selector := range actionSelectors {
				actionLocator := row.Locator(selector)
				actionCount, err := actionLocator.Count()
				if err == nil && actionCount > 0 {
					err = actionLocator.First().Click()
					if err == nil {
						time.Sleep(300 * time.Millisecond)
						return nil
					}
				}
			}
			break
		}
	}

	return fmt.Errorf("未找到操作按钮: %s", actionText)
}

// GetTableRowData 获取表格行数据
func GetTableRowData(page playwright.Page, tableSelector SelectorConfig, rowConfig TableRowConfig) (map[string]string, error) {
	// 找到行
	rowElement, err := FindTableRow(page, tableSelector, rowConfig)
	if err != nil {
		return nil, err
	}

	// 获取行的文本内容用于重新定位
	rowText, _ := rowElement.TextContent()

	// 重新通过 page 定位行
	var tableLocator playwright.Locator
	switch tableSelector.Type {
	case "css", "id":
		tableLocator = page.Locator(tableSelector.Value)
	case "xpath":
		tableLocator = page.Locator(tableSelector.Value)
	default:
		if tableSelector.Value == "" {
			tableLocator = page.Locator("table").First()
		} else {
			tableLocator = page.Locator(tableSelector.Value)
		}
	}

	// 获取表头
	headers := tableLocator.Locator("thead th, th")
	headerCount, _ := headers.Count()

	// 找到对应的行
	rows := tableLocator.Locator("tbody tr, tr")
	count, _ := rows.Count()

	var rowLocator playwright.Locator
	for i := 0; i < count; i++ {
		row := rows.Nth(i)
		text, _ := row.TextContent()
		if strings.Contains(text, rowText) || text == rowText {
			rowLocator = row
			break
		}
	}

	if rowLocator == nil {
		return nil, fmt.Errorf("无法重新定位行")
	}

	// 获取单元格
	cells := rowLocator.Locator("td, th")
	cellCount, _ := cells.Count()

	// 构建数据映射
	data := make(map[string]string)
	for i := 0; i < cellCount && i < headerCount; i++ {
		header, _ := headers.Nth(i).TextContent()
		cell, _ := cells.Nth(i).TextContent()
		data[strings.TrimSpace(header)] = strings.TrimSpace(cell)
	}

	// 如果没有表头，使用列索引作为key
	if headerCount == 0 {
		for i := 0; i < cellCount; i++ {
			cell, _ := cells.Nth(i).TextContent()
			data[fmt.Sprintf("column_%d", i+1)] = strings.TrimSpace(cell)
		}
	}

	return data, nil
}

// AssertTableData 断言表格数据
func AssertTableData(page playwright.Page, tableSelector SelectorConfig, rowConfig TableRowConfig, columnConfig TableColumnConfig, expectedValue string, mode string) error {
	// 找到单元格
	cellElement, err := FindTableCell(page, tableSelector, rowConfig, columnConfig)
	if err != nil {
		return err
	}

	// 获取单元格文本
	actualText, err := cellElement.TextContent()
	if err != nil {
		return fmt.Errorf("获取单元格文本失败: %v", err)
	}
	actualText = strings.TrimSpace(actualText)

	// 根据模式进行断言
	switch mode {
	case "equals":
		if actualText != expectedValue {
			return fmt.Errorf("表格数据断言失败: 期望 '%s', 实际 '%s'", expectedValue, actualText)
		}
	case "contains":
		if !strings.Contains(actualText, expectedValue) {
			return fmt.Errorf("表格数据断言失败: 期望包含 '%s', 实际 '%s'", expectedValue, actualText)
		}
	case "not_equals":
		if actualText == expectedValue {
			return fmt.Errorf("表格数据断言失败: 期望不等于 '%s', 但实际是 '%s'", expectedValue, actualText)
		}
	case "not_contains":
		if strings.Contains(actualText, expectedValue) {
			return fmt.Errorf("表格数据断言失败: 期望不包含 '%s', 但实际包含", expectedValue)
		}
	default:
		return fmt.Errorf("不支持的断言模式: %s", mode)
	}

	return nil
}

// escapeRegexForTable 转义正则表达式特殊字符（用于表格操作）
func escapeRegexForTable(text string) string {
	specialChars := []string{"\\", "^", "$", ".", "|", "?", "*", "+", "(", ")", "[", "]", "{", "}"}
	result := text
	for _, char := range specialChars {
		result = strings.ReplaceAll(result, char, "\\"+char)
	}
	return result
}
