package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type syncSummary struct {
	Files int
	Bytes int64
}

type syncJobs struct {
	Entry      FileEntry
	OutputPath string
}

type syncOutcome struct {
	Path  string
	Bytes int64
	Err   error
}

func closer(wg *sync.WaitGroup, ch chan syncOutcome) {

	wg.Wait()

	close(ch)

}

func employer(ctx context.Context, result ScanResult, ch chan syncJobs, outputDir string) {

	defer close(ch)

	for _, file := range result.Files {

		select {
		case ch <- syncJobs{
			Entry:      file,
			OutputPath: filepath.Join(outputDir, file.Path),
		}:
		case <-ctx.Done():
			return
		}

	}

}

func worker(wg *sync.WaitGroup, ctx context.Context, downloadURL string, chJobs chan syncJobs, chOut chan syncOutcome) {

	defer wg.Done()

	for job := range chJobs {

		numBytes, err := downloadFile(ctx, downloadURL, job.Entry.Path, job.OutputPath)

		if err != nil {

			chOut <- syncOutcome{
				Path: job.Entry.Path,
				Err:  fmt.Errorf("%v文件下载失败, 错误信息:%w", job.Entry.Path, err),
			}

			return

		}

		chOut <- syncOutcome{
			Path:  job.Entry.Path,
			Bytes: numBytes,
			Err:   nil,
		}

	}

}

func checkIfPathValid(path string) error {

	if path == "" {
		return fmt.Errorf("存在空字符串索引")
	} else if path == "." {
		return fmt.Errorf("存在.索引")
	} else if !filepath.IsLocal(path) {
		return fmt.Errorf("IsLocal未通过")
	} else if filepath.Clean(path) != path {
		return fmt.Errorf("filepath.Clean() != path")
	}
	return nil

}

func syncFiles(fctx context.Context, result ScanResult, downloadURL, outputDir string, workers int) (syncSummary, error) {

	ctx, cancel := context.WithCancel(fctx)

	defer cancel()

	ret := syncSummary{
		Files: 0,
		Bytes: 0,
	}

	for _, file := range result.Files {

		err := checkIfPathValid(file.Path)

		if err != nil {

			return ret, fmt.Errorf("存在非法索引, 错误信息为:%w", err)

		}

		destinationPath := filepath.Join(outputDir, file.Path)
		parentDir := filepath.Dir(destinationPath)

		if err = os.MkdirAll(parentDir, 0755); err != nil {

			return ret, fmt.Errorf("客户端创建接收文件目录失败, 错误信息为%w", err)

		}

	}

	var wg sync.WaitGroup

	chJobs := make(chan syncJobs)
	chOutcomes := make(chan syncOutcome)

	go employer(ctx, result, chJobs, outputDir)

	for i := 0; i < workers; i++ {

		wg.Add(1)
		go worker(&wg, ctx, downloadURL, chJobs, chOutcomes)

	}

	go closer(&wg, chOutcomes)

	var firstErr error

	for outcome := range chOutcomes {

		if outcome.Err != nil {

			if firstErr == nil {

				firstErr = outcome.Err

				cancel()

			}

			continue

		}

		ret.Files++
		ret.Bytes += outcome.Bytes

	}

	if firstErr != nil { //确实很妙
		return ret, firstErr
	}

	if err := fctx.Err(); err != nil {
		return ret, err
	}

	return ret, nil

}
