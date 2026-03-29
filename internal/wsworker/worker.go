
package wsworker

import (
	"context"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sudesh856/suddpanzer/internal/worker"
)


type WSJob struct {
	EndpointName   string
	URL            string        
	Payload        string       
	Headers        map[string]string
	ConnectTimeout time.Duration
	ReadTimeout    time.Duration 
}


func RunWSWorker(ctx context.Context, jobs <-chan WSJob, results chan<- worker.Result) {

	conns := make(map[string]*websocket.Conn)

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
			results <- fireWS(ctx, job, conns)
		}
	}
}

func fireWS(ctx context.Context, job WSJob, conns map[string]*websocket.Conn) worker.Result {
	start := time.Now()

	conn, err := getConn(ctx, job, conns)
	if err != nil {
		return worker.Result{
			Latency:      time.Since(start),
			Err:          fmt.Errorf("ws dial %q: %w", job.URL, err),
			EndpointName: job.EndpointName,
		}
	}

	
	if err := conn.WriteMessage(websocket.TextMessage, []byte(job.Payload)); err != nil {
		
		delete(conns, job.URL)
		conn.Close()
		return worker.Result{
			Latency:      time.Since(start),
			Err:          fmt.Errorf("ws write: %w", err),
			EndpointName: job.EndpointName,
		}
	}

	
	readTimeout := job.ReadTimeout
	if readTimeout == 0 {
		readTimeout = 5 * time.Second
	}
	conn.SetReadDeadline(time.Now().Add(readTimeout))

	
	_, msg, err := conn.ReadMessage()
	latency := time.Since(start)

	if err != nil {
		delete(conns, job.URL)
		conn.Close()
		return worker.Result{
			Latency:      latency,
			Err:          fmt.Errorf("ws read: %w", err),
			EndpointName: job.EndpointName,
		}
	}

	return worker.Result{
		Latency:      latency,
		StatusCode:   101, 
		Bytes:        int64(len(msg)),
		Body:         msg,
		EndpointName: job.EndpointName,
	}
}


func getConn(ctx context.Context, job WSJob, conns map[string]*websocket.Conn) (*websocket.Conn, error) {
	if conn, ok := conns[job.URL]; ok {
		return conn, nil
	}

	connectTimeout := job.ConnectTimeout
	if connectTimeout == 0 {
		connectTimeout = 10 * time.Second
	}

	dialCtx, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()

	dialer := websocket.Dialer{
		HandshakeTimeout: connectTimeout,
	}


	reqHeader := make(map[string][]string)
	for k, v := range job.Headers {
		reqHeader[k] = []string{v}
	}

	conn, _, err := dialer.DialContext(dialCtx, job.URL, reqHeader)
	if err != nil {
		return nil, err
	}

	conns[job.URL] = conn
	return conn, nil
}