// internal/grpcworker/payload.go
//
// BuildPayload converts a JSON string into raw proto bytes using
// google.protobuf.Struct — the universal proto message type.
//
// This means you can call ANY gRPC service from a YAML file without
// compiling a .proto file. The server just needs to accept Struct,
// or you can use it with grpc reflection.
//
// For services that require a specific proto message (not Struct),
// you'd compile the .proto and replace this with proto.Marshal(yourMsg).

package grpcworker

import (
	"encoding/json"
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)



func BuildPayload(jsonPayload string) ([]byte, error) {
	if jsonPayload == "" {
		jsonPayload = "{}"
	}

	var m map[string]interface{}
	if err := json.Unmarshal([]byte(jsonPayload), &m); err != nil {
		return nil, fmt.Errorf("grpc payload: invalid JSON %q: %w", jsonPayload, err)
	}

	s, err := structpb.NewStruct(m)
	if err != nil {
		return nil, fmt.Errorf("grpc payload: structpb conversion: %w", err)
	}

	b, err := proto.Marshal(s)
	if err != nil {
		return nil, fmt.Errorf("grpc payload: proto marshal: %w", err)
	}

	return b, nil
}



func MustBuildPayload(jsonPayload string) []byte {
	b, err := BuildPayload(jsonPayload)
	if err != nil {
		panic(err)
	}
	return b
}