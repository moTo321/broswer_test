# 自动化测试框架

## 功能特性

### 🎯 基于文本的元素定位
- **简单易用**：测试人员只需使用页面上的文本内容即可定位元素，无需学习复杂的xpath
- **智能匹配**：支持多种定位策略，自动尝试最佳匹配方式
  - 直接文本匹配（按钮、链接等）
  - placeholder属性（输入框）
  - label标签关联
  - aria-label属性
  - title属性
  - data-*属性

### 📝 测试用例格式
使用简单的JSON格式编写测试用例，支持以下操作：

- **goto**: 页面跳转
- **input**: 文本输入（支持expect验证）
- **click**: 元素点击
- **assert**: 元素断言
- **menu_click**: 多级菜单点击（支持 `>` 分隔符）
- **captcha_input**: 验证码识别和自动输入
- **select_option**: 下拉框选择（单选）
- **select_options**: 下拉框选择（多选，支持 `<select multiple>`）
- **checkbox_toggle**: 复选框切换（选中/取消选中）
- **checkbox_set**: 复选框设置（指定选中状态）
- **checkboxes_set**: 批量复选框设置（同时设置多个复选框）
- **radio_select**: 单选按钮选择（单个）
- **radios_select**: 单选按钮选择（多个，不同组）
- **table_edit**: 表格编辑操作
- **table_delete**: 表格删除操作
- **table_assert**: 表格数据断言
- **search**: 查询操作（输入查询条件并点击查询按钮）

### ✅ 验证功能
- `value_equals`: 验证输入框的值
- `text_equals`: 验证文本内容完全匹配
- `text_contains`: 验证文本内容包含
- `visible`: 验证元素可见性

## 测试用例示例

```json
[
  {
    "name": "用户登录",
    "steps": [
      { "action": "goto", "url": "https://example.com/login" },
      {
        "action": "input",
        "selector": { "type": "text", "value": "用户名输入框" },
        "text": "testuser",
        "expect": { 
          "type": "text", 
          "value": "用户名输入框", 
          "mode": "value_equals", 
          "text": "testuser" 
        }
      },
      {
        "action": "input",
        "selector": { "type": "text", "value": "密码输入框" },
        "text": "password123"
      },
      { 
        "action": "click", 
        "selector": { "type": "text", "value": "登录" } 
      },
      { 
        "action": "assert", 
        "selector": { "type": "text", "value": "欢迎您，testuser" } 
      }
    ]
  }
]
```

## 选择器类型

### text（文本定位，推荐）
```json
{ "type": "text", "value": "登录" }
{ "type": "text", "value": "用户名输入框" }
```

### xpath（高级用法）
```json
{ "type": "xpath", "value": "//h2[contains(text(), '手机')]" }
```

### css（CSS选择器）
```json
{ "type": "css", "value": ".submit-button" }
```

### id（ID选择器）
```json
{ "type": "id", "value": "username" }
```

## 使用方法

### 1. 环境配置

#### 使用 Makefile（推荐）
```bash
# 初始化项目（自动安装依赖、创建目录等）
make init
```

#### 手动安装
```bash
# 1. 初始化 Go 模块
go mod init autotest
go mod tidy

# 2. 安装 Playwright
go install github.com/playwright-community/playwright-go/cmd/playwright@latest
playwright install chromium

# 3. 安装 Tesseract OCR（验证码识别功能需要）
# Debian/Ubuntu:
sudo apt-get install tesseract-ocr libtesseract-dev

# CentOS/RHEL:
sudo yum install tesseract tesseract-devel

# macOS:
brew install tesseract
```

**注意**: 
- 如果不需要验证码识别功能，可以不安装 Tesseract OCR
- 但编译时会出现警告，不影响其他功能使用
- 使用 `make install-tesseract` 可以自动安装（根据系统类型）

### 2. 编写测试用例
在 `testcase/` 目录下创建JSON格式的测试用例文件

### 3. 配置说明
配置文件 `config.yaml` 支持以下选项：
```yaml
browser: chromium      # 浏览器类型: chromium, firefox, webkit
headless: false        # 是否无头模式
timeout: 5000          # 超时时间（毫秒）
retry_captcha: 3       # 验证码重试次数
```

### 4. 运行测试

#### 使用 Makefile（推荐）

```bash
# 编译项目
make build

# 运行编译后的程序
make run-bin

# 运行指定测试用例
make test TEST_FILE=testcase/login_example.json

# 使用编译后的程序运行测试
make test-bin TEST_FILE=testcase/login_example.json

# 查看所有可用命令
make help
```

#### 直接使用 Go 命令

**命令行参数**
- `-c`: 指定配置文件路径（默认: `config.yaml`）
- `-f`: 指定测试用例文件路径（默认: `testcase/login_example.json`）
- `-h`: 显示帮助信息

**使用示例**
```bash
# 使用默认配置和默认测试文件
go run main.go

# 指定测试文件
go run main.go -f testcase/login_example.json

# 指定配置文件和测试文件
go run main.go -c config.yaml -f testcase/login_example.json

# 只指定配置文件
go run main.go -c my_config.yaml

# 显示帮助信息
go run main.go -h

# 编译后运行
go build -o autotest main.go
./autotest -f testcase/login_example.json
```

## 技术方案说明

### 为什么选择基于文本定位？

1. **简单易学**：测试人员无需学习xpath、CSS选择器等复杂语法
2. **直观明了**：直接使用页面上可见的文本内容，符合测试人员的思维习惯
3. **维护方便**：当页面结构变化时，只要文本内容不变，测试用例仍然有效
4. **智能匹配**：系统会自动尝试多种定位策略，提高成功率

### 定位策略优先级

1. 精确文本匹配（`text=用户名输入框`）
2. 正则文本匹配
3. placeholder属性匹配
4. label关联匹配
5. aria-label属性匹配
6. title属性匹配
7. data-*属性匹配

### 注意事项

- 文本定位适用于大多数常见场景，但对于动态生成的文本或重复文本，建议使用xpath或css选择器
- 系统会自动等待元素可见后再进行操作
- 测试失败时会自动截图保存到 `assets/errors/` 目录

## 高级功能

### 菜单点击 (menu_click)
支持多级菜单导航，使用 `>` 分隔菜单层级：

```json
{
  "action": "menu_click",
  "menu_path": "系统管理 > 用户管理 > 新增用户"
}
```

系统会自动：
1. 逐级点击菜单项
2. 等待子菜单展开
3. 处理菜单悬停和点击

### 验证码识别 (captcha_input)
自动识别并输入验证码，支持两种模式：

#### 自动识别模式（推荐）
系统会自动查找验证码图片和输入框：
```json
{
  "action": "captcha_input",
  "captcha": {
    "auto": true
  }
}
```

#### 手动指定选择器模式
```json
{
  "action": "captcha_input",
  "captcha": {
    "auto": false,
    "image_selector": { "type": "css", "value": "img.captcha-image" },
    "input_selector": { "type": "text", "value": "验证码" }
  }
}
```

验证码识别流程：
1. 定位验证码图片并截图
2. 使用 OCR 识别验证码文本
3. 自动输入到验证码输入框
4. 验证码图片保存在 `assets/captcha/` 目录

### 表单元素操作

#### 下拉框选择 (select_option)
支持通过文本定位下拉框并选择选项：

```json
{
  "action": "select_option",
  "selector": { "type": "text", "value": "城市" },
  "text": "北京"
}
```

系统会自动：
1. 通过label文本定位select元素
2. 支持标准select和自定义下拉框
3. 通过选项文本或值进行选择

#### 复选框操作

**切换复选框 (checkbox_toggle)**
```json
{
  "action": "checkbox_toggle",
  "selector": { "type": "text", "value": "同意协议" }
}
```

**设置复选框状态 (checkbox_set)**
```json
{
  "action": "checkbox_set",
  "selector": { "type": "text", "value": "接收邮件通知" },
  "checked": true
}
```

#### 单选按钮选择 (radio_select)
```json
{
  "action": "radio_select",
  "selector": { "type": "text", "value": "男" }
}
```

### 多选功能

#### 下拉框多选 (select_options)
支持在 `<select multiple>` 中选择多个选项：

```json
{
  "action": "select_options",
  "selector": { "type": "text", "value": "兴趣爱好" },
  "options": ["阅读", "运动", "音乐"]
}
```

系统会自动：
1. 检测是否为 multiple select
2. 支持标准 multiple select 和自定义下拉框
3. 对于自定义下拉框，使用 Ctrl+点击进行多选

#### 批量复选框操作 (checkboxes_set)
同时设置多个复选框的状态：

```json
{
  "action": "checkboxes_set",
  "selectors": [
    { "type": "text", "value": "接收邮件通知" },
    { "type": "text", "value": "接收短信通知" },
    { "type": "text", "value": "接收推送通知" }
  ],
  "checked": true
}
```

#### 多个单选按钮选择 (radios_select)
选择多个不同组的单选按钮（同一组内只能选一个）：

```json
{
  "action": "radios_select",
  "selectors": [
    { "type": "text", "value": "男" },
    { "type": "text", "value": "已婚" },
    { "type": "text", "value": "本科" }
  ]
}
```

**注意**：单选按钮通常在同一组内只能选一个，但可以选择不同组的单选按钮。

### 表格操作

#### 表格编辑 (table_edit)
在表格中点击编辑按钮：

```json
{
  "action": "table_edit",
  "table": {
    "selector": { "type": "css", "value": "#user_table" },
    "row": { "type": "contains", "value": "张三" }
  }
}
```

**行定位方式：**
- `index`: 通过行索引（从1开始）
- `text`: 行文本完全匹配
- `contains`: 行文本包含指定内容

#### 表格删除 (table_delete)
在表格中点击删除按钮：

```json
{
  "action": "table_delete",
  "table": {
    "selector": { "type": "css", "value": "#user_table" },
    "row": { "type": "index", "value": "1" }
  }
}
```

#### 表格数据断言 (table_assert)
断言表格中某个单元格的数据：

```json
{
  "action": "table_assert",
  "table": {
    "selector": { "type": "css", "value": "#user_table" },
    "row": { "type": "contains", "value": "张三" },
    "column": { "type": "header", "value": "姓名" },
    "value": "张三",
    "mode": "equals"
  }
}
```

**列定位方式：**
- `index`: 通过列索引（从1开始）
- `header`: 通过表头文本

**断言模式：**
- `equals`: 完全匹配
- `contains`: 包含
- `not_equals`: 不等于
- `not_contains`: 不包含

### 查询功能

#### 查询操作 (search)
输入查询条件并点击查询按钮：

```json
{
  "action": "search",
  "search": {
    "inputs": [
      {
        "selector": { "type": "text", "value": "用户名" },
        "text": "张三"
      },
      {
        "selector": { "type": "text", "value": "状态" },
        "text": "启用"
      }
    ],
    "button": { "type": "text", "value": "查询" }
  }
}
```

系统会自动：
1. 依次输入所有查询条件
2. 点击查询按钮
3. 等待查询结果加载

### 元素定位增强

系统现在支持更多元素类型的定位：

- **输入框**: input, textarea
- **下拉框**: select（通过label或选项文本定位）
- **复选框**: checkbox（通过label定位）
- **单选按钮**: radio（通过label定位）
- **按钮**: button, a, span等
- **菜单**: li, a, div等

定位策略自动尝试：
1. 直接文本匹配
2. placeholder属性
3. label标签关联（支持input、select、textarea、checkbox、radio）
4. aria-label属性
5. title属性
6. data-*属性
7. select选项文本定位
8. checkbox/radio的value或label定位

### 其他功能

- OCR 自动识别验证码
- 自动截图、录屏
- 自动生成测试报告
- 多级菜单点击、表格/列表断言
