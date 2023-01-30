package main

import (
	"context"
	"fmt"
	"log"

	authv3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.Dial("127.0.0.1:9000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf(err.Error())
	}
	defer func() { _ = conn.Close() }()
	grpcV3Client := authv3.NewAuthorizationClient(conn)

	request := &authv3.CheckRequest{
		Attributes: &authv3.AttributeContext{
			Request: &authv3.AttributeContext_Request{
				Http: &authv3.AttributeContext_HttpRequest{
					Host:    "127.0.0.1",
					Path:    "/check",
					Headers: map[string]string{"x-ext-authz": "deny"},
				},
			},
		},
	}

	response, err := grpcV3Client.Check(context.Background(), request)

	if err != nil {
		log.Fatal("Failed to get response from ext_auth service: %v", err)
	}

	// Allow reponse code is 0
	if response.Status.Code == int32(codes.OK) {
		fmt.Printf("the request is allowed")
	} else {
		fmt.Printf("the request is denied")
	}
}
