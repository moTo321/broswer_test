package main

import (
	"autotest/config"
	"autotest/driver"
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
		configFile = flag.String("c", "config.yaml", "配置文件路径 (默认: config.yaml)")
		testFile   = flag.String("f", "testcase/login_example.json", "测试用例文件路径 (默认: testcase/login_example.json)")
		help       = flag.Bool("h", false, "显示帮助信息")
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
	fmt.Printf("📋 加载配置文件: %s\n", *configFile)
	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		fmt.Printf("⚠️  配置加载失败，使用默认配置: %v\n", err)
		cfg = config.DefaultConfig()
	} else {
		fmt.Printf("   - 浏览器: %s\n", cfg.Browser)
		fmt.Printf("   - 无头模式: %t\n", cfg.Headless)
		fmt.Printf("   - 超时时间: %dms\n", cfg.Timeout)
	}

	// 启动 Playwright 浏览器
	page := driver.StartWithConfig(cfg)
	// 根据配置决定是否在测试结束后关闭浏览器
	if !cfg.KeepBrowserOpen {
		defer driver.Stop()
	}

	// 创建测试运行器
	testRunner := runner.NewRunner(page)

	// 执行测试套件
	fmt.Printf("📂 加载测试文件: %s\n", *testFile)
	err = testRunner.RunTestSuiteFromFile(*testFile)
	if err != nil {
		fmt.Printf("❌ 测试执行失败: %v\n", err)
		driver.TakeErrorScreenshot(page)
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
