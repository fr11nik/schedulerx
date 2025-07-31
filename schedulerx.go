package schedulerx

import (
	"context"
	"sync"
	"time"

	"golang.org/x/exp/slog"
)

// usecase operation with things that we will parse per time
// usecase операции будут являться объектом парсинга сервиса Parser

// parser does not use own append function he operates with manager append ops

type task struct {
	fn            func() error
	flushInterval time.Duration
}

type Parser struct {
	loggerCtx context.Context
	wg        *sync.WaitGroup
	quit      chan struct{}
	mu        sync.Mutex
	running   bool
	logger    *slog.Logger
	handlers  []task
}

func NewParser(opts ...Option) *Parser {
	p := &Parser{
		wg:       &sync.WaitGroup{},
		quit:     make(chan struct{}),
		handlers: []task{},
	}
	for _, opt := range opts {
		err := opt(p)
		if err != nil {
			return nil
		}
	}
	return p
}

func (j *Parser) Stop() {
	close(j.quit)
}

func (j *Parser) Wait() {
	j.wg.Wait()
}

func (j *Parser) Start() {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.running {
		return
	}
	j.running = true
	for _, h := range j.handlers {
		j.Run(h.fn, h.flushInterval)
	}
	j.Wait()
}

func (j *Parser) Register(fn func() error, flushInterval time.Duration) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.handlers = append(j.handlers, task{fn: fn, flushInterval: flushInterval})
}

func (j *Parser) Run(fn func() error, flushInterval time.Duration) {
	j.wg.Add(1)
	go func() {
		defer j.wg.Done()
		for {
			select {
			case <-time.After(flushInterval):
				err := fn()
				if err != nil {
					slog.ErrorCtx(j.loggerCtx, "Error: "+err.Error())
				}
			case <-j.quit: // Check if the quit signal is received
				return
			}
		}
	}()
}
