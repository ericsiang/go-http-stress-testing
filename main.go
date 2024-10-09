package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

const (
	targetURL      = "" //測試目標 url
	totalRequests  = 5000                      // 總請求數
	concurrency    = 1000                       // 總併發數
	requestsPerSec = 1000                       // 每秒發送的請求數
)

// 統計數據
var (
	successCount   int
	failureCount   int
	totalTime      time.Duration
	mutex          sync.Mutex
	failureMessage map[int]int
)

// 發送HTTP請求的函式
func sendRequest(i int, failureMessage map[int]int) {
	// log.Println("i :",i)
	start := time.Now()
	resp, err := http.Get(targetURL)
	duration := time.Since(start)

	mutex.Lock()
	totalTime += duration
	if err != nil {
		log.Fatal(err)
		failureMessage[i] = -1 // 使用 -1 表示請求失敗
		failureCount++
	} else {
		defer resp.Body.Close() // 確保關閉 response body
		if resp.StatusCode != http.StatusOK {
			failureMessage[i] = resp.StatusCode
			failureCount++
		} else {
			successCount++
		}
	}

	mutex.Unlock()
}

func main() {
	start := time.Now()
	failureMessage := make(map[int]int)

	// WaitGroup 等待所有請求完成
	var wg sync.WaitGroup
	requestChan := make(chan struct{}, concurrency)

	ticker := time.NewTicker(1 * time.Second) // 每秒發送請求的間隔
	defer ticker.Stop()

	requestsSent := 0 // 用於跟蹤已發送的請求數量
	wg.Add(1)
	go func() {
		defer wg.Done()
		for range ticker.C {
			tickerTotalTime := time.Since(start)
			log.Println(tickerTotalTime)
			for i := 0; i < requestsPerSec && requestsSent < totalRequests; i++ {
				requestChan <- struct{}{} // 將請求加入併發控制管道

				wg.Add(1)
				go func(i int) {
					defer wg.Done()
					sendRequest(i, failureMessage)
					<-requestChan // 完成請求後釋放chain
				}(requestsSent)
				requestsSent++
			}

			// 如果所有請求已發送完畢，停止發送
			if requestsSent >= totalRequests {
				break
			}
		}
	}()
	wg.Wait()
	endTotalTime := time.Since(start)

	// 輸出結果
	fmt.Printf("Total Requests: %d\n", totalRequests)
	fmt.Printf("Success Count: %d\n", successCount)
	fmt.Printf("Failure Count: %d\n", failureCount)
	fmt.Printf("Total Time: %v\n", endTotalTime)
	fmt.Printf("Average Time per Request: %v\n", totalTime/time.Duration(successCount+failureCount))
	fmt.Println("Faile Message :", failureMessage)
}
