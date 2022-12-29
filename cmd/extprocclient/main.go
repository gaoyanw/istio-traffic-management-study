package main

import (
	"context"
	"flag"
	"log"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	pb "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var server = flag.String("server", "localhost:3443", "server address")

func main() {
	flag.Parse()

	opt := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.Dial(*server, opt)
	if err != nil {
		log.Fatalf("create client connection: %v", err)
	}
	defer conn.Close()

	c := pb.NewExternalProcessorClient(conn)
	ctx := context.Background()

	pc, err := c.Process(ctx)
	if err != nil {
		log.Fatalf("create streaming client: %v", err)
	}

	req := &pb.ProcessingRequest{
		AsyncMode: false,
		Request: &pb.ProcessingRequest_RequestHeaders{
			RequestHeaders: &pb.HttpHeaders{
				Headers: &corev3.HeaderMap{
					Headers: []*corev3.HeaderValue{
						{
							Key:   "authorization",
							Value: "bearer abc",
						},
					},
				},
			},
		},
	}

	if err := pc.Send(req); err != nil {
		log.Fatalf("send request: %v", err)
	}

	resp, err := pc.Recv()
	if err != nil {
		log.Fatalf("receive response: %v", err)
	}

	log.Printf("get resp: %q", proto.MarshalTextString(resp))
}
