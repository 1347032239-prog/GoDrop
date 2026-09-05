package main

import (
	"encoding/json"
	"fmt"
	"net/http"
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
