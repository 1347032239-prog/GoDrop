package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
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

		return result, fmt.Errorf("bad status:%v", resp.Status)

	}

}

func fetchIndexWithContext(ctx context.Context, rawURL string) (ScanResult, error) {

	var result ScanResult

	client := &http.Client{}

	req, err := http.NewRequestWithContext(ctx, "GET", rawURL, nil)

	if err != nil {

		return result, fmt.Errorf("请求报文生成失败, 错误信息为%w", err)

	}

	resp, err := client.Do(req)

	if err != nil {

		return result, fmt.Errorf("Do失败, 错误信息为%w", err)

	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {

		if Derr := json.NewDecoder(resp.Body).Decode(&result); Derr != nil {

			return result, Derr

		}

		return result, nil

	} else {

		return result, fmt.Errorf("bad status:%v", resp.Status)

	}

}

func downloadFile(ctx context.Context, rawURL, remotePath, outputPath string) (int64, error) {

	var n int64 = 0

	stage := 1

	u, err := url.Parse(rawURL)

	if err != nil {

		return n, fmt.Errorf("解析rawURL出错, 错误信息:%w", err)

	}

	q := u.Query()

	q.Set("path", remotePath)

	u.RawQuery = q.Encode()

	finalURL := u.String()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, finalURL, nil)

	if err != nil {

		return n, fmt.Errorf("请求创建失败, 错误信息为:%w", err)

	}

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {

		return n, fmt.Errorf("Do失败, 错误信息为:%w", err)

	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {

		sha256Value := resp.Header.Get("X-GoDrop-SHA256")

		digest, err := hex.DecodeString(sha256Value)

		if err != nil {

			return n, fmt.Errorf("解析校验码失败! 错误信息:%w", err)

		}

		if len(digest) != sha256.Size {

			return n, fmt.Errorf("校验码有误")

		}

		hasher := sha256.New()

		tmpfile, err := os.CreateTemp(filepath.Dir(outputPath), "."+filepath.Base(outputPath)+".*.part")

		if err != nil {

			return n, fmt.Errorf("临时文件创建失败, 错误信息为:%w", err)

		}

		multiWriter := io.MultiWriter(hasher, tmpfile)

		tmpPath := tmpfile.Name()

		defer func() {

			switch stage {
			case 1:
				tmpfile.Close()

				os.Remove(tmpPath)
			case 2:
				os.Remove(tmpPath)
			}

		}()

		n, err = io.Copy(multiWriter, resp.Body)

		if err != nil {

			return n, fmt.Errorf("传输失败, 错误信息为:%w", err)

		}

		if ctxErr := ctx.Err(); ctxErr != nil {

			return n, fmt.Errorf("下载被取消: %w", ctxErr)

		}

		cpDigest := hasher.Sum(nil)

		if !bytes.Equal(digest, cpDigest) {

			return n, fmt.Errorf("校验码不相等")

		}

		syncErr := tmpfile.Sync()

		if syncErr != nil {

			return n, fmt.Errorf("Sync failed, syncErr:%w", syncErr)

		}

		stage = 2

		if closeErr := tmpfile.Close(); closeErr != nil {

			return n, fmt.Errorf("closing failed, closeErr:%w", closeErr)

		}

		renameErr := os.Rename(tmpPath, outputPath)

		if renameErr != nil {

			return n, fmt.Errorf("rename failed, renameErr:%w", renameErr)

		}

		stage = 3

		return n, nil

	} else {

		return n, fmt.Errorf("下载失败，服务端状态: %s", resp.Status)

	}

}
