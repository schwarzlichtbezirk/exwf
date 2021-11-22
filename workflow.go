package exwf

import (
	"context"
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
	// context to indicate about service shutdown
	exitctx context.Context
	exitfn  context.CancelFunc
	// wait group for all service goroutines
	exitwg sync.WaitGroup
)

var (
	// Threads - list of all readed chains.
	Threads []*Chain
	// client - HTTP client for all requests
	client = &http.Client{}
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
		// check on exit signal during running
		select {
		case <-exitctx.Done():
			err = io.EOF
			return
		default:
		}
	}
	return
}

// WaitInterrupt returns shutdown signal was recivied and cancels some context.
func WaitInterrupt(cancel context.CancelFunc) {
	// Make exit signal on function exit.
	defer cancel()

	var sigint = make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C) or SIGTERM (Ctrl+/)
	// SIGKILL, SIGQUIT will not be caught.
	signal.Notify(sigint, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	// Block until we receive our signal.
	<-sigint
	log.Println("shutting down by break")
}

// Init performs program data initialization.
func Init() {
	// create context and wait the break
	exitctx, exitfn = context.WithCancel(context.Background())
	go WaitInterrupt(exitfn)

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

	log.Printf("run")
	var start = time.Now()
	exitwg.Wait()
	var rundur = time.Since(start)
	log.Printf("running time: %s, request number: %d\n", rundur.String(), ReqCount)
}

// The End.
