package runner

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"autotest/utils"

	"github.com/playwright-community/playwright-go"
)

// SelectorConfig 选择器配置（使用 utils 中的定义）
type SelectorConfig = utils.SelectorConfig

// ExpectConfig 期望验证配置
type ExpectConfig struct {
	Type  string `json:"type"`           // "text", "xpath", "css", "id"
	Value string `json:"value"`          // 选择器的值
	Mode  string `json:"mode"`           // "value_equals", "text_equals", "text_contains", "visible"
	Text  string `json:"text,omitempty"` // 期望的文本内容
}

// TestStep 测试步骤
type TestStep struct {
	Action   string            `json:"action"`             // "goto", "input", "click", "assert", "menu_click", "captcha_input", "select_option", "select_options", "checkbox_toggle", "checkbox_set", "checkboxes_set", "radio_select", "radios_select", "table_edit", "table_delete", "table_assert", "search"
	URL      string            `json:"url,omitempty"`      // goto的URL
	Selector *SelectorConfig   `json:"selector,omitempty"` // 元素选择器（单个）
	Selectors []SelectorConfig `json:"selectors,omitempty"` // 元素选择器（多个，用于批量操作）
	Text     string            `json:"text,omitempty"`     // input的文本内容，或select的选项值（单个）
	Options  []string          `json:"options,omitempty"`  // select的选项值（多个，用于多选）
	Expect   *ExpectConfig     `json:"expect,omitempty"`   // 期望验证配置
	MenuPath string            `json:"menu_path,omitempty"` // 菜单路径，格式: "系统管理 > 用户管理 > 新增用户"
	Captcha  *CaptchaConfig    `json:"captcha,omitempty"`  // 验证码配置
	Checked  *bool             `json:"checked,omitempty"`  // checkbox_set时使用，true表示选中，false表示取消选中
	Table    *TableConfig      `json:"table,omitempty"`   // 表格配置
	Search   *SearchConfig     `json:"search,omitempty"`  // 查询配置
}

// TableConfig 表格配置
type TableConfig struct {
	Selector SelectorConfig      `json:"selector"` // 表格选择器
	Row      *TableRowConfig    `json:"row,omitempty"` // 行定位配置
	Column   *TableColumnConfig `json:"column,omitempty"` // 列定位配置
	Action   string             `json:"action,omitempty"` // 操作类型: "edit", "delete"
	Value    string             `json:"value,omitempty"` // 断言期望值
	Mode     string             `json:"mode,omitempty"` // 断言模式: "equals", "contains", "not_equals", "not_contains"
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

// SearchConfig 查询配置
type SearchConfig struct {
	Inputs []SearchInput    `json:"inputs,omitempty"` // 查询输入框配置
	Button *SelectorConfig  `json:"button,omitempty"` // 查询按钮选择器
}

// SearchInput 查询输入配置
type SearchInput struct {
	Selector *SelectorConfig `json:"selector"` // 输入框选择器
	Text     string          `json:"text"`     // 输入文本
}

// CaptchaConfig 验证码配置
type CaptchaConfig struct {
	ImageSelector *SelectorConfig `json:"image_selector,omitempty"` // 验证码图片选择器
	InputSelector *SelectorConfig `json:"input_selector,omitempty"` // 验证码输入框选择器
	Auto          bool            `json:"auto,omitempty"`           // 是否自动识别（自动查找验证码图片和输入框）
}

// TestCase 测试用例
type TestCase struct {
	Name  string     `json:"name"`
	Steps []TestStep `json:"steps"`
}

// TestSuite 测试套件（支持多个用例）
type TestSuite []TestCase

// Runner 测试运行器
type Runner struct {
	page playwright.Page
}

// NewRunner 创建新的测试运行器
func NewRunner(page playwright.Page) *Runner {
	return &Runner{page: page}
}

// RunTestCase 执行单个测试用例
func (r *Runner) RunTestCase(testCase TestCase) error {
	fmt.Printf("📋 开始执行用例: %s\n", testCase.Name)

	for i, step := range testCase.Steps {
		fmt.Printf("  [%d/%d] 执行步骤: %s\n", i+1, len(testCase.Steps), step.Action)

		var err error
		switch step.Action {
		case "goto":
			err = r.handleGoto(step)
		case "input":
			err = r.handleInput(step)
		case "click":
			err = r.handleClick(step)
		case "assert":
			err = r.handleAssert(step)
		case "menu_click":
			err = r.handleMenuClick(step)
		case "captcha_input":
			err = r.handleCaptchaInput(step)
		case "select_option":
			err = r.handleSelectOption(step)
		case "select_options":
			err = r.handleSelectOptions(step)
		case "checkbox_toggle":
			err = r.handleCheckboxToggle(step)
		case "checkbox_set":
			err = r.handleCheckboxSet(step)
		case "checkboxes_set":
			err = r.handleCheckboxesSet(step)
		case "radio_select":
			err = r.handleRadioSelect(step)
		case "radios_select":
			err = r.handleRadiosSelect(step)
		case "table_edit":
			err = r.handleTableEdit(step)
		case "table_delete":
			err = r.handleTableDelete(step)
		case "table_assert":
			err = r.handleTableAssert(step)
		case "search":
			err = r.handleSearch(step)
		default:
			err = fmt.Errorf("未知的 action: %s", step.Action)
		}

		if err != nil {
			// 错误截图
			timestamp := time.Now().Unix()
			screenshotPath := fmt.Sprintf("assets/errors/error_%d.png", timestamp)
			r.page.Screenshot(playwright.PageScreenshotOptions{
				Path: playwright.String(screenshotPath),
			})
			return fmt.Errorf("步骤 [%d] %s 执行失败: %v", i+1, step.Action, err)
		}

		// 步骤间等待
		time.Sleep(300 * time.Millisecond)
	}

	fmt.Printf("✅ 用例执行完成: %s\n", testCase.Name)
	return nil
}

// RunTestSuite 执行测试套件
func (r *Runner) RunTestSuite(suite TestSuite) error {
	for _, testCase := range suite {
		if err := r.RunTestCase(testCase); err != nil {
			return err
		}
	}
	return nil
}

// RunTestSuiteFromFile 从文件加载并执行测试套件
func (r *Runner) RunTestSuiteFromFile(filePath string) error {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取测试文件失败: %v", err)
	}

	var suite TestSuite
	err = json.Unmarshal(content, &suite)
	if err != nil {
		return fmt.Errorf("解析测试文件失败: %v", err)
	}

	return r.RunTestSuite(suite)
}

// handleGoto 处理页面跳转
func (r *Runner) handleGoto(step TestStep) error {
	if step.URL == "" {
		return errors.New("goto action 需要提供 url")
	}
	_, err := r.page.Goto(step.URL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	return err
}

// handleInput 处理输入操作
func (r *Runner) handleInput(step TestStep) error {
	if step.Selector == nil {
		return errors.New("input action 需要提供 selector")
	}
	if step.Text == "" {
		return errors.New("input action 需要提供 text")
	}

	// 定位元素
	selector := utils.SelectorConfig{
		Type:  step.Selector.Type,
		Value: step.Selector.Value,
		Scope: step.Selector.Scope,
	}
	element, err := utils.LocateElement(r.page, selector)
	if err != nil {
		return fmt.Errorf("定位输入框失败: %v", err)
	}

	// 清空并输入
	err = element.Fill(step.Text)
	if err != nil {
		return fmt.Errorf("输入文本失败: %v", err)
	}

	// 如果有expect验证，执行验证
	if step.Expect != nil {
		return r.verifyExpect(step.Expect, step.Text)
	}

	return nil
}

// handleClick 处理点击操作
func (r *Runner) handleClick(step TestStep) error {
	if step.Selector == nil {
		return errors.New("click action 需要提供 selector")
	}

	// 定位元素
	selector := utils.SelectorConfig{
		Type:  step.Selector.Type,
		Value: step.Selector.Value,
		Scope: step.Selector.Scope,
	}
	element, err := utils.LocateElement(r.page, selector)
	if err != nil {
		return fmt.Errorf("定位元素失败: %v", err)
	}

	// 确保元素在可视区域
	_ = element.ScrollIntoViewIfNeeded()

	// 使用 Playwright 点击元素（force 避免因轻微遮挡导致无法点击）
	err = element.Click(playwright.ElementHandleClickOptions{
		Force: playwright.Bool(true),
	})
	if err != nil {
		return fmt.Errorf("点击失败: %v", err)
	}

	// 等待页面响应
	time.Sleep(500 * time.Millisecond)

	return nil
}

// handleAssert 处理断言操作
func (r *Runner) handleAssert(step TestStep) error {
	if step.Selector == nil {
		return errors.New("assert action 需要提供 selector")
	}

	// 定位元素
	selector := utils.SelectorConfig{
		Type:  step.Selector.Type,
		Value: step.Selector.Value,
		Scope: step.Selector.Scope,
	}
	element, err := utils.LocateElement(r.page, selector)
	if err != nil {
		return fmt.Errorf("定位元素失败: %v", err)
	}

	// 检查元素是否可见
	visible, err := element.IsVisible()
	if err != nil {
		return fmt.Errorf("检查元素可见性失败: %v", err)
	}
	if !visible {
		return errors.New("断言失败: 元素不可见")
	}

	// 如果有expect配置，进行更详细的验证
	if step.Expect != nil {
		return r.verifyExpect(step.Expect, "")
	}

	return nil
}

// verifyExpect 验证期望结果
func (r *Runner) verifyExpect(expect *ExpectConfig, inputText string) error {
	// 定位期望验证的元素
	element, err := utils.LocateElement(r.page, utils.SelectorConfig{
		Type:  expect.Type,
		Value: expect.Value,
	})
	if err != nil {
		return fmt.Errorf("定位期望元素失败: %v", err)
	}

	switch expect.Mode {
	case "value_equals":
		// 验证输入框的值
		value, err := utils.GetElementValue(element)
		if err != nil {
			return fmt.Errorf("获取元素值失败: %v", err)
		}
		if value != expect.Text {
			return fmt.Errorf("值验证失败: 期望 '%s', 实际 '%s'", expect.Text, value)
		}

	case "text_equals":
		// 验证文本内容完全匹配
		text, err := utils.GetElementText(element)
		if err != nil {
			return fmt.Errorf("获取元素文本失败: %v", err)
		}
		if text != expect.Text {
			return fmt.Errorf("文本验证失败: 期望 '%s', 实际 '%s'", expect.Text, text)
		}

	case "text_contains":
		// 验证文本内容包含
		text, err := utils.GetElementText(element)
		if err != nil {
			return fmt.Errorf("获取元素文本失败: %v", err)
		}
		// 使用strings.Contains进行简单的包含检查
		if !strings.Contains(text, expect.Text) {
			return fmt.Errorf("文本包含验证失败: 期望包含 '%s', 实际文本 '%s'", expect.Text, text)
		}

	case "visible":
		// 验证元素可见
		visible, err := utils.IsElementVisible(element)
		if err != nil {
			return fmt.Errorf("检查元素可见性失败: %v", err)
		}
		if !visible {
			return errors.New("可见性验证失败: 元素不可见")
		}

	default:
		return fmt.Errorf("未知的验证模式: %s", expect.Mode)
	}

	return nil
}

// handleMenuClick 处理菜单点击操作
func (r *Runner) handleMenuClick(step TestStep) error {
	if step.MenuPath == "" {
		return errors.New("menu_click action 需要提供 menu_path")
	}

	return utils.ClickMenu(r.page, step.MenuPath)
}

// handleCaptchaInput 处理验证码识别和输入操作
func (r *Runner) handleCaptchaInput(step TestStep) error {
	if step.Captcha == nil {
		return errors.New("captcha_input action 需要提供 captcha 配置")
	}

	// 如果启用自动识别
	if step.Captcha.Auto {
		_, err := utils.AutoSolveCaptcha(r.page)
		return err
	}

	// 手动指定选择器
	if step.Captcha.ImageSelector == nil || step.Captcha.InputSelector == nil {
		return errors.New("captcha_input action 需要提供 image_selector 和 input_selector，或设置 auto: true")
	}

	// 转换选择器类型
	imageSelector := utils.SelectorConfig{
		Type:  step.Captcha.ImageSelector.Type,
		Value: step.Captcha.ImageSelector.Value,
		Scope: step.Captcha.ImageSelector.Scope,
	}
	inputSelector := utils.SelectorConfig{
		Type:  step.Captcha.InputSelector.Type,
		Value: step.Captcha.InputSelector.Value,
		Scope: step.Captcha.InputSelector.Scope,
	}

	_, err := utils.SolveAndInputCaptcha(r.page, imageSelector, inputSelector)
	return err
}

// handleSelectOption 处理下拉框选择操作
func (r *Runner) handleSelectOption(step TestStep) error {
	if step.Selector == nil {
		return errors.New("select_option action 需要提供 selector")
	}
	if step.Text == "" {
		return errors.New("select_option action 需要提供 text（选项文本或值）")
	}

	selector := utils.SelectorConfig{
		Type:  step.Selector.Type,
		Value: step.Selector.Value,
		Scope: step.Selector.Scope,
	}

	return utils.SelectOption(r.page, selector, step.Text)
}

// handleCheckboxToggle 处理复选框切换操作
func (r *Runner) handleCheckboxToggle(step TestStep) error {
	if step.Selector == nil {
		return errors.New("checkbox_toggle action 需要提供 selector")
	}

	selector := utils.SelectorConfig{
		Type:  step.Selector.Type,
		Value: step.Selector.Value,
		Scope: step.Selector.Scope,
	}

	return utils.ToggleCheckbox(r.page, selector)
}

// handleCheckboxSet 处理复选框设置操作
func (r *Runner) handleCheckboxSet(step TestStep) error {
	if step.Selector == nil {
		return errors.New("checkbox_set action 需要提供 selector")
	}
	if step.Checked == nil {
		return errors.New("checkbox_set action 需要提供 checked 字段（true/false）")
	}

	selector := utils.SelectorConfig{
		Type:  step.Selector.Type,
		Value: step.Selector.Value,
		Scope: step.Selector.Scope,
	}

	return utils.SetCheckbox(r.page, selector, *step.Checked)
}

// handleRadioSelect 处理单选按钮选择操作
func (r *Runner) handleRadioSelect(step TestStep) error {
	if step.Selector == nil {
		return errors.New("radio_select action 需要提供 selector")
	}

	selector := utils.SelectorConfig{
		Type:  step.Selector.Type,
		Value: step.Selector.Value,
		Scope: step.Selector.Scope,
	}

	return utils.SelectRadio(r.page, selector)
}

// handleSelectOptions 处理下拉框多选操作
func (r *Runner) handleSelectOptions(step TestStep) error {
	if step.Selector == nil {
		return errors.New("select_options action 需要提供 selector")
	}
	if len(step.Options) == 0 {
		return errors.New("select_options action 需要提供 options（选项数组）")
	}

	selector := utils.SelectorConfig{
		Type:  step.Selector.Type,
		Value: step.Selector.Value,
		Scope: step.Selector.Scope,
	}

	return utils.SelectOptions(r.page, selector, step.Options)
}

// handleCheckboxesSet 处理批量复选框设置操作
func (r *Runner) handleCheckboxesSet(step TestStep) error {
	if len(step.Selectors) == 0 {
		return errors.New("checkboxes_set action 需要提供 selectors（选择器数组）")
	}
	if step.Checked == nil {
		return errors.New("checkboxes_set action 需要提供 checked 字段（true/false）")
	}

	selectors := make([]utils.SelectorConfig, len(step.Selectors))
	for i, sel := range step.Selectors {
		selectors[i] = utils.SelectorConfig{
			Type:  sel.Type,
			Value: sel.Value,
			Scope: sel.Scope,
		}
	}

	return utils.SetCheckboxes(r.page, selectors, *step.Checked)
}

// handleRadiosSelect 处理多个单选按钮选择操作
func (r *Runner) handleRadiosSelect(step TestStep) error {
	if len(step.Selectors) == 0 {
		return errors.New("radios_select action 需要提供 selectors（选择器数组）")
	}

	selectors := make([]utils.SelectorConfig, len(step.Selectors))
	for i, sel := range step.Selectors {
		selectors[i] = utils.SelectorConfig{
			Type:  sel.Type,
			Value: sel.Value,
			Scope: sel.Scope,
		}
	}

	return utils.SelectRadios(r.page, selectors)
}

// handleTableEdit 处理表格编辑操作
func (r *Runner) handleTableEdit(step TestStep) error {
	if step.Table == nil {
		return errors.New("table_edit action 需要提供 table 配置")
	}
	if step.Table.Row == nil {
		return errors.New("table_edit action 需要提供 table.row 配置")
	}

	// 如果未指定表格选择器，使用空配置（将自动查找页面中的第一个表格）
	tableSelector := utils.SelectorConfig{
		Type:  step.Table.Selector.Type,
		Value: step.Table.Selector.Value,
	}

	rowConfig := utils.TableRowConfig{
		Type:  step.Table.Row.Type,
		Value: step.Table.Row.Value,
	}

	actionText := "编辑"
	if step.Table.Action != "" {
		actionText = step.Table.Action
	}

	return utils.ClickTableAction(r.page, tableSelector, rowConfig, actionText)
}

// handleTableDelete 处理表格删除操作
func (r *Runner) handleTableDelete(step TestStep) error {
	if step.Table == nil {
		return errors.New("table_delete action 需要提供 table 配置")
	}
	if step.Table.Row == nil {
		return errors.New("table_delete action 需要提供 table.row 配置")
	}

	// 如果未指定表格选择器，使用空配置（将自动查找页面中的第一个表格）
	tableSelector := utils.SelectorConfig{
		Type:  step.Table.Selector.Type,
		Value: step.Table.Selector.Value,
	}

	rowConfig := utils.TableRowConfig{
		Type:  step.Table.Row.Type,
		Value: step.Table.Row.Value,
	}

	actionText := "删除"
	if step.Table.Action != "" {
		actionText = step.Table.Action
	}

	return utils.ClickTableAction(r.page, tableSelector, rowConfig, actionText)
}

// handleTableAssert 处理表格断言操作
func (r *Runner) handleTableAssert(step TestStep) error {
	if step.Table == nil {
		return errors.New("table_assert action 需要提供 table 配置")
	}
	if step.Table.Row == nil {
		return errors.New("table_assert action 需要提供 table.row 配置")
	}
	if step.Table.Column == nil {
		return errors.New("table_assert action 需要提供 table.column 配置")
	}
	if step.Table.Value == "" {
		return errors.New("table_assert action 需要提供 table.value（期望值）")
	}

	// 如果未指定表格选择器，使用空配置（将自动查找页面中的第一个表格）
	tableSelector := utils.SelectorConfig{
		Type:  step.Table.Selector.Type,
		Value: step.Table.Selector.Value,
	}

	rowConfig := utils.TableRowConfig{
		Type:  step.Table.Row.Type,
		Value: step.Table.Row.Value,
	}

	columnConfig := utils.TableColumnConfig{
		Type:  step.Table.Column.Type,
		Value: step.Table.Column.Value,
	}

	mode := step.Table.Mode
	if mode == "" {
		mode = "equals" // 默认完全匹配
	}

	return utils.AssertTableData(r.page, tableSelector, rowConfig, columnConfig, step.Table.Value, mode)
}

// handleSearch 处理查询操作
func (r *Runner) handleSearch(step TestStep) error {
	if step.Search == nil {
		return errors.New("search action 需要提供 search 配置")
	}
	if step.Search.Button == nil {
		return errors.New("search action 需要提供 search.button（查询按钮选择器）")
	}

	// 输入查询条件
	if len(step.Search.Inputs) > 0 {
		for _, input := range step.Search.Inputs {
			if input.Selector == nil {
				return errors.New("search.inputs 中的每个输入需要提供 selector")
			}

			selector := utils.SelectorConfig{
				Type:  input.Selector.Type,
				Value: input.Selector.Value,
			}

			element, err := utils.LocateElement(r.page, selector)
			if err != nil {
				return fmt.Errorf("定位查询输入框失败: %v", err)
			}

			err = element.Fill(input.Text)
			if err != nil {
				return fmt.Errorf("输入查询条件失败: %v", err)
			}

			time.Sleep(200 * time.Millisecond)
		}
	}

	// 点击查询按钮
	buttonSelector := utils.SelectorConfig{
		Type:  step.Search.Button.Type,
		Value: step.Search.Button.Value,
	}

	buttonElement, err := utils.LocateElement(r.page, buttonSelector)
	if err != nil {
		return fmt.Errorf("定位查询按钮失败: %v", err)
	}

	err = buttonElement.Click()
	if err != nil {
		return fmt.Errorf("点击查询按钮失败: %v", err)
	}

	// 等待查询结果
	time.Sleep(500 * time.Millisecond)

	return nil
}
