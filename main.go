package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {

	argCount := len(os.Args)
	if argCount > 1 {
		if os.Args[1] == "scan" {
			scanCmd := flag.NewFlagSet("scan", flag.ContinueOnError)
			dirPtr := scanCmd.String("dir", "", "目标路径")
			jsonPtr := scanCmd.String("out", "", "json保存路径")
			err := scanCmd.Parse(os.Args[2:])
			if err != nil {
				fmt.Printf("参数解析失败!可能传入了非法flag, 错误信息:%v\n", err)
				return
			}
			if *dirPtr == "" {
				fmt.Println("必须使用 -dir 指定扫描目录，可能未传入路径")
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

			printScanResult(result)

			if *jsonPtr != "" {
				writeIndexErr := writeIndex(filepath.Clean(*jsonPtr), result)
				if writeIndexErr != nil {
					fmt.Printf("json写入失败 , 错误信息:%v\n", writeIndexErr)
					return
				}

				fmt.Printf("索引已写入:%v\n", filepath.Clean(*jsonPtr))
			}

		} else if os.Args[1] == "show" {
			showCmd := flag.NewFlagSet("show", flag.ContinueOnError)
			indexPtr := showCmd.String("index", "", "要展示的json保存路径")
			err := showCmd.Parse(os.Args[2:])
			if err != nil {
				fmt.Printf("参数解析失败!可能传入了非法flag, 错误信息:%v\n", err)
				return
			}

			if *indexPtr == "" {
				fmt.Println("必须使用 -index 指定展示路径，可能未传入路径")
				return
			}

			if nArg := showCmd.NArg(); nArg > 0 {
				fmt.Print("错误: 不支持多余参数:")
				for i := 0; i < nArg; i++ {
					fmt.Printf("%v ", showCmd.Arg(i))
				}
				return
			}

			cleanPath := filepath.Clean(*indexPtr)
			result, err := readIndex(cleanPath)
			if err != nil {
				fmt.Printf("读json反序列化失败, 错误信息:%v", err)
				return
			}

			printScanResult(result)

		} else if os.Args[1] == "serve" {

			serveCmd := flag.NewFlagSet("serve", flag.ContinueOnError)
			indexPtr := serveCmd.String("index", "", "要提供的json路径")
			addrPtr := serveCmd.String("addr", "127.0.0.1:8080", "服务端socket")
			err := serveCmd.Parse(os.Args[2:])

			if err != nil {

				fmt.Printf("参数解析失败, 错误信息:%v\n", err)

				return

			}

			if *indexPtr == "" {

				fmt.Println("未获取到路径")

				return

			}

			if nArg := serveCmd.NArg(); nArg > 0 {
				fmt.Print("错误: 不支持多余参数:")
				for i := 0; i < nArg; i++ {
					fmt.Printf("%v ", serveCmd.Arg(i))
				}
				return
			}

			result, err := readIndex(*indexPtr)

			if err != nil {

				fmt.Printf("读取json失败, 错误信息:%v\n", err)

				return

			}

			mux := newHTTPHandler(result)

			fmt.Printf("服务启动于 http://%v\n", *addrPtr)

			if err := http.ListenAndServe(*addrPtr, mux); err != nil {
				log.Fatalf("Server failed to start: %v", err)
			}

		} else {
			fmt.Println("请输入正确的功能参数！")
		}

		return
	}

	fmt.Println("支持用法:scan , show, serve")

}
