package exwf

import (
	"context"
	"errors"
	"flag"
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

// Command line parameters.
var (
	pdur = flag.Duration("d", 90*time.Minute, "duration limit of program working (in format '1d8h15m30s')")
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

// Init performs program data initialization.
func Init() {
	log.Println("starts")

	flag.Parse()

	log.Printf("execution duration limit is %s", pdur.String())

	// create context and wait the break
	exitctx, exitfn = context.WithTimeout(context.Background(), *pdur)
	go func() {
		// Make exit signal on function exit.
		defer exitfn()

		var sigint = make(chan os.Signal, 1)
		var sigterm = make(chan os.Signal, 1)
		// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C) or SIGTERM (Ctrl+/)
		// SIGKILL, SIGQUIT will not be caught.
		signal.Notify(sigint, syscall.SIGINT)
		signal.Notify(sigterm, syscall.SIGTERM)
		// Block until we receive our signal.
		select {
		case <-exitctx.Done():
			if errors.Is(exitctx.Err(), context.DeadlineExceeded) {
				log.Println("shutting down by timeout")
			} else if errors.Is(exitctx.Err(), context.Canceled) {
				log.Println("shutting down by cancel")
			} else {
				log.Printf("shutting down by %s", exitctx.Err().Error())
			}
		case <-sigint:
			log.Println("shutting down by break")
		case <-sigterm:
			log.Println("shutting down by process termination")
		}
		signal.Stop(sigint)
		signal.Stop(sigterm)
	}()

	var err error
	// get confiruration path
	if ConfigPath, err = DetectConfigPath(); err != nil {
		log.Fatal(err)
	}
	log.Printf("config path: %s\n", ConfigPath)

	// read main config file
	if err = ReadYaml(cfgfile, &Threads); err != nil {
		log.Fatal(err)
	}
	if len(Threads) == 0 {
		log.Fatal("no any chain of requests entries was readed")
	}
	for _, chain := range Threads {
		for _, ent := range chain.Entries {
			if ent.Method == "" {
				if ent.Data == "" {
					ent.Method = "GET"
				} else {
					ent.Method = "POST"
				}
			}
		}
		if chain.Repeats == 0 {
			chain.Repeats = -1
		}
	}
	log.Printf("readed file: '%s', threads: %d\n", cfgfile, len(Threads))
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
}

// Done performs graceful network shutdown,
// waits until all server threads will be stopped.
func Done() {
	var start = time.Now()
	// wait for exit signal
	<-exitctx.Done()
	// wait until all server threads will be stopped.
	exitwg.Wait()
	var rundur = time.Since(start)
	log.Printf("running time: %s, request number: %d\n", rundur.String(), ReqCount)
	log.Println("shutting down complete.")
}

// The End.
