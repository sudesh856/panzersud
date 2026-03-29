package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
)

func init() {
	encoding.RegisterCodec(passthroughCodec{})
}

type passthroughCodec struct{}

func (passthroughCodec) Name() string { return "proto" }

func (passthroughCodec) Marshal(v interface{}) ([]byte, error) {
	if b, ok := v.([]byte); ok {
		return b, nil
	}
	if b, ok := v.(*[]byte); ok {
		return *b, nil
	}
	return nil, fmt.Errorf("cannot marshal %T", v)
}

func (passthroughCodec) Unmarshal(data []byte, v interface{}) error {
	if b, ok := v.(*[]byte); ok {
		*b = append((*b)[:0], data...)
		return nil
	}
	return fmt.Errorf("cannot unmarshal into %T", v)
}

// ✅ interface (REQUIRED)
type greeterService interface {
	SayHello(context.Context, []byte) ([]byte, error)
}

type greeterServer struct{}

func (g *greeterServer) SayHello(ctx context.Context, req []byte) ([]byte, error) {
	return []byte(`{"message":"hello from suddpanzer grpc server"}`), nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	srv := grpc.NewServer()

	srv.RegisterService(&grpc.ServiceDesc{
		ServiceName: "helloworld.Greeter",
		HandlerType: (*greeterService)(nil), // ✅ FIX
		Methods: []grpc.MethodDesc{
			{
				MethodName: "SayHello",
				Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
					var req []byte
					if err := dec(&req); err != nil {
						return nil, err
					}
					return srv.(greeterService).SayHello(ctx, req)
				},
			},
		},
		Streams:  []grpc.StreamDesc{},
		Metadata: "helloworld.proto",
	}, &greeterServer{})

	log.Println("gRPC test server listening on :50051")
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}