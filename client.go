package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

func fetchIndex(rawURL string) (ScanResult, error) {

	var result ScanResult

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequest("GET", rawURL, nil) //第三个参数由于本次不需要请求体，省略

	if err != nil {

		return result, err

	}

	resp, err := client.Do(req)

	if err != nil {

		return result, err

	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {

		if Derr := json.NewDecoder(resp.Body).Decode(&result); Derr != nil {

			return result, Derr

		}

		return result, nil

	} else {

		return result, fmt.Errorf("bad status %v", resp.Status)

	}

}

func downloadFile(rawURL, remotePath, outputPath string) (int64, error) {

	var n int64 = 0

	u, err := url.Parse(rawURL)

	if err != nil {

		return n, fmt.Errorf("解析rawURL出错, 错误信息:%w", err)

	}

	q := u.Query()

	q.Set("path", remotePath)

	u.RawQuery = q.Encode()

	finalURL := u.String()

	req, err := http.NewRequest("GET", finalURL, nil)

	if err != nil {

		return n, fmt.Errorf("请求创建失败, 错误信息为:%w\n", err)

	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)

	if err != nil {

		return n, fmt.Errorf("Do失败, 错误信息为:%w\n", err)

	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {

		file, err := os.Create(outputPath)

		if err != nil {

			return n, fmt.Errorf("文件创建失败, 错误信息为:%w\n", err)

		}

		defer file.Close()

		n, err = io.Copy(file, resp.Body)

		if err != nil {

			return n, fmt.Errorf("传输失败, 错误信息为:%w\n", err)

		}

		return n, nil

	} else {

		return n, fmt.Errorf("下载失败，服务端状态: %s", resp.Status)

	}

}
