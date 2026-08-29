package main

import "fmt"
import "os"
import "flag"
import "path/filepath"

func main() {

	argCount := len(os.Args)
	if argCount > 1 {
		if os.Args[1] == "scan" {
			if argCount == 2 {
				fmt.Println("错误：必须使用 -dir 指定扫描目录")
			} else {
				scanCmd := flag.NewFlagSet("scan", flag.ContinueOnError)
				dirPtr := scanCmd.String("dir", ".", "目标路径")
				err := scanCmd.Parse(os.Args[2:])
				if err != nil {

				}
				cleanPath := filepath.Clean(*dirPtr)

				fmt.Printf("准备扫描目录:%v\n", cleanPath)
			}
		} else {
			fmt.Println("请输入正确的功能参数！")
		}

		return
	}

	fmt.Println("用法: GoDrop scan")

}
