package main

import (
	apistemplate "autotest/apis-template"
	browseTemplate "autotest/browse-template"
	"autotest/runner"

	"bufio"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// 定义命令行参数
	var (
		browseConfigFile = flag.String("c", "browse-template/browse-config.yaml", "配置playright浏览器文件路径")
		apiTemplateFile  = flag.String("a", "apis-template/apis.json", "API模板文件路径")
		testFile         = flag.String("f", "testcase/apis/api_test.json", "测试用例文件路径")
		help             = flag.Bool("h", false, "显示帮助信息")
	)

	// 解析命令行参数
	flag.Parse()

	// 显示帮助信息
	if *help {
		showHelp()
		return
	}

	fmt.Println("🚀 自动化测试框架启动")

	// 加载配置
	fmt.Printf("📋 加载配置文件: %s\n", *browseConfigFile)
	cfg, err := browseTemplate.LoadConfig(*browseConfigFile)
	if err != nil {
		fmt.Printf("⚠️  配置加载失败，使用默认配置: %v\n", err)
		cfg = browseTemplate.DefaultConfig()
	} else {
		fmt.Printf("   - 浏览器: %s\n", cfg.Browser)
		fmt.Printf("   - 无头模式: %t\n", cfg.Headless)
		fmt.Printf("   - 超时时间: %dms\n", cfg.Timeout)
	}

	// 2. 加载 API Templates (新增步骤)
	fmt.Printf("📋 加载 API 模板: %s\n", *apiTemplateFile)
	apiTemplates, err := apistemplate.LoadAPITemplates(*apiTemplateFile)
	if err != nil {
		// 这里可以选择报错退出，或者只是打印警告（如果只有 UI 测试）
		fmt.Printf("⚠️  API 模板加载失败 (如果是纯 UI 测试请忽略): %v\n", err)
		apiTemplates = make(apistemplate.APITemplates) // 空 map 防止空指针
	}

	// 启动 Playwright 浏览器
	page := browseTemplate.StartWithConfig(cfg)
	// 根据配置决定是否在测试结束后关闭浏览器
	if !cfg.KeepBrowserOpen {
		defer browseTemplate.Stop()
	}

	// 创建测试运行器
	testRunner := runner.NewRunner(page, apiTemplates)

	// 执行测试套件
	fmt.Printf("📂 加载测试文件: %s\n", *testFile)
	err = testRunner.RunTestSuiteFromFile(*testFile)
	if err != nil {
		fmt.Printf("❌ 测试执行失败: %v\n", err)
		browseTemplate.TakeErrorScreenshot(page)
		if cfg.KeepBrowserOpen {
			waitForUserInput("浏览器将保持打开状态，请按 Enter 键退出程序")
		} else {
			os.Exit(1)
		}
		return
	}

	fmt.Println("✅ 所有测试用例执行完成")
	if cfg.KeepBrowserOpen {
		waitForUserInput("浏览器将保持打开状态，请按 Enter 键退出程序")
	}
}

// waitForUserInput 等待用户输入或信号，保持程序运行
func waitForUserInput(message string) {
	fmt.Println("⚠️  " + message)

	// 设置信号处理，捕获 Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// 启动 goroutine 等待用户输入
	inputChan := make(chan bool, 1)
	go func() {
		reader := bufio.NewReader(os.Stdin)
		reader.ReadString('\n')
		inputChan <- true
	}()

	// 等待用户输入或信号
	select {
	case <-inputChan:
		fmt.Println("\n收到用户输入，程序退出")
	case <-sigChan:
		fmt.Println("\n收到退出信号，程序退出")
	}
}

// showHelp 显示帮助信息
func showHelp() {
	fmt.Println("自动化测试框架")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  go run main.go [选项]")
	fmt.Println()
	fmt.Println("选项:")
	fmt.Println("  -c string    配置文件路径 (默认: config.yaml)")
	fmt.Println("  -f string    测试用例文件路径 (默认: testcase/login_example.json)")
	fmt.Println("  -h           显示帮助信息")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  go run main.go -c config.yaml -f testcase/login_example.json")
	fmt.Println("  go run main.go -f testcase/my_test.json")
	fmt.Println("  go run main.go -c my_config.yaml")
}
