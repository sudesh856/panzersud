
package grpcworker

import (
	"context"

	"github.com/sudesh856/suddpanzer/internal/worker"
)


type GRPCPool struct {
	jobs    chan GRPCJob
	results chan worker.Result
}

func NewPool(bufferSize int) *GRPCPool {
	return &GRPCPool{
		jobs:    make(chan GRPCJob, bufferSize),
		results: make(chan worker.Result, bufferSize),
	}
}


func (p *GRPCPool) Start(ctx context.Context, n int) {
	for i := 0; i < n; i++ {
		go RunGRPCWorker(ctx, p.jobs, p.results)
	}
}


func (p *GRPCPool) Submit(job GRPCJob) {
	p.jobs <- job
}


func (p *GRPCPool) Results() <-chan worker.Result {
	return p.results
}


func (p *GRPCPool) Close() {
	close(p.jobs)
}