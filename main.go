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

// variables for goroutine synchronization and communication through shared memory
var (
	// using a map[string]struct{} as a set is more efficient
	// because an empty struct{} occupies 0 bytes in memory
	queuedLinks  map[string]struct{}
	crawledLinks map[string]struct{}
	mutexLock    *sync.RWMutex
	semaphores   *semaphore.Weighted
)

func main() {
	// parse command line arguments
	parseArguments()

	// process interrupt signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// context to propagate cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ticker := time.NewTicker(time.Duration(interval) * time.Millisecond)

	// mutual exclusion lock to safeguard and synchronize critical sections
	mutexLock = &sync.RWMutex{}

	// counting semaphores to control the number of active threads.
	semaphores = semaphore.NewWeighted(concurrency)

	// wait group to synchronize waiting on threads
	waitGroup := sync.WaitGroup{}

	crawledLinks = make(map[string]struct{})

	queuedLinks = make(map[string]struct{})
	queuedLinks[target] = struct{}{}

	for {
		select {
		case <-quit:
			fmt.Printf("\ncrawler has been interrupted\n")
			ticker.Stop()
			cancel()
			return
		case <-ticker.C:

			if len(queuedLinks) == 0 {
				fmt.Printf("\ncrawler has completed\n")
				return
			}

			// get the next URL to crawl
			var link string
			mutexLock.RLock()
			for queuedLink := range queuedLinks {
				if _, crawled := crawledLinks[queuedLink]; crawled {
					fmt.Printf("skipping the link '%s' which has already been crawled\n", queuedLink)

					// remove the crawled link from the queued set and add it to the
					// crawled set within a synchronized operation for thread safety
					evict(queuedLinks, crawledLinks, link, mutexLock)

					continue
				}

				link = queuedLink
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
					// remove the crawledLinks link from the state
					evict(queuedLinks, crawledLinks, link, mutexLock)

					semaphores.Release(1)
				}()

				// set this link as crawledLinks so that it is not duplicated
				markAsCrawled(crawledLinks, link, mutexLock)

				childrenLinks, err := crawl(link, directory, overwrite, target)
				if err != nil {
					fmt.Printf("an error has occurred while crawling '%s': '%v'\n", link, err)

					return
				}

				for _, childLink := range childrenLinks {
					// add child link to queued set
					queueLinkToCrawl(queuedLinks, crawledLinks, childLink, mutexLock)
				}

			}(ctx, link)
		}

		waitGroup.Wait()
	}
}

func parseArguments() {
	flag.StringVar(&target, "url", "", "target url to crawl")
	flag.StringVar(&directory, "dir", "", "target directory to save crawled files")
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
}

// evict removes a link from tracking state
func evict(queue map[string]struct{}, crawled map[string]struct{}, link string, lock *sync.RWMutex) {
	lock.Lock()

	// delete from active set
	delete(queue, link)

	// add to crawled set which helps avoid unnecessary
	// adding of crawled links to the queuedLinks set
	crawled[link] = struct{}{}

	lock.Unlock()
}

// markAsCrawled sets a link as crawled so that it is not duplicated in other goroutines
func markAsCrawled(crawled map[string]struct{}, link string, lock *sync.RWMutex) {
	lock.Lock()
	crawled[link] = struct{}{}
	lock.Unlock()
}

// queueLinkToCrawl adds a link to the queued set if not already crawled
func queueLinkToCrawl(queue map[string]struct{}, crawled map[string]struct{}, link string, lock *sync.RWMutex) {
	// we first check whether the link is already crawled
	var isCrawled bool
	mutexLock.RLock()
	_, isCrawled = crawled[link]
	mutexLock.RUnlock()

	// if it is not, we queue it
	if !isCrawled {
		lock.Lock()
		queue[link] = struct{}{}
		lock.Unlock()
	}
}
