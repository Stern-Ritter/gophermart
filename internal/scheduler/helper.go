package scheduler

import (
	"context"
	"sync"
	"time"

	"gopkg.in/h2non/gentleman.v2"
)

func sendGetRequest(client *gentleman.Client, endpoint string) (*gentleman.Response, error) {
	req := client.Request()
	req.Method("GET")
	req.Path(endpoint)

	return req.Send()
}

func setInterval(ctx context.Context, wg *sync.WaitGroup, task func(), interval time.Duration) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				wg.Done()
				return
			default:
				task()
				time.Sleep(interval)
			}
		}
	}()
}
