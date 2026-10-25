package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

const defaultTimeout = 15

type ResponseData struct {
	Status string
	Head   http.Header
	Body   []byte
}

func GetResponse(ctx context.Context, url string, ch chan<- ResponseData, wg *sync.WaitGroup) {
	defer wg.Done()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	ch <- ResponseData{
		Status: resp.Proto + " " + resp.Status,
		Head:   resp.Header,
		Body:   body,
	}
}

func run() int {
	var timeoutSec int
	flag.IntVar(&timeoutSec, "timeout", defaultTimeout, "таймаут в `SECONDS`")
	flag.IntVar(&timeoutSec, "t", defaultTimeout, "таймаут в `SECONDS` (сокращение)")
	flag.CommandLine.SetOutput(os.Stdout)
	flag.Parse()

	urls := flag.Args()
	if len(urls) == 0 {
		fmt.Fprintln(os.Stderr, "Error: no url provided")
		return 1
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
	defer cancel()

	ch := make(chan ResponseData, len(urls))
	var wg sync.WaitGroup
	for _, url := range urls {
		wg.Add(1)
		go GetResponse(ctx, url, ch, &wg)
	}
	go func() {
		wg.Wait()
		close(ch)
	}()

	select {
	case data, ok := <-ch:
		if ok {
			fmt.Println(data.Status)
			for k, v := range data.Head {
				fmt.Printf("%s: %v\n", k, v)
			}
			fmt.Println()
			fmt.Println(string(data.Body))
			return 0
		}
		if ctx.Err() != nil {
			fmt.Fprintln(os.Stderr, "Error: timeout")
			return 228
		}
		fmt.Fprintln(os.Stderr, "Error: all requests failed")
		return 1
	case <-ctx.Done():
		fmt.Fprintln(os.Stderr, "Error: timeout")
		return 228
	}
}

func main() {
	os.Exit(run())
}
