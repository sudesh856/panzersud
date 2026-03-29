package wsworker

import (
	"context"

	"github.com/sudesh856/suddpanzer/internal/worker"
)

type WSPool struct {
	jobs    chan WSJob
	results chan worker.Result
}

func NewPool(bufferSize int) *WSPool {
	return &WSPool{
		jobs:    make(chan WSJob, bufferSize),
		results: make(chan worker.Result, bufferSize),
	}
}

func (p *WSPool) Start(ctx context.Context, n int) {
	for i := 0; i < n; i++ {
		go RunWSWorker(ctx, p.jobs, p.results)
	}
}

func (p *WSPool) Submit(job WSJob) {
	p.jobs <- job
}

func (p *WSPool) Results() <-chan worker.Result {
	return p.results
}


func (p *WSPool) Close() {
	close(p.jobs)
}