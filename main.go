package main

import (
	"context"
	"flag"
	"fmt"
	"golang.org/x/sync/semaphore"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// variables command line arguments
var (
	target      string
	directory   string
	overwrite   bool
	concurrency int64
	interval    int64
)

// variables for goroutine synchronization
// and communication through shared memory
var (
	crawledState map[string]bool
	mutexLock    *sync.RWMutex
	semaphores   *semaphore.Weighted
)

func main() {
	flag.StringVar(&target, "url", "", "target url to crawl")
	flag.StringVar(&directory, "dir", "", "target directory to save crawledState files")
	flag.BoolVar(&overwrite, "overwrite", false, "overwrite download of files")
	flag.Int64Var(&concurrency, "concurrency", 10, "number of concurrent crawlers")
	flag.Int64Var(&interval, "interval", 1000, "the number of milliseconds to wait between crawls")
	flag.Parse()

	if len(target) == 0 {
		log.Fatal("a target url has not been defined with the -url option")
	}

	if len(directory) == 0 {
		log.Fatal("a destination directory has not been defined with the -dir option")
	}

	// process interrupt signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// context to propagate cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ticker := time.NewTicker(time.Duration(interval) * time.Millisecond)

	// mutual exclusion lock to safeguard
	// and synchronize critical sections
	mutexLock = &sync.RWMutex{}

	// counting semaphores to control the number of active threads.
	semaphores = semaphore.NewWeighted(concurrency)

	// wait group to synchronize waiting on threads
	waitGroup := sync.WaitGroup{}

	crawledState = make(map[string]bool)
	crawledState[target] = false

	for {
		select {
		case <-quit:
			fmt.Printf("\ncrawler has been interrupted\n")
			ticker.Stop()
			cancel()
			return
		case <-ticker.C:

			if len(crawledState) == 0 {
				fmt.Printf("\ncrawler has completed\n")
				return
			}

			// get the next URL to crawl
			var link string
			mutexLock.RLock()
			for url, crawled := range crawledState {
				if crawled {
					fmt.Printf("skipping the link '%s' which has already been crawled\n", url)

					// remove the crawled link from the state
					evict(crawledState, link, mutexLock)

					continue
				}

				link = url
			}
			mutexLock.RUnlock()

			// fire-off a goroutine for this link
			waitGroup.Add(1)

			go func(ctx context.Context, link string) {
				defer waitGroup.Done()

				// try to acquire a semaphore token for each child link
				if err := semaphores.Acquire(ctx, 1); err != nil {
					fmt.Printf("failed to acquire semaphore: %s\n", err)

					return
				}

				defer func() {
					// remove the crawled link from the state
					evict(crawledState, link, mutexLock)

					semaphores.Release(1)
				}()

				// set this link as crawledState so that it is
				// not picked up by another thread
				mutexLock.Lock()
				crawledState[link] = true
				mutexLock.Unlock()

				childrenLinks, err := crawl(link, directory, overwrite, target)
				if err != nil {
					fmt.Printf("an error has occurred while crawling '%s': '%v'\n", link, err)

					return
				}

				for _, childLink := range childrenLinks {
					// add child link to state and set its status to false
					mutexLock.Lock()
					crawledState[childLink] = false
					mutexLock.Unlock()
				}

			}(ctx, link)
		}

		waitGroup.Wait()
	}
}

// evict removes a link from tracking state
func evict(state map[string]bool, link string, lock *sync.RWMutex) {
	lock.Lock()
	delete(state, link)
	lock.Unlock()
}
