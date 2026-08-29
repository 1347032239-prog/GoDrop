package main

import "fmt"
import "os"

func main() {

	argCount := len(os.Args)
	if argCount > 1 {
		if os.Args[1] == "scan" {
			fmt.Println("准备扫描文件")
		}
		else {
			fmt.Println("请输入正确的参数！")
		}
	}

	fmt.Println("用法: GoDrop scan")

}
