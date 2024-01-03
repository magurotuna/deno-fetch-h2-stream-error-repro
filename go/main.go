package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
)

func request(client *http.Client, id int, wg *sync.WaitGroup) error {
	defer wg.Done()

	req, err := http.NewRequest("GET", "https://deno-fetch-h2-repro.dev", nil)
	if err != nil {
		fmt.Printf("❌ %d %s\n", id, err)
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ %d %s\n", id, err)
		return err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ %d %s\n", id, err)
		return err
	}
	fmt.Printf("✅ %d %s\n", id, string(b))
	return nil
}

func getConcurrency() int {
	v := os.Getenv("CONCURRENCY")
	if v == "" {
		return 1
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 1
	}
	if n <= 0 {
		return 1
	}
	return n
}

func main() {
	var wg sync.WaitGroup

	client := &http.Client{
		Transport: &http.Transport{
			ForceAttemptHTTP2: true,
			TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		},
	}

	for i := 0; i < getConcurrency(); i++ {
		wg.Add(1)
		go request(client, i, &wg)
	}

	wg.Wait()
	fmt.Println("done")
}
