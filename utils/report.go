package utils

import (
	"fmt"
)

func GenerateReport(name string, success bool) {
	status := "失败 ❌"
	if success {
		status = "成功 ✅"
	}
	fmt.Printf("测试用例 [%s] 执行结果: %s\n", name, status)
}
