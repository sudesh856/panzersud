

package grpcworker

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/sudesh856/suddpanzer/internal/worker"
)

func init() {
	

	encoding.RegisterCodec(passthroughCodec{})
}


type passthroughCodec struct{}

func (passthroughCodec) Name() string { return "proto" }

func (passthroughCodec) Marshal(v interface{}) ([]byte, error) {
	switch m := v.(type) {
	case []byte:
		return m, nil
	case *[]byte:
		return *m, nil
	default:
		return nil, fmt.Errorf("passthroughCodec: cannot marshal %T", v)
	}
}

func (passthroughCodec) Unmarshal(data []byte, v interface{}) error {
	switch m := v.(type) {
	case *[]byte:
		*m = append((*m)[:0], data...)
		return nil
	default:
		return fmt.Errorf("passthroughCodec: cannot unmarshal into %T", v)
	}
}

type GRPCJob struct {
	EndpointName string
	Target       string           
	Method       string           
	Payload      []byte            
	Headers      map[string]string 
	Insecure     bool              
	Timeout      time.Duration     
}


type connCache map[string]*grpc.ClientConn

func (c connCache) get(target string, insecureFlag bool) (*grpc.ClientConn, error) {
	if conn, ok := c[target]; ok {
		return conn, nil
	}

	var opts []grpc.DialOption
	if insecureFlag {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		

		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.NewClient(target, opts...)
	if err != nil {
		return nil, err
	}

	c[target] = conn
	return conn, nil
}

func (c connCache) closeAll() {
	for _, conn := range c {
		conn.Close()
	}
}



func RunGRPCWorker(ctx context.Context, jobs <-chan GRPCJob, results chan<- worker.Result) {
	cache := make(connCache)
	defer cache.closeAll()

	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-jobs:
			if !ok {
				return
			}
			results <- fire(ctx, job, cache)
		}
	}
}


func fire(ctx context.Context, job GRPCJob, cache connCache) worker.Result {
	start := time.Now()

	conn, err := cache.get(job.Target, job.Insecure)
	if err != nil {
		return worker.Result{
			Latency:      time.Since(start),
			Err:          fmt.Errorf("grpc dial %q: %w", job.Target, err),
			EndpointName: job.EndpointName,
		}
	}

	timeout := job.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	callCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if len(job.Headers) > 0 {
		md := metadata.New(job.Headers)
		callCtx = metadata.NewOutgoingContext(callCtx, md)
	}

	
	
	req := job.Payload
	var respBuf []byte

	path := job.Method
	if len(path) > 0 && path[0] != '/' {
		path = "/" + path
	}

	err = conn.Invoke(callCtx, path, req, &respBuf)
	latency := time.Since(start)

	if err != nil {
		st, _ := status.FromError(err)
		return worker.Result{
			Latency:      latency,
			Err:          fmt.Errorf("grpc %s: %s", st.Code(), st.Message()),
			StatusCode:   int(st.Code()),
			EndpointName: job.EndpointName,
		}
	}

	return worker.Result{
		Latency:      latency,
		StatusCode:   int(codes.OK), // 0
		Bytes:        int64(len(respBuf)),
		EndpointName: job.EndpointName,
	}
}