package tcpworker

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/sudesh856/suddpanzer/internal/worker"
)


type TCPJob struct {
	EndpointName   string
	Target         string        
	Payload        string        
	ReadBytes      int        
	ReadTimeout    time.Duration 
	ConnectTimeout time.Duration 
}



func RunTCPWorker(ctx context.Context, jobs <-chan TCPJob, results chan<- worker.Result) {
	conns := make(map[string]net.Conn)

	defer func() {
		for _, conn := range conns {
			conn.Close()
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-jobs:
			if !ok {
				return
			}
			results <- fireTCP(ctx, job, conns)
		}
	}
}

func fireTCP(ctx context.Context, job TCPJob, conns map[string]net.Conn) worker.Result {
	start := time.Now()

	conn, err := getConn(ctx, job, conns)
	if err != nil {
		return worker.Result{
			Latency:      time.Since(start),
			Err:          fmt.Errorf("tcp dial %q: %w", job.Target, err),
			EndpointName: job.EndpointName,
		}
	}


	_, err = conn.Write([]byte(job.Payload))
	if err != nil {
		delete(conns, job.Target)
		conn.Close()
		return worker.Result{
			Latency:      time.Since(start),
			Err:          fmt.Errorf("tcp write: %w", err),
			EndpointName: job.EndpointName,
		}
	}


	readTimeout := job.ReadTimeout
	if readTimeout == 0 {
		readTimeout = 5 * time.Second
	}
	conn.SetReadDeadline(time.Now().Add(readTimeout))


	var buf []byte
	if job.ReadBytes > 0 {

		buf = make([]byte, job.ReadBytes)
		_, err = io.ReadFull(conn, buf)
	} else {

		tmp := make([]byte, 4096)
		n, readErr := conn.Read(tmp)
		buf = tmp[:n]
		err = readErr
	}

	latency := time.Since(start)

	if err != nil && err != io.EOF {
		delete(conns, job.Target)
		conn.Close()
		return worker.Result{
			Latency:      latency,
			Err:          fmt.Errorf("tcp read: %w", err),
			EndpointName: job.EndpointName,
		}
	}

	return worker.Result{
		Latency:      latency,
		StatusCode:   200, 
		Bytes:        int64(len(buf)),
		Body:         buf,
		EndpointName: job.EndpointName,
	}
}

func getConn(ctx context.Context, job TCPJob, conns map[string]net.Conn) (net.Conn, error) {
	if conn, ok := conns[job.Target]; ok {
		return conn, nil
	}

	connectTimeout := job.ConnectTimeout
	if connectTimeout == 0 {
		connectTimeout = 10 * time.Second
	}

	dialer := net.Dialer{Timeout: connectTimeout}
	conn, err := dialer.DialContext(ctx, "tcp", job.Target)
	if err != nil {
		return nil, err
	}

	conns[job.Target] = conn
	return conn, nil
}