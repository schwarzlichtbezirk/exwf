package exwf

import (
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

var (
	// channel to indicate about server shutdown
	exitchan chan struct{}
	// wait group for all server goroutines
	exitwg sync.WaitGroup
)

var (
	// Threads - list of all readed chains.
	Threads []*Chain
	// client - HTTP client for all requests
	client = &http.Client{}
	// timeStart - iterations start time
	timeStart time.Time
	// timeEnd - iterations end time
	timeEnd time.Time
)

// Iteration makes one iteration calls along the chain.
func (c *Chain) Iteration() (err error) {
	for i, ent := range c.Entries {
		// make request
		var req *http.Request
		if req, err = http.NewRequest(ent.Method, ent.URL, strings.NewReader(ent.Data)); err != nil {
			return
		}
		if len(ent.Token) > 0 {
			req.Header.Set("Authorization", "Bearer "+ent.Token)
		}
		// do request
		if ent.WaitRpl {
			var resp *http.Response
			if resp, err = client.Do(req); err != nil {
				log.Printf("request #%d failed", i)
				return
			}
			atomic.AddInt64(&ReqCount, 1)
			_ = resp
		} else {
			go client.Do(req)
			atomic.AddInt64(&ReqCount, 1)
		}
		// wait some delay if it has
		if ent.DelayMin > 0 || ent.DelayMax > 0 {
			var add time.Duration
			if ent.DelayMax > 0 {
				add = time.Duration(rand.Int63n(int64(ent.DelayMax - ent.DelayMin)))
			}
			time.Sleep(ent.DelayMin + add)
		}
		// check exit channel
		select {
		case <-exitchan:
			err = io.EOF
			return
		default:
		}
	}
	return
}

// Init performs program data initialization.
func Init() {
	// inits exit channel
	exitchan = make(chan struct{})

	var err error
	if err = ReadConfig(); err != nil {
		log.Fatal(err)
	}
	if len(Threads) == 0 {
		log.Fatal("no any chain of requests entries was readed")
	}
}

// Run launches program threads.
func Run() {
	timeStart = time.Now()
	for i, chain := range Threads {
		var i = i
		var chain = chain

		exitwg.Add(1)
		go func() {
			defer exitwg.Done()

			for i := 0; i != chain.Repeats; i++ {
				if err := chain.Iteration(); err != nil {
					log.Printf("iteration #%d: %v\n", i+1, err)
					return
				}
			}

			log.Printf("thread %d complete\n", i)
		}()
	}
}

// WaitBreak blocks goroutine until Ctrl+C will be pressed.
func WaitBreak() {
	var sigint = make(chan os.Signal)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C) or SIGTERM (Ctrl+/)
	// SIGKILL, SIGQUIT will not be caught.
	signal.Notify(sigint, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive our signal.
	<-sigint
	// Make exit signal.
	close(exitchan)
}

// WaitExit waits until all server threads will be stopped.
func WaitExit() {
	exitwg.Wait()
}

// Shutdown performs graceful network shutdown.
func Shutdown() {
	timeEnd = time.Now()
	log.Printf("running time: %v, request number: %d\n",
		timeEnd.Sub(timeStart), ReqCount)
}

// The End.
