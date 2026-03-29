package tcpworker

import (
	"context"

	"github.com/sudesh856/suddpanzer/internal/worker"
)

type TCPPool struct {
	jobs    chan TCPJob
	results chan worker.Result
}

func NewPool(bufferSize int) *TCPPool {
	return &TCPPool{
		jobs:    make(chan TCPJob, bufferSize),
		results: make(chan worker.Result, bufferSize),
	}
}

func (p *TCPPool) Start(ctx context.Context, n int) {
	for i := 0; i < n; i++ {
		go RunTCPWorker(ctx, p.jobs, p.results)
	}
}

func (p *TCPPool) Submit(job TCPJob) {
	p.jobs <- job
}

func (p *TCPPool) Results() <-chan worker.Result {
	return p.results
}

func (p *TCPPool) Close() {
	close(p.jobs)
}