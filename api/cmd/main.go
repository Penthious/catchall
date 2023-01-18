package main

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/mailgun/catchall"
	"net/http"
	"sync"
	"time"
)

// I should switch this to be a bounded channel instead of letting it grow as this will end up crashing about 100k
// events as there will be to many in flight go routines and the OS will kill the process. Currently not an issue as
// the event pool is only used for testing to see what an avg load would be.
func main() {
	eventLimit := 10_000
	wg := sync.WaitGroup{}
	bus := catchall.SpawnEventPool()
	defer bus.Close()
	fmt.Println("About to send", humanize.Comma(int64(eventLimit)), "events")
	start := time.Now()
	for i := 0; i < eventLimit; i++ {
		go func() {
			wg.Add(1)
			e := bus.GetEvent()
			req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("http://localhost:7000/v1/events/%s/%s", e.Domain, e.Type), nil)
			if err != nil {
				panic(err)
			}
			http.DefaultClient.Do(req)
			bus.RecycleEvent(e)
			wg.Done()
		}()
	}

	wg.Wait()
	fmt.Println("\nDone in", time.Since(start))
	fmt.Println("avg:", time.Since(start)/time.Duration(eventLimit))
	fmt.Println("Done sending events")
}
