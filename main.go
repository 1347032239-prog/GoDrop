package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

func runScan(args []string) {
	scanCmd := flag.NewFlagSet("scan", flag.ContinueOnError)
	dirPtr := scanCmd.String("dir", "", "目标路径")
	jsonPtr := scanCmd.String("out", "", "json保存路径")
	err := scanCmd.Parse(args)
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
}

func runShow(args []string) {

	showCmd := flag.NewFlagSet("show", flag.ContinueOnError)
	indexPtr := showCmd.String("index", "", "要展示的json保存路径")
	err := showCmd.Parse(args)
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

}

func runServe(args []string) {

	serveCmd := flag.NewFlagSet("serve", flag.ContinueOnError)
	indexPtr := serveCmd.String("index", "", "要提供的json路径")
	addrPtr := serveCmd.String("addr", "127.0.0.1:8080", "服务端socket")
	dirPtr := serveCmd.String("dir", "", "可下载文件路径 ")

	err := serveCmd.Parse(args)

	if err != nil {

		fmt.Printf("参数解析失败, 错误信息:%v\n", err)

		return

	}

	if *indexPtr == "" {

		fmt.Println("未获取到路径")

		return

	}

	if *dirPtr == "" {

		fmt.Println("未提供可下载文件路径")

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

	root, err := os.OpenRoot(*dirPtr)

	if err != nil {

		fmt.Printf("打开文件目录失败, 错误信息:%v\n", err)

		return

	}

	defer root.Close()

	mux := newHTTPHandler(result, root)

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))

	slog.Info("服务正在启动", "addr", *addrPtr)

	signalCtx, stopSignal := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)

	defer stopSignal()

	server := &http.Server{
		Addr:              *addrPtr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErrCh := make(chan error, 1)

	go func() {

		serverErrCh <- server.ListenAndServe()

	}()

	select {

	case serveErr := <-serverErrCh:

		slog.Error("监听服务异常退出", "addr", *addrPtr, "error", serveErr)

		return

	case <-signalCtx.Done():

		slog.Info("收到退出信号", "addr", *addrPtr)
		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelShutdown()

		shutdownErr := server.Shutdown(shutdownCtx)

		if shutdownErr != nil {

			slog.Error("优雅关闭失败", "addr", *addrPtr, "error", shutdownErr)

			closeErr := server.Close()

			if closeErr != nil {

				slog.Error("强制关闭失败", "error", closeErr)

			}

		}

		serveErr := <-serverErrCh

		if !errors.Is(serveErr, http.ErrServerClosed) {

			slog.Error("收到信号后监听服务出现异常", "addr", *addrPtr, "error", serveErr)

			return

		}

		if shutdownErr != nil {

			return

		}

		slog.Info("服务安全关闭", "addr", *addrPtr)

	}

}

func runFetch(args []string) {

	fetchCmd := flag.NewFlagSet("fetch", flag.ContinueOnError)
	urlPtr := fetchCmd.String("url", "", "url")
	err := fetchCmd.Parse(args)

	if err != nil {

		fmt.Printf("参数解析失败, 错误信息:%v\n", err)

		return

	}

	if *urlPtr == "" {

		fmt.Println("url获取失败")

		return

	}

	if nArg := fetchCmd.NArg(); nArg > 0 {
		fmt.Print("错误: 不支持多余参数:")
		for i := 0; i < nArg; i++ {
			fmt.Printf("%v ", fetchCmd.Arg(i))
		}
		return
	}

	result, err := fetchIndex(*urlPtr)

	if err != nil {

		fmt.Printf("错误信息:%v\n", err)

		return

	}

	printScanResult(result)

}

func runDownload(args []string) {

	downloadCmd := flag.NewFlagSet("download", flag.ContinueOnError)
	urlPtr := downloadCmd.String("url", "", "url")
	pathPtr := downloadCmd.String("path", "", "服务端共享目录中的相对地址")
	outPtr := downloadCmd.String("out", "", "客户端保存地址")
	timeoutPtr := downloadCmd.Duration("timeout", 30*time.Second, "下载超时时间")
	err := downloadCmd.Parse(args)

	if err != nil {

		fmt.Printf("参数解析失败, 错误信息:%v\n", err)

		return

	}

	if *urlPtr == "" {

		fmt.Println("url获取失败")

		return

	}

	if *pathPtr == "" {

		fmt.Println("未填目标资源在共享目录下的相对地址")

		return

	}

	if *outPtr == "" {

		fmt.Println("out字段必填本地保存地址")

		return

	}

	if *timeoutPtr <= 0 {

		fmt.Println("超时时间小于或等于0, 错误!")

		return

	}

	if nArg := downloadCmd.NArg(); nArg > 0 {
		fmt.Print("错误: 不支持多余参数:")
		for i := 0; i < nArg; i++ {
			fmt.Printf("%v ", downloadCmd.Arg(i))
		}
		return
	}

	signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	defer stop()

	ctx, cancel := context.WithTimeout(signalCtx, *timeoutPtr)

	defer cancel()

	n, err := downloadFile(ctx, *urlPtr, *pathPtr, *outPtr)

	if err != nil {

		fmt.Printf("文件下载失败, 错误信息:%v\n", err)

		return

	}

	fmt.Printf("下载文件成功, 文件大小:%v bytes\n", n)

}

func runSync(args []string) {
	syncCmd := flag.NewFlagSet("sync", flag.ContinueOnError)
	indexURLPtr := syncCmd.String("index-url", "", "indexURL")
	downloadPtr := syncCmd.String("download-url", "", "downloadURL")
	outPtr := syncCmd.String("out", "", "客户端下载的文件保存在哪")
	workerPtr := syncCmd.Int("workers", 4, "开启几个下载协程")
	timeoutPtr := syncCmd.Duration("timeout", 2*time.Minute, "timeout")
	err := syncCmd.Parse(args)

	if err != nil {

		fmt.Println("Parse failed")

		return

	}

	if *indexURLPtr == "" {

		fmt.Println("未成功获取索引URL")

		return

	}

	if *downloadPtr == "" {

		fmt.Println("未成功获取下载URL")

		return

	}

	if *outPtr == "" {

		fmt.Println("未成功获取客户端本地保存路径")

		return

	}

	if *workerPtr < 1 || *workerPtr > 32 {

		fmt.Println("允许协程数范围:1-32")

		return

	}

	if *timeoutPtr <= 0 {

		fmt.Println("timeout必须大于零!")

		return

	}

	if nArg := syncCmd.NArg(); nArg > 0 {

		fmt.Print("错误: 不支持多余参数:")
		for i := 0; i < nArg; i++ {
			fmt.Printf("%v ", syncCmd.Arg(i))
		}
		return

	}

	parent := context.Background()

	ctxTimeout, stop := context.WithTimeout(parent, *timeoutPtr)

	defer stop()

	ctx, cancel := signal.NotifyContext(ctxTimeout, syscall.SIGTERM, os.Interrupt) //os.Interrupt是跨平台中断信号

	defer cancel()

	result, err := fetchIndexWithContext(ctx, *indexURLPtr)

	if err != nil {

		fmt.Println("索引获取失败")

		return

	}

	summary, err := syncFiles(ctx, result, *downloadPtr, *outPtr, *workerPtr)

	if err != nil {

		fmt.Printf("同步失败, 错误信息:%v\n", err)

		return

	}

	fmt.Printf("同步完成: %v files, %v bytes\n", summary.Files, summary.Bytes)

}

func main() {

	if len(os.Args) <= 1 {

		fmt.Println("支持用法:scan , show, serve, fetch, download, sync")

		return
	}

	args := os.Args[2:]

	switch os.Args[1] {
	case "scan":
		runScan(args)
	case "show":
		runShow(args)
	case "serve":
		runServe(args)
	case "fetch":
		runFetch(args)
	case "download":
		runDownload(args)
	case "sync":
		runSync(args)
	default:
		fmt.Println("请输入正确的功能参数！")
	}

}
