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
			if nArg := scanCmd.NArg(); nArg > 0 {
				fmt.Print("错误: 不支持多余参数:")
				for i := 0; i < nArg; i++ {
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
			isDir := info.IsDir()
			if !isDir {
				fmt.Println("目标不是目录")
				return
			}

			result, err := scanDirectory(cleanPath)

			if err != nil {
				fmt.Printf("错误信息为:%v\n", err)
				return
			}

			length := len(result.Files)
			for i := 0; i < length; i++ {
				fmt.Printf("%v %v bytes\n", result.Files[i].Path, result.Files[i].Size)
			}
			fmt.Printf("文件数量: %v\n", length)
			fmt.Printf("总大小: %v bytes\n", result.TotalSize)

		} else {
			fmt.Println("请输入正确的功能参数！")
		}

		return
	}

	fmt.Println("用法: GoDrop scan")

}
