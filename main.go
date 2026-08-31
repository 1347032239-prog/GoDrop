package main

import "fmt"
import "os"
import "flag"
import "path/filepath"
import "errors"

func main() {

	argCount := len(os.Args)
	if argCount > 1 {
		if os.Args[1] == "scan" {
			scanCmd := flag.NewFlagSet("scan", flag.ContinueOnError)
			dirPtr := scanCmd.String("dir", "", "目标路径")
			err := scanCmd.Parse(os.Args[2:])
			if err != nil {
				fmt.Printf("参数解析失败!可能传入了非法flag, 错误信息:%v\n", err)
				return
			}
			if *dirPtr == "" {
				fmt.Println("必须使用 -dir 指定扫描目录， 可能未传入路径")
				return
			}
			if NArg := scanCmd.NArg(); NArg > 0 {
				fmt.Print("错误: 不支持多余参数:")
				for i := 0; i < NArg; i++ {
					fmt.Printf("%v ", scanCmd.Arg(i))
				}
				return
			}
			cleanPath := filepath.Clean(*dirPtr)
			info, err := os.Stat(cleanPath)
			if err != nil {
				if y := errors.Is(err, os.ErrNotExist); y {
					fmt.Println("文件不存在")
					return
				}
				fmt.Printf("文件存在仍错误, 错误信息:%v\n", err)
				return
			}
			is_Dir := info.IsDir()
			if !is_Dir {
				fmt.Println("目标不是目录")
				return
			}
			fmt.Printf("准备扫描目录:%v\n", cleanPath)
		} else {
			fmt.Println("请输入正确的功能参数！")
		}

		return
	}

	fmt.Println("用法: GoDrop scan")

}
